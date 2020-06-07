package ws

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
	"github.com/gorilla/websocket"
)

type serverConn struct {
	*websocket.Conn
	UserID int64
}

type Server struct {
	websocket.Upgrader
	*Router
	timeout          time.Duration
	PreHandler       Handler
	uidToConns       sync.Map // uid:map[conn]bool
	HandshakeHandler func(rw ReadWriter) error
	ResultLogger     func(req *Request, resp *Response, cost time.Duration)
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := &Server{
		Router:       NewRouter(),
		timeout:      10 * time.Second,
		ResultLogger: logResult,
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			logger.Errorf("%v: %+v\n", r, e)
			logger.Errorf("\n%s\n", string(debug.Stack()))
		}
	}()

	conn := new(serverConn)
	var err error
	conn.Conn, err = s.Upgrade(w, r, nil)
	if err != nil {
		wine.Error(err).Respond(r.Context(), w)
		return
	}

	if s.HandshakeHandler != nil {
		if err = s.HandshakeHandler(conn); err != nil {
			logger.Errorf("Cannot handshake: %v", err)
			conn.Close()
			return
		}
	}
	for {
		if err = conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
			logger.Errorf("SetReadDeadline: %v", err)
			break
		}
		req := new(Request)
		if err := conn.ReadJSON(req); err != nil {
			logger.Errorf("ReadJSON: %v", err)
			break
		}
		req.remoteAddr = conn.RemoteAddr()
		go s.HandleRequest(conn, req)
	}
	conn.Close()
	s.deleteUserConn(conn)
}

func (s *Server) deleteUserConn(conn *serverConn) {
	if conn.UserID == 0 {
		return
	}
	m, ok := s.uidToConns.Load(conn.UserID)
	if !ok {
		return
	}
	m.(*sync.Map).Delete(conn)
}

func (s *Server) saveUserConn(conn *serverConn) {
	if conn.UserID == 0 {
		return
	}
	m, _ := s.uidToConns.LoadOrStore(conn.UserID, &sync.Map{})
	m.(*sync.Map).Store(conn, true)
}

func (s *Server) HandleRequest(conn *serverConn, req *Request) {
	if req.ID == 0 {
		// Heartbreak Request
		return
	}
	resp := &Response{
		ID: req.ID,
	}
	if s.ResultLogger != nil {
		defer s.logResult(req, resp, time.Now())
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if conn.UserID > 0 {
		ctx = wine.WithUserID(ctx, conn.UserID)
	}
	result, err := s.Handle(ctx, req)
	if err != nil {
		if s := errors.GetCode(err); s > 0 {
			resp.Error = errors.Format(s, err.Error())
		} else {
			resp.Error = errors.Format(http.StatusInternalServerError, err.Error())
		}
	} else {
		resp.Data = result
	}
	if getUid, ok := result.(GetAuthUserID); ok {
		uid := getUid.GetAuthUserID()
		if conn.UserID != uid {
			conn.UserID = uid
			s.deleteUserConn(conn)
			s.saveUserConn(conn)
		}
	}
	if err = conn.WriteJSON(resp); err != nil {
		logger.Errorf("WriteJSON: %v", err)
		conn.Close()
	}
}

func (s *Server) Handle(ctx context.Context, req *Request) (interface{}, error) {
	r, _ := s.Match("", router.Normalize(req.Name))
	if r == nil {
		return nil, errors.NotFound("")
	}

	data := req.Body
	if m := r.Model(); m != nil {
		pv := reflect.New(reflect.TypeOf(m))
		if err := conv.Assign(pv.Interface(), req.Body); err != nil {
			return nil, err
		}
		data = pv.Elem().Interface()
	}

	h := (*handlerElem)(r.FirstHandler())
	if s.PreHandler != nil {
		return s.PreHandler.HandleRequest(withNextHandler(ctx, h), data)
	} else {
		return h.HandleRequest(ctx, data)
	}
}

func (s *Server) Push(ctx context.Context, userID int64, data interface{}) error {
	conns, ok := s.uidToConns.Load(userID)
	if !ok {
		return nil
	}
	var firstErr error
	conns.(*sync.Map).Range(func(key, value interface{}) bool {
		err := key.(*serverConn).WriteJSON(&Response{
			Data: data,
		})
		if err != nil {
			logger.Errorf("WriteJSON: %v", err)
			if firstErr != nil {
				firstErr = err
			}
		}
		return true
	})
	return firstErr
}

func (s *Server) logResult(req *Request, resp *Response, startAt time.Time) {
	if s.ResultLogger == nil {
		return
	}
	s.ResultLogger(req, resp, time.Since(startAt))
}

func logResult(req *Request, resp *Response, cost time.Duration) {
	info := fmt.Sprintf("%s %d %s | %v",
		req.remoteAddr,
		req.ID,
		req.Name,
		cost)
	if getUid, ok := resp.Data.(GetAuthUserID); ok {
		uid := getUid.GetAuthUserID()
		if uid > 0 {
			info = fmt.Sprintf("%s | user=%d", info, uid)
		}
	}
	if resp.Error != nil {
		logger.Errorf("%s | %v", info, resp.Error)
	} else {
		logger.Infof(info)
	}
}
