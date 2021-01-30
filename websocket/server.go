package websocket

import (
	"context"
	"fmt"
	"github.com/gopub/wine/ctxutil"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gopub/environ"
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
	"github.com/gorilla/websocket"
)

type Request struct {
	ID   int32
	Name string
	Data *Data

	// server side
	remoteAddr net.Addr
	Model      interface{}
}

func (r *Request) RemoteAddr() net.Addr {
	return r.remoteAddr
}

func (r *Request) bind(m interface{}) error {
	if m == nil {
		return nil
	}
	pv := reflect.New(reflect.TypeOf(m))
	if err := r.Data.Unmarshal(pv.Interface()); err != nil {
		return err
	}
	r.Model = pv.Elem().Interface()
	return wine.Validate(r.Model)
}

type serverConn struct {
	*Conn
	id       string // connections from the same user can share the same id
	userID   int64
	header   http.Header
	metadata map[string]string // metadata
}

func (c *serverConn) buildContext(ctx context.Context) context.Context {
	if c.userID > 0 {
		ctx = ctxutil.WithUserID(ctx, c.userID)
	}
	ctx = withServerConn(ctx, c)
	ctx = ctxutil.WithRequestHeader(ctx, c.header)
	return ctx
}

func (c *serverConn) GetID() string {
	return c.id
}

func (c *serverConn) GetHeader(key string) string {
	return c.header.Get(key)
}

func (c *serverConn) GetMetadata(key string) string {
	return c.metadata[key]
}

func (c *serverConn) GetValue(key string) string {
	v := c.GetMetadata(key)
	if v != "" {
		return v
	}
	return c.GetHeader(key)
}

// Server implements websocket server
type Server struct {
	websocket.Upgrader
	*Router
	readTimeout time.Duration
	timeout     time.Duration
	PreHandler  Handler
	conns       sync.Map // id:map[conn]bool
	Handshake   func(rw PacketReadWriter) error
	CallLogger  func(req *Request, resultOrErr interface{}, cost time.Duration)
	Recovery    bool
}

// Server implements http.Handler in order to take over http conn and upgrade to websocket conn
var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := &Server{
		Router:      NewRouter(),
		readTimeout: environ.Duration("wine.read_timeout", 20*time.Second),
		timeout:     environ.Duration("wine.timeout", 10*time.Second),
		CallLogger:  logCall,
		Recovery:    environ.Bool("wine.recovery", true),
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Recovery {
		defer func() {
			if e := recover(); e != nil {
				logger.Errorf("%v: %+v\n", r, e)
				logger.Errorf("\n%s\n", string(debug.Stack()))
			}
		}()
	}

	wconn, err := s.Upgrade(w, r, nil)
	if err != nil {
		wine.Error(err).Respond(r.Context(), w)
		return
	}
	conn := &serverConn{
		Conn:     NewConn(wconn),
		header:   r.Header,
		metadata: map[string]string{},
	}
	conn.readTimeout = s.readTimeout
	logger.Debugf("New conn %s", wconn.RemoteAddr())
	if s.Handshake != nil {
		logger.Debugf("Handshaking")
		if err = s.Handshake(conn); err != nil {
			logger.Errorf("Cannot handshake: %v", err)
			conn.Close()
			return
		}
		logger.Debugf("Handshake completed")
	}
	for {
		p, err := conn.Read()
		if err != nil {
			logger.Errorf("Cannot read: %v", err)
			break
		}
		switch v := p.V.(type) {
		case *Packet_Call:
			req := new(Request)
			req.ID = v.Call.Id
			req.Name = v.Call.Name
			req.Data = v.Call.Data
			req.remoteAddr = wconn.RemoteAddr()
			go s.HandleRequest(conn, req)
		case *Packet_Metadata:
			for k, val := range v.Metadata.Entries {
				conn.metadata[k] = val
			}
			logger.Debug("Metadata:", conn.metadata)
		case *Packet_Hello:
			go conn.Hello()
		default:
			break
		}
	}
	conn.Close()
	s.deleteConn(conn)
	if conn.userID != 0 {
		logger.Debugf("Close conn: %s, user=%d", wconn.RemoteAddr(), conn.userID)
	} else {
		logger.Debugf("Close conn: %s", wconn.RemoteAddr())
	}
}

