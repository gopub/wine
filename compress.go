package wine

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

type compressedResponseWriter struct {
	http.ResponseWriter
	compressedWriter io.Writer
}

func newCompressedResponseWriter(w http.ResponseWriter, compressionName string) (*compressedResponseWriter, error) {
	switch compressionName {
	case "gzip":
		cw := &compressedResponseWriter{}
		cw.ResponseWriter = w
		cw.compressedWriter = gzip.NewWriter(w)
		return cw, nil
	case "defalte":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		cw := &compressedResponseWriter{}
		cw.compressedWriter = fw
		cw.ResponseWriter = w
		return cw, nil
	default:
		return nil, errors.New("Unsupported compressionName")
	}
}

func (g *compressedResponseWriter) Write(data []byte) (int, error) {
	return g.compressedWriter.Write(data)
}

func (g *compressedResponseWriter) Close() error {
	if closer, ok := g.compressedWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func compressionWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := r.Header.Get("Accept-Encoding")
		if strings.Contains(enc, "gzip") {
			cw, err := newCompressedResponseWriter(w, "gzip")
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Encoding", "gzip")
			h.ServeHTTP(cw, r)
			return
		}

		if strings.Contains(enc, "deflate") {
			cw, err := newCompressedResponseWriter(w, "deflate")
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Encoding", "deflate")
			h.ServeHTTP(cw, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}
