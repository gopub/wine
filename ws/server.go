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
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
	"github.com/gorilla/websocket"
)

type ReadWriter interface {
	ReadJSON(i interface{}) error
	WriteJSON(i interface{}) error
}

var logger = wine.Logger()

func SetLogger(l *log.Logger) {
	logger = l
}

type Request struct {
	ID   int64       `json:"id,omitempty"`
	Name string      `json:"name,omitempty"`
	Data interface{} `json:"data,omitempty"`

	uid        int64
	remoteAddr net.Addr
}

func (r *Request) UserID() int64 {
	return r.uid
}

func (r *Request) SetUserID(id int64) {
	r.uid = id
}

func (r *Request) RemoteAddr() net.Addr {
	return r.remoteAddr
}

type Response struct {
	ID    int64         `json:"id,omitempty"`
	Data  interface{}   `json:"data,omitempty"`
	Error *errors.Error `json:"error,omitempty"`
}

type contextKey int

// Context keys
const (
	ckNextHandler contextKey = iota + 1
)

func Next(ctx context.Context, req interface{}) (interface{}, error) {
	i, _ := ctx.Value(ckNextHandler).(Handler)
	if i == nil {
		return nil, errors.NotImplemented("")
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ckNextHandler, h)
}

type Server struct {
	websocket.Upgrader
	*Router
	timeout          time.Duration
	PreHandler       Handler
	connUserIDs      sync.Map
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

	conn, err := s.Upgrade(w, r, nil)
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
	s.connUserIDs.Delete(conn)
}

func (s *Server) HandleRequest(conn *websocket.Conn, req *Request) {
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
	if id, ok := s.connUserIDs.Load(conn); ok {
		req.uid, _ = id.(int64)
		if req.uid > 0 {
			ctx = wine.WithUserID(ctx, req.uid)
		}
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
	s.connUserIDs.Store(conn, req.uid)
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

	data := req.Data
	if m := r.Model(); m != nil {
		pv := reflect.New(reflect.TypeOf(m))
		if err := conv.Assign(pv.Interface(), req.Data); err != nil {
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
	if req.uid > 0 {
		info = fmt.Sprintf("%s | user=%d", info, req.uid)
	}
	if resp.Error != nil {
		logger.Errorf("%s | %v", info, resp.Error)
	} else {
		logger.Infof(info)
	}
}
