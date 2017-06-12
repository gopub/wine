package wine

import (
	"net/http"
	"strings"
)

import (
	"compress/flate"
	"compress/gzip"
	"io"
)

type compressionResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *compressionResponseWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

func compressionWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := r.Header.Get("Accept-Encoding")
		if strings.Contains(enc, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			cw := &compressionResponseWriter{}
			cw.ResponseWriter = w
			cw.Writer = gzip.NewWriter(w)
			h.ServeHTTP(cw, r)
			return
		}

		if strings.Contains(enc, "deflate") {
			fw, err := flate.NewWriter(w, flate.DefaultCompression)
			if err == nil {
				w.Header().Set("Content-Encoding", "deflate")
				cw := &compressionResponseWriter{}
				cw.Writer = fw
				cw.ResponseWriter = w
				h.ServeHTTP(cw, r)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
