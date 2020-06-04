package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gopub/conv"

	"github.com/gopub/wine/router"

	"github.com/gopub/types"
	"github.com/gopub/wine"
	"github.com/gopub/wine/errors"
	"github.com/gorilla/websocket"
)

type Request struct {
	ID     int64   `json:"id"`
	Path   string  `json:"path"`
	Params types.M `json:"params"`
}

func (r *Request) BindModel(m interface{}) error {
	// Unsafe assignment, so ignore error
	data, err := json.Marshal(r.Params)
	if err != nil {
		return err
	}
	_ = json.Unmarshal(data, m)
	// As all values in query will be parsed into string type
	// conv.Assign can convert string to int automatically
	_ = conv.Assign(m, r.Params)
	return nil
}

type Response struct {
	ID    int64         `json:"id"`
	Data  interface{}   `json:"data"`
	Error *errors.Error `json:"error"`
}

type Server struct {
	websocket.Upgrader
	*Router
	timeout    time.Duration
	PreHandler Handler
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := new(Server)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrade(w, r, nil)
	if err != nil {
		wine.Error(err).Respond(r.Context(), w)
		return
	}
	for {
		req := new(Request)
		err := conn.ReadJSON(req)
		if err != nil {
			logger.Errorf("ReadJSON: %v", err)
			break
		}
	}
	conn.Close()
}

func (s *Server) HandleRequest(conn *websocket.Conn, req *Request) {
	resp := &Response{
		ID: req.ID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
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
	if err = conn.WriteJSON(result); err != nil {
		logger.Errorf("WriteJSON: %v", err)
		conn.Close()
	}
}

func (s *Server) Handle(ctx context.Context, req *Request) (interface{}, error) {
	path := router.Normalize(req.Path)
	r, params := s.Match("", req.Path)
	if r == nil {
		return nil, errors.NotFound("")
	}
	for k, v := range params {
		req.Params[k] = v
	}

	var payload interface{} = req.Params
	if m := r.Model(); m != nil {
		err := req.BindModel(&m)
		if err != nil {
			return nil, err
		}
		if v, ok := m.(interface{ Validate() error }); ok {
			if err = v.Validate(); err != nil {
				return nil, err
			}
		}
		payload = m
	}

	h := (*handlerElem)(r.FirstHandler())
	if s.PreHandler != nil && !reservedPaths[path] {
		return s.PreHandler.HandleRequest(withNextHandler(ctx, h), payload)
	} else {
		return h.HandleRequest(ctx, payload)
	}
}