func (s *Server) deleteConn(conn *serverConn) {
	conns, ok := s.conns.Load(conn.id)
	if !ok {
		return
	}
	m := conns.(*sync.Map)
	m.Delete(conn)

	// TODO: race condition with func storeConn
	empty := true
	m.Range(func(key, value interface{}) bool {
		empty = false
		return false
	})
	if empty {
		s.conns.Delete(conn.id)
	}
}

func (s *Server) storeConn(conn *serverConn) {
	if conn.id == "" {
		return
	}
	m, _ := s.conns.LoadOrStore(conn.id, &sync.Map{})
	m.(*sync.Map).Store(conn, true)
}

func (s *Server) HandleRequest(conn *serverConn, req *Request) {
	startAt := time.Now()
	if s.Recovery {
		defer func() {
			if e := recover(); e != nil {
				s.logCall(req, e, startAt)
				logger.Errorf("\n%s\n", string(debug.Stack()))
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	ctx = conn.buildContext(ctx)
	var resultOrErr interface{}
	result, err := s.Handle(ctx, req)
	if err != nil {
		resultOrErr = err
	} else {
		resultOrErr = result
		if getUid, ok := result.(GetAuthUserID); ok {
			conn.userID = getUid.GetAuthUserID()
		}
		if getConnID, ok := result.(GetConnID); ok {
			connID := getConnID.GetConnID()
			if conn.id != connID {
				// delete conn before assigning new id
				s.deleteConn(conn)
				conn.id = connID
				s.storeConn(conn)
			}
		}
	}

	if err = conn.Reply(req.ID, resultOrErr); err != nil {
		logger.Errorf("Cannot write reply: %v", err)
		conn.Close()
	}
	s.logCall(req, resultOrErr, startAt)
}

func (s *Server) Handle(ctx context.Context, req *Request) (interface{}, error) {
	r, _ := s.Match("", router.Normalize(req.Name))
	if r == nil {
		return nil, errors.NotFound("")
	}

	if err := req.bind(r.Model()); err != nil {
		return nil, fmt.Errorf("cannot bind model %T: %w", r.Model(), err)
	}

	h := (*handlerElem)(r.FirstHandler())
	if s.PreHandler != nil {
		return s.PreHandler.HandleRequest(withNextHandler(ctx, h), req.Model)
	} else {
		return h.HandleRequest(ctx, req.Model)
	}
}

func (s *Server) Push(ctx context.Context, connID string, typ int32, data interface{}) error {
	logger := log.FromContext(ctx).With("conn", connID, "type", typ, "data", data)
	conns, ok := s.conns.Load(connID)
	if !ok {
		return nil
	}
	d, err := MarshalData(data)
	if err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	var firstErr error
	conns.(*sync.Map).Range(func(key, value interface{}) bool {
		conn := key.(*serverConn)
		if err = conn.Push(typ, d); err != nil {
			logger.Errorf("Push: %v", err)
			if firstErr != nil {
				firstErr = err
			}
		} else {
			logger.Debugf("Push successfully")
		}

		if ctx.Err() != nil {
			if firstErr == nil {
				firstErr = ctx.Err()
			}
			return false
		}
		return true
	})
	return firstErr
}

func (s *Server) logCall(req *Request, resultOrErr interface{}, startAt time.Time) {
	if s.CallLogger == nil {
		return
	}
	s.CallLogger(req, resultOrErr, time.Since(startAt))
}

func logCall(req *Request, resultOrErr interface{}, cost time.Duration) {
	info := fmt.Sprintf("%s #%d %s %v",
		req.remoteAddr,
		req.ID,
		req.Name,
		cost)
	if getUid, ok := resultOrErr.(GetAuthUserID); ok {
		uid := getUid.GetAuthUserID()
		if uid > 0 {
			info = fmt.Sprintf("%s | user=%d", info, uid)
		}
	}
	if err, ok := resultOrErr.(error); ok {
		logger.Errorf("%s | %s | %v", info, req.Data.LogString(), err)
	} else if s, ok := resultOrErr.(wine.LogStringer); ok {
		logger.Debugf("%s | %s", info, s.LogString())
	} else {
		switch v := reflect.ValueOf(resultOrErr); v.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			if req.Model != nil {
				logger.Debugf("%s | %s | size=%d", info, req.Data.LogString(), v.Len())
			} else {
				logger.Debugf("%s | size=%d", info, v.Len())
			}
		default:
			logger.Debug(info)
		}
	}
}
