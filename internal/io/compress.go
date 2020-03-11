package io

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var _ statusGetter = (*CompressResponseWriter)(nil)
var _ http.Hijacker = (*CompressResponseWriter)(nil)
var _ http.Flusher = (*CompressResponseWriter)(nil)

type CompressResponseWriter struct {
	*ResponseWriter
	compressWriter io.Writer
	err            error
}

func NewCompressResponseWriter(w *ResponseWriter, encoding string) (*CompressResponseWriter, error) {
	switch encoding {
	case "gzip":
		cw := &CompressResponseWriter{}
		cw.ResponseWriter = w
		cw.compressWriter = gzip.NewWriter(w)
		w.Header().Set("Content-Encoding", encoding)
		return cw, nil
	case "deflate":
		fw, err := flate.NewWriter(w, flate.DefaultCompression)
		if err != nil {
			return nil, fmt.Errorf("new flate writer: %w", err)
		}
		cw := &CompressResponseWriter{}
		cw.compressWriter = fw
		cw.ResponseWriter = w
		w.Header().Set("Content-Encoding", encoding)
		return cw, nil
	default:
		return nil, errors.New("unsupported encoding")
	}
}

func (w *CompressResponseWriter) Write(data []byte) (int, error) {
	return w.compressWriter.Write(data)
}

func (w *CompressResponseWriter) Flush() {
	// Flush the compressed writer, then flush httpResponseWriter
	if f, ok := w.compressWriter.(flusher); ok {
		if err := f.Flush(); err != nil {
			logger.Errorf("Flush: %v", err)
			w.err = err
		}
		w.ResponseWriter.Flush()
	}
}

func (w *CompressResponseWriter) Error() error {
	return w.err
}

func (w *CompressResponseWriter) Close() error {
	if closer, ok := w.compressWriter.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
