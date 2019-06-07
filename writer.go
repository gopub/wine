package wine

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

type statusGetter interface {
	Status() int
}

// http.Flusher doesn't return error, however gzip.Writer/deflate.Writer only implement `Flush() error`
type flusher interface {
	Flush() error
}

var _ statusGetter = (*responseWriterWrapper)(nil)
var _ http.Hijacker = (*responseWriterWrapper)(nil)

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	if w.status > 0 {
		logger.Warnf("Failed to overwrite status code")
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterWrapper) Status() int {
	return w.status
}

func (w *responseWriterWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
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

func (w *compressedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
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
