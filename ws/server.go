package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/environ"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
	"github.com/gorilla/websocket"
)

type Request struct {
	ID   int32
	Name string
	Data []byte

	// server side
	remoteAddr net.Addr
}

func (r *Request) RemoteAddr() net.Addr {
	return r.remoteAddr
}

type serverConn struct {
	*Conn
	userID int64
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
	conn := &serverConn{Conn: NewConn(wconn)}
	conn.readTimeout = s.readTimeout
	logger.Debugf("New conn %s", wconn.RemoteAddr())
	if s.Handshake != nil {
		logger.Debugf("Start handshaking")
		if err = s.Handshake(conn); err != nil {
			logger.Errorf("Cannot handshake: %v", err)
			conn.Close()
			return
		}
		logger.Debugf("Finish handshaking")
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
	if req.ID == 0 {
		logger.Debugf("Received ping")
		return
	}
	startAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if conn.userID > 0 {
		ctx = wine.WithUserID(ctx, conn.userID)
	}
	//if deviceID, ok := conn.GetHeader("device_id").(string); ok && deviceID != "" {
	//	ctx = wine.WithDeviceID(ctx, deviceID)
	//}
	var reply *Reply
	var resultOrErr interface{}
	ctx = wine.WithRemoteAddr(ctx, conn.conn.RemoteAddr().String())
	result, err := s.Handle(ctx, req)
	if err != nil {
		resultOrErr = err
		if s := errors.GetCode(err); s > 0 {
			reply = NewErrorReply(req.ID, errors.Format(s, err.Error()))
		} else {
			reply = NewErrorReply(req.ID, errors.Format(http.StatusInternalServerError, err.Error()))
		}
	} else {
		resultOrErr = result
		data, err := json.Marshal(result)
		if err != nil {
			reply = NewErrorReply(req.ID, errors.Format(http.StatusInternalServerError, err.Error()))
		}
		reply = NewDataReply(req.ID, data)
	}
	if getUid, ok := result.(GetAuthUserID); ok {
		uid := getUid.GetAuthUserID()
		if conn.userID != uid {
			conn.userID = uid
			s.deleteUserConn(conn)
			s.saveUserConn(conn)
		}
	}
	if err = conn.Reply(reply); err != nil {
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

	var params interface{} = req.Data
	if m := r.JSONModel(); m != nil {
		pv := reflect.New(reflect.TypeOf(m))
		if err := json.Unmarshal(req.Data, pv.Interface()); err != nil {
			return nil, fmt.Errorf("cannot unmarshal json: %w", err)
		}
		params = pv.Elem().Interface()
	} else if msg := r.ProtobufModel(); msg != nil {
		pv := reflect.New(reflect.TypeOf(msg))
		if err := proto.Unmarshal(req.Data, pv.Interface().(proto.Message)); err != nil {
			return nil, fmt.Errorf("cannot unmarshal protobuf: %w", err)
		}
		params = pv.Elem().Interface()
	}

	h := (*handlerElem)(r.FirstHandler())
	if s.PreHandler != nil {
		return s.PreHandler.HandleRequest(withNextHandler(ctx, h), params)
	} else {
		return h.HandleRequest(ctx, params)
	}
}

func (s *Server) Push(ctx context.Context, userID int64, v interface{}) error {
	conns, ok := s.userConns.Load(userID)
	if !ok {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal: %w", err)
	}
	var firstErr error
	conns.(*sync.Map).Range(func(key, value interface{}) bool {
		conn := key.(*serverConn)
		err := conn.Push(data)
		if err != nil {
			logger.Errorf("Cannot push: %v", err)
			if firstErr != nil {
				firstErr = err
			}
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
	info := fmt.Sprintf("%s %d %s | %v",
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
		logger.Errorf("%s | %s | %v", info, req.Data, err)
	} else {
		logger.Infof(info)
	}
}
