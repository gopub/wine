package ws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gopub/conv"
	"github.com/gopub/environ"
	"github.com/gopub/errors"
	"github.com/gopub/types"
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
	userID int64
	header map[string]string
}

func (c *serverConn) BuildContext(ctx context.Context) context.Context {
	if c.userID > 0 {
		ctx = wine.WithUserID(ctx, c.userID)
	}
	if deviceID := c.header["device_id"]; deviceID != "" {
		ctx = wine.WithDeviceID(ctx, deviceID)
	}
	if loc, _ := types.NewPointFromString(c.header["coordinate"]); loc != nil {
		ctx = wine.WithCoordinate(ctx, loc)
	}
	return ctx
}

type Server struct {
	websocket.Upgrader
	*Router
	readTimeout time.Duration
	timeout     time.Duration
	PreHandler  Handler
	userConns   sync.Map // uid:map[conn]bool
	Handshake   func(rw PacketReadWriter) error
	CallLogger  func(req *Request, resultOrErr interface{}, cost time.Duration)
	Recovery    bool
}

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
		Conn:   NewConn(wconn),
		header: map[string]string{},
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
		case *Packet_Header:
			for k, val := range v.Header.Entries {
				conn.header[k] = val
			}
		case *Packet_Hello:
			go conn.Hello()
		default:
			break
		}
	}
	conn.Close()
	s.deleteUserConn(conn)
	if conn.userID != 0 {
		logger.Debugf("Close conn: %s, user=%d", wconn.RemoteAddr(), conn.userID)
	} else {
		logger.Debugf("Close conn: %s", wconn.RemoteAddr())
	}
}

func (s *Server) deleteUserConn(conn *serverConn) {
	if conn.userID == 0 {
		return
	}
	m, ok := s.userConns.Load(conn.userID)
	if !ok {
		return
	}
	m.(*sync.Map).Delete(conn)
}

func (s *Server) saveUserConn(conn *serverConn) {
	if conn.userID == 0 {
		return
	}
	m, _ := s.userConns.LoadOrStore(conn.userID, &sync.Map{})
	m.(*sync.Map).Store(conn, true)
}

func (s *Server) HandleRequest(conn *serverConn, req *Request) {
	startAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	ctx = conn.BuildContext(ctx)
	var resultOrErr interface{}
	ctx = wine.WithRemoteAddr(ctx, conn.conn.RemoteAddr().String())
	ctx = withPusher(ctx, s)
	result, err := s.Handle(ctx, req)
	if err != nil {
		resultOrErr = err
	} else {
		resultOrErr = result
	}
	if getUid, ok := resultOrErr.(GetAuthUserID); ok {
		uid := getUid.GetAuthUserID()
		if conn.userID != uid {
			conn.userID = uid
			s.deleteUserConn(conn)
			s.saveUserConn(conn)
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

func (s *Server) Push(ctx context.Context, userID int64, v interface{}) error {
	conns, ok := s.userConns.Load(userID)
	if !ok {
		return nil
	}
	data, err := MarshalData(v)
	if err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	var firstErr error
	conns.(*sync.Map).Range(func(key, value interface{}) bool {
		conn := key.(*serverConn)
		if err = conn.WriteData(data); err != nil {
			logger.Errorf("Write data: user=%d, %v", userID, err)
			if firstErr != nil {
				firstErr = err
			}
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
		logger.Errorf("%s | %s | %v", info, conv.MustJSONString(req.Model), err)
	} else {
		if s, ok := resultOrErr.(wine.LogStringer); ok {
			logger.Infof("%s | %s", info, s.LogString())
		} else {
			switch v := reflect.ValueOf(resultOrErr); v.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				if req.Model != nil {
					logger.Infof("%s | %s | size=%d", info, conv.MustJSONString(req.Model), v.Len())
				} else {
					logger.Infof("%s | size=%d", info, v.Len())
				}
			default:
				logger.Info(info)
			}
		}
	}
}
