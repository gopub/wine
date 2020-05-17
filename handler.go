package wine

import (
	"container/list"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/internal/resource"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/mime"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req *Request) Responder
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req *Request) Responder

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req *Request) Responder {
	return h(ctx, req)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

type handlerChain struct {
	handlers *list.List
	current  *list.Element
}

func toHandlerChain(handlers ...Handler) *handlerChain {
	hl := list.New()
	for _, h := range handlers {
		hl.PushBack(h)
	}
	return newHandlerChain(hl)
}

func newHandlerChain(hl *list.List) *handlerChain {
	l := &handlerChain{
		handlers: hl,
		current:  hl.Front(),
	}
	return l
}

func (c *handlerChain) HandleRequest(ctx context.Context, req *Request) Responder {
	if c.current == nil {
		return nil
	}
	h := c.current.Value.(Handler)
	c.current = c.current.Next()
	return h.HandleRequest(ctx, req)
}

// Some built-in handlers
func handleFavIcon(_ context.Context, _ *Request) Responder {
	return respond.Func(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header().Set(mime.ContentType, mime.ICON)
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(resource.Favicon); err != nil {
			log.FromContext(ctx).Errorf("Write: %v", err)
		}
	})
}

func handleEcho(_ context.Context, req *Request) Responder {
	v, err := httputil.DumpRequest(req.request, true)
	if err != nil {
		return Text(http.StatusInternalServerError, err.Error())
	}
	return Text(http.StatusOK, string(v))
}

func handleDate(_ context.Context, req *Request) Responder {
	ts := req.Params().DefaultInt64("timestamp", time.Now().Unix())
	t := time.Unix(ts, 0)
	res := types.M{
		"timestamp": t.Unix(),
		"time":      t.Format("15:04:05"),
		"date":      t.Format("2006-01-02"),
		"zone":      t.Format("-0700"),
		"weekday":   t.Format("Mon"),
		"month":     t.Format("Jan"),
	}
	return JSON(http.StatusOK, res)
}

func toHandlers(fs ...HandlerFunc) []Handler {
	l := make([]Handler, len(fs))
	for i, f := range fs {
		l[i] = f
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
			name = name[:len(name)-3]
		}
		s.WriteString(shortenFilename(name))
	}
	return s.String()
}

func shortenFilename(filename string) string {
	var trimmed string
	if len(log.PackagePath) > 0 {
		trimmed = strings.TrimPrefix(filename, log.PackagePath)
	} else {
		start := strings.Index(filename, log.GoSrc)
		if start > 0 {
			start += len(log.GoSrc)
			trimmed = filename[start:]
		}
	}

	l := strings.Split(trimmed, "/")
	for i := 0; i < len(l)-1; i++ {
		l[i] = l[i][0:1]
	}
	return strings.Join(l, "/")
}
