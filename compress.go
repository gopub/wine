package wine

import (
	"net/http"
	"strings"
)

import (
	"compress/gzip"
	"io"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func newGzipResponseWriter(rw http.ResponseWriter) *gzipResponseWriter {
	w := &gzipResponseWriter{}
	w.Writer = gzip.NewWriter(rw)
	w.ResponseWriter = rw
	return w
}

func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

func compressionWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		gw := newGzipResponseWriter(w)
		gw.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gw, r)
	})
}
