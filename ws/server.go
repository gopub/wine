package ws

import (
	"net/http"

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

type Response struct {
	ID    int64         `json:"id"`
	Data  interface{}   `json:"data"`
	Error *errors.Error `json:"error"`
}

type Server struct {
	websocket.Upgrader
	*Router
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := new(Server)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := s.Upgrade(w, req, nil)
	if err != nil {
		wine.Error(err).Respond(req.Context(), w)
		return
	}
	_ = conn
	conn.Close()
}
