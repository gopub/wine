package ws

import (
	"context"
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

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

type request struct {
	ID   int64       `json:"id"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type response struct {
	ID    int64         `json:"id"`
	Data  interface{}   `json:"data"`
	Error *errors.Error `json:"error"`
}

type GetAuthUserID interface {
	GetAuthUserID() int64
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
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := new(Server)
	s.Router = NewRouter()
	s.timeout = 10 * time.Second
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
		req := new(request)
		if err := conn.ReadJSON(req); err != nil {
			logger.Errorf("ReadJSON: %v", err)
			break
		}
		go s.HandleRequest(conn, req)
	}
	conn.Close()
	s.connUserIDs.Delete(conn)
}

func (s *Server) HandleRequest(conn *websocket.Conn, req *request) {
	if req.ID == 0 {
		// Heartbreak request
		return
	}
	resp := &response{
		ID: req.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	uid, _ := s.connUserIDs.Load(conn)
	if n, _ := uid.(int64); n > 0 {
		ctx = wine.WithUserID(ctx, n)
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
	if u, ok := result.(GetAuthUserID); ok {
		s.connUserIDs.Store(conn, u.GetAuthUserID())
	}
	if err = conn.WriteJSON(resp); err != nil {
		logger.Errorf("WriteJSON: %v", err)
		conn.Close()
	}
}

func (s *Server) Handle(ctx context.Context, req *request) (interface{}, error) {
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
