package ws

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopub/conv"
	"github.com/gopub/environ"
	"github.com/gopub/errors"
	"github.com/gopub/types"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
	"github.com/gorilla/websocket"
)

type serverConn struct {
	*websocket.Conn
	userID int64

	mu     sync.RWMutex
	header types.M

	id int64
}

func newServerConn(conn *websocket.Conn) *serverConn {
	return &serverConn{
		Conn:   conn,
		header: types.M{},
	}
}

func (c *serverConn) SetHeader(k string, v interface{}) {
	c.mu.Lock()
	if v == nil {
		delete(c.header, k)
	} else {
		c.header[k] = v
	}
	c.mu.Unlock()
}

func (c *serverConn) GetHeader(k string) interface{} {
	c.mu.RLock()
	v := c.header[k]
	c.mu.RUnlock()
	return v
}

func (c *serverConn) WriteJSON(i interface{}) error {
	c.mu.Lock()
	err := c.Conn.WriteJSON(i)
	c.mu.Unlock()
	return err
}

func (c *serverConn) NextID() int64 {
	atomic.AddInt64(&c.id, 2)
	return c.id
}

type Server struct {
	websocket.Upgrader
	*Router
	readTimeout      time.Duration
	timeout          time.Duration
	PreHandler       Handler
	uidToConns       sync.Map // uid:map[conn]bool
	HandshakeHandler func(rw ReadWriter) error
	ResultLogger     func(req *Request, resp *Response, cost time.Duration)
	Recovery         bool
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := &Server{
		Router:       NewRouter(),
		readTimeout:  environ.Duration("wine.read_timeout", 20*time.Second),
		timeout:      environ.Duration("wine.timeout", 10*time.Second),
		ResultLogger: logResult,
		Recovery:     environ.Bool("wine.recovery", true),
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
	conn := newServerConn(wconn)
	logger.Debugf("New conn %s", wconn.RemoteAddr())
	if s.HandshakeHandler != nil {
		logger.Debugf("Start handshaking")
		if err = s.HandshakeHandler(conn); err != nil {
			logger.Errorf("Cannot handshake: %v", err)
			conn.Close()
			return
		}
		logger.Debugf("Finish handshaking")
	}
	for {
		if err = conn.SetReadDeadline(time.Now().Add(s.readTimeout)); err != nil {
			logger.Errorf("Cannot set read deadline: %v", err)
			break
		}
		req := new(Request)
		if err := conn.ReadJSON(req); err != nil {
			logger.Errorf("Cannot read: %v", err)
			break
		}
		req.remoteAddr = conn.RemoteAddr()
		for k, v := range req.Header {
			conn.SetHeader(k, v)
		}
		go s.HandleRequest(conn, req)
	}
	conn.Close()
	s.deleteUserConn(conn)
	if conn.userID != 0 {
		logger.Debugf("Close conn: %s, user=%d", conn.RemoteAddr().String(), conn.userID)
	} else {
		logger.Debugf("Close conn: %s", conn.RemoteAddr().String())
	}
}

func (s *Server) deleteUserConn(conn *serverConn) {
	if conn.userID == 0 {
		return
	}
	m, ok := s.uidToConns.Load(conn.userID)
	if !ok {
		return
	}
	m.(*sync.Map).Delete(conn)
}

func (s *Server) saveUserConn(conn *serverConn) {
	if conn.userID == 0 {
		return
	}
	m, _ := s.uidToConns.LoadOrStore(conn.userID, &sync.Map{})
	m.(*sync.Map).Store(conn, true)
}

func (s *Server) HandleRequest(conn *serverConn, req *Request) {
	if req.ID == 0 {
		logger.Debugf("Received ping")
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
	if conn.userID > 0 {
		ctx = wine.WithUserID(ctx, conn.userID)
	}
	if deviceID, ok := conn.GetHeader("device_id").(string); ok && deviceID != "" {
		ctx = wine.WithDeviceID(ctx, deviceID)
	}
	ctx = wine.WithRemoteAddr(ctx, conn.RemoteAddr().String())
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
		if conn.userID != uid {
			conn.userID = uid
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
		conn := key.(*serverConn)
		err := conn.WriteJSON(&Response{
			ID:   conn.NextID(),
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
		logger.Errorf("%s | %s | %v", info, wine.JSONString(req.Body), resp.Error)
	} else {
		logger.Infof(info)
	}
}
