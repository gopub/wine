package stream

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"syscall"

	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

type logResponseWriter struct {
	w    http.ResponseWriter
	done chan types.Void
	err  error
}

func (w *logResponseWriter) Write(data []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.w.Write(data)
	if err != nil {
		w.err = err
		w.done <- types.Void{}
		return n, err
	}
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
	if e, ok := w.w.(interface{ Error() error }); ok {
		if e.Error() != nil {
			w.err = e.Error()
			w.done <- types.Void{}
			return 0, w.err
		}
	}
	return n, err
}

func Log(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
	return wine.ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(mime.ContentType, mime.JsonUTF8)
		o := &logResponseWriter{
			w:    w,
			done: make(chan types.Void),
		}
		_, err := o.Write([]byte(strings.Repeat("-", 80) + "\n"))
		if err != nil {
			wine.Logger().Errorf("Write: %v", err)
			return
		}
		log.Default().AddOutput(o)
		<-o.done
		log.Default().RemoveOutput(w)
		if errors.Is(o.err, syscall.EPIPE) {
			return
		}
		wine.Logger().Errorf("Error: %v", o.err)
	})
}
