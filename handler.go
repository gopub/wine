package wine

import (
	"container/list"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine/internal/resource"
	"github.com/gopub/wine/mime"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req *Request, next Invoker) Responder
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req *Request, next Invoker) Responder

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req *Request, next Invoker) Responder {
	return h(ctx, req, next)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

// Invoker defines the function to be called in order to pass on the request
type Invoker func(ctx context.Context, req *Request) Responder

type InvokerList struct {
	handlers *list.List
	current  *list.Element
}

func newInvokerList(handlers *list.List) *InvokerList {
	l := &InvokerList{
		handlers: handlers,
		current:  handlers.Front(),
	}
	return l
}

func (l *InvokerList) Invoke(ctx context.Context, req *Request) Responder {
	if l.current == nil {
		return nil
	}
	h := l.current.Value.(Handler)
	l.current = l.current.Next()
	return h.HandleRequest(ctx, req, l.Invoke)
}

// Some built-in handlers
func handleFavIcon(ctx context.Context, req *Request, next Invoker) Responder {
	return ResponderFunc(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header()[mime.ContentType] = []string{"image/x-icon"}
		rw.WriteHeader(http.StatusOK)
		if err := gox.WriteAll(rw, resource.Favicon); err != nil {
			log.FromContext(ctx).Errorf("Write all: %v", err)
		}
	})
}

func handleNotFound(ctx context.Context, req *Request, next Invoker) Responder {
	return Text(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func handleNotImplemented(ctx context.Context, req *Request, next Invoker) Responder {
	return Text(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func toHandlers(fs ...HandlerFunc) []Handler {
	l := make([]Handler, len(fs))
	for i, f := range fs {
		l[i] = f
	}
	return l
}

func toHandlerList(hs ...Handler) *list.List {
	l := list.New()
	for _, h := range hs {
		l.PushBack(h)
	}
	return l
}

func handlerListToString(l *list.List) string {
	s := new(strings.Builder)
	for h := l.Front(); h != nil; h = h.Next() {
		if s.Len() > 0 {
			s.WriteString(", ")
		}

		var name string
		if s, ok := h.Value.(fmt.Stringer); ok {
			name = s.String()
		} else {
			name = reflect.TypeOf(h.Value).Name()
		}

		if strings.HasSuffix(name, "-fm") {
			name = name[0 : len(name)-3]
		}
		s.WriteString(shortenFilename(name))
	}
	return s.String()
}

func shortenFilename(filename string) string {
	if len(log.PackagePath) > 0 {
		filename = strings.TrimPrefix(filename, log.PackagePath)
	} else {
		start := strings.Index(filename, log.GoSrc)
		if start > 0 {
			start += len(log.GoSrc)
			filename = filename[start:]
		}
	}

	names := strings.Split(filename, "/")
	for i := 0; i < len(names)-1; i++ {
		names[i] = names[i][0:1]
	}
	filename = strings.Join(names, "/")
	return filename
}
