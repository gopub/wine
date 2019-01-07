package wine

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

// http.Flusher doesn't return error, however gzip.Writer/deflate.Writer only implement `Flush() error`
type flusher interface {
	Flush() error
}

type compressedResponseWriter struct {
	http.ResponseWriter
	compressedWriter io.Writer
}

func newCompressedResponseWriter(w http.ResponseWriter, encoding string) (*compressedResponseWriter, error) {
	switch encoding {
	case "gzip":
		cw := &compressedResponseWriter{}
		cw.ResponseWriter = w
		cw.compressedWriter = gzip.NewWriter(w)
		return cw, nil
	case "deflate":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		cw := &compressedResponseWriter{}
		cw.compressedWriter = fw
		cw.ResponseWriter = w
		return cw, nil
	default:
		return nil, errors.New("Unsupported encoding")
	}
}

func (w *compressedResponseWriter) Write(data []byte) (int, error) {
	return w.compressedWriter.Write(data)
}

func (w *compressedResponseWriter) Flush() {
	// Flush the compressed writer, then flush httpResponseWriter
	if f, ok := w.compressedWriter.(flusher); ok {
		f.Flush()
		if ff, ok := w.ResponseWriter.(http.Flusher); ok {
			ff.Flush()
		}
	}
}

func (w *compressedResponseWriter) Close() error {
	if closer, ok := w.compressedWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func compressionHandler(h http.Handler) http.Handler {
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
