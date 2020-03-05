package wine

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/log"
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
var _ http.Flusher = (*responseWriterWrapper)(nil)
var _ http.Hijacker = (*compressedResponseWriter)(nil)
var _ http.Flusher = (*compressedResponseWriter)(nil)

// responseWriterWrapper is a wrapper of http.responseWriterWrapper to make sure write status code only one time
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

func wrapResponseWriter(rw http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: rw,
	}
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	if w.status > 0 {
		logger.Warnf("Status code already written")
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
	compressWriter io.Writer
	err            error
}

func newCompressedResponseWriter(w http.ResponseWriter, encoding string) (*compressedResponseWriter, error) {
	switch encoding {
	case "gzip":
		cw := &compressedResponseWriter{}
		cw.ResponseWriter = w
		cw.compressWriter = gzip.NewWriter(w)
		w.Header().Set("Content-Encoding", encoding)
		return cw, nil
	case "deflate":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("new flate writer: %w", err)
		}
		cw := &compressedResponseWriter{}
		cw.compressWriter = fw
		cw.ResponseWriter = w
		w.Header().Set("Content-Encoding", encoding)
		return cw, nil
	default:
		return nil, errors.New("unsupported encoding")
	}
}

func (w *compressedResponseWriter) Write(data []byte) (int, error) {
	return w.compressWriter.Write(data)
}

func (w *compressedResponseWriter) Flush() {
	// Flush the compressed writer, then flush httpResponseWriter
	if f, ok := w.compressWriter.(flusher); ok {
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
	if closer, ok := w.compressWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// GetResponseWriter returns response writer from the context
func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(CKHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func withResponseWriter(ctx context.Context, rw http.ResponseWriter) context.Context {
	return context.WithValue(ctx, CKHTTPResponseWriter, rw)
}

func compressWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	// Add compression to responseWriterWrapper
	if enc := req.Header.Get("Accept-Encoding"); gox.IndexOfString(acceptEncodings, enc) >= 0 {
		cw, err := newCompressedResponseWriter(rw, enc)
		if err != nil {
			log.Errorf("Create compressed response writer: %v", err)
			return rw
		}
		return cw
	}
	return rw
}
