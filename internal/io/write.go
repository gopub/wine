package io

import (
	"bufio"
	"errors"
	"net"
	"net/http"

	"github.com/gopub/wine/httpvalue"
)

type statusGetter interface {
	Status() int
}

// http.Flusher doesn't return error, however gzip.Writer/deflate.Writer only implement `Flush() error`
type flusher interface {
	Flush() error
}

var (
	_ statusGetter  = (*ResponseWriter)(nil)
	_ http.Hijacker = (*ResponseWriter)(nil)
	_ http.Flusher  = (*ResponseWriter)(nil)
)

// ResponseWriter is a wrapper of http.ResponseWriter to make sure write status code only one time
type ResponseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
	}
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	if !httpvalue.IsValidStatus(statusCode) {
		logger.Errorf("Cannot write invalid status code: %d", statusCode)
		statusCode = http.StatusInternalServerError
	}
	if w.status > 0 {
		logger.Warnf("Status code already written")
		return
	}
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body = data
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

func (w *ResponseWriter) Body() []byte {
	return w.body
}
