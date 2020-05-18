package wine

import (
	"container/list"
	"context"
	"net/http"
	"net/http/httputil"
	"reflect"
	"runtime"
	"time"

	"github.com/gopub/types"
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

type handlerElem list.Element

func (h *handlerElem) Next() *handlerElem {
	return (*handlerElem)((*list.Element)(h).Next())
}

func (h *handlerElem) HandleRequest(ctx context.Context, req *Request) Responder {
	return h.Value.(Handler).HandleRequest(withNextHandler(ctx, h.Next()), req)
}

func linkHandlers(handlers ...Handler) *list.List {
	hl := list.New()
	for _, h := range handlers {
		hl.PushBack(h)
	}
	return hl
}

func linkHandlerFuncs(funcs ...HandlerFunc) *list.List {
	hl := list.New()
	for _, h := range funcs {
		hl.PushBack(h)
	}
	return hl
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

func handleAuth(ctx context.Context, req *Request) Responder {
	if GetUserID(ctx) <= 0 {
		return Text(http.StatusUnauthorized, "")
	}
	return Next(ctx, req)
}

func newUptimeHandler() HandlerFunc {
	upAt := time.Now()
	return func(_ context.Context, _ *Request) Responder {
		return Text(http.StatusOK, time.Now().Sub(upAt).String())
	}
}
