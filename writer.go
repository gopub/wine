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

var _ statusGetter = (*ResponseWriter)(nil)
var _ http.Hijacker = (*ResponseWriter)(nil)

// ResponseWriter is a wrapper of http.ResponseWriter to make sure write status code only one time
type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
	}
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	if w.status > 0 {
		logger.Warnf("Failed to overwrite status code")
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *ResponseWriter) Status() int {
	return w.status
}

func (w *ResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

type CompressedResponseWriter struct {
	http.ResponseWriter
	compressedWriter io.Writer
}

func NewCompressedResponseWriter(w http.ResponseWriter, encoding string) (*CompressedResponseWriter, error) {
	switch encoding {
	case "gzip":
		cw := &CompressedResponseWriter{}
		cw.ResponseWriter = w
		cw.compressedWriter = gzip.NewWriter(w)
		return cw, nil
	case "deflate":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		cw := &CompressedResponseWriter{}
		cw.compressedWriter = fw
		cw.ResponseWriter = w
		return cw, nil
	default:
		return nil, errors.New("unsupported encoding")
	}
}

func (w *CompressedResponseWriter) Write(data []byte) (int, error) {
	return w.compressedWriter.Write(data)
}

func (w *CompressedResponseWriter) Flush() {
	// Flush the compressed writer, then flush httpResponseWriter
	if f, ok := w.compressedWriter.(flusher); ok {
		if err := f.Flush(); err != nil {
			logger.Errorf("cannot flush: %v", err)
		}
		if ff, ok := w.ResponseWriter.(http.Flusher); ok {
			ff.Flush()
		}
	}
}

func (w *CompressedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

func (w *CompressedResponseWriter) Close() error {
	if closer, ok := w.compressedWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func CompressionHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := r.Header.Get("Accept-Encoding")
		if strings.Contains(enc, "gzip") {
			cw, err := NewCompressedResponseWriter(w, "gzip")
			if err != nil {
				panic(err)
			}
			w.Header().Set("Content-Encoding", "gzip")
			h.ServeHTTP(cw, r)
			return
		}

		if strings.Contains(enc, "deflate") {
			cw, err := NewCompressedResponseWriter(w, "deflate")
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
