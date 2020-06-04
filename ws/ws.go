package ws

import (
	"context"

	"github.com/gopub/log"
	"github.com/gopub/wine/errors"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

type contextKey int

// Context keys
const (
	ckNextHandler contextKey = iota + 1
)

func Next(ctx context.Context, req interface{}) (interface{}, error) {
	i, _ := ctx.Value(ckNextHandler).(Handler)
	if i == nil {
		return nil, errors.NotImplemented("")
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ckNextHandler, h)
}

const (
	datePath     = "_wine/date"
	uptimePath   = "_wine/uptime"
	versionPath  = "_wine/version"
	endpointPath = "_wine/endpoints"
	echoPath     = "_wine/echo"
)

var reservedPaths = map[string]bool{
	datePath:     true,
	uptimePath:   true,
	versionPath:  true,
	endpointPath: true,
	echoPath:     true,
}
