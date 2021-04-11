package wine

import (
	"embed"
	"fmt"
	"net/http"
)

func EmbedFileServer(fs embed.FS) *Server {
	s := NewServer(nil)
	s.Handle("/*", HTTPHandlerFunc(http.FileServer(http.FS(fs))))
	ip := "127.0.0.1"
	port := SelectLocalPort(ip, 10000, 20000)
	addr := fmt.Sprintf("%s:%d", ip, port)
	go s.Run(addr)
	return s
}
