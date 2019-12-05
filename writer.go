package wine

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"context"
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

var _ statusGetter = (*responseWriter)(nil)
var _ http.Hijacker = (*responseWriter)(nil)

// responseWriter is a wrapper of http.responseWriter to make sure write status code only one time
type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: rw,
	}
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.status > 0 {
		logger.Warnf("Failed to overwrite status code")
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("hijack not supported")
}

type compressedResponseWriter struct {
	http.ResponseWriter
	compressedWriter io.Writer
	err              error
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
		return nil, errors.New("unsupported encoding")
	}
}

func (w *compressedResponseWriter) Write(data []byte) (int, error) {
	return w.compressedWriter.Write(data)
}

func (w *compressedResponseWriter) Flush() {
	// Flush the compressed writer, then flush httpResponseWriter
	if f, ok := w.compressedWriter.(flusher); ok {
		if err := f.Flush(); err != nil {
			logger.Errorf("Flush: %v", err)
			w.err = err
		}
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

func (w *compressedResponseWriter) Error() error {
	return w.err
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

// GetResponseWriter returns response writer from the context
func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(CKHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func withResponseWriter(ctx context.Context, rw http.ResponseWriter) context.Context {
	return context.WithValue(ctx, CKHTTPResponseWriter, rw)
}
