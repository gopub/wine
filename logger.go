package wine

import (
	"context"
	"github.com/gopub/log"
	"time"
)

var (
	_greenColor   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	_whiteColor   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	_yellowColor  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	_redColor     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	_blueColor    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	_magentaColor = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	_cyanColor    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	_resetColor   = string([]byte{27, 91, 48, 109})
)

// Logger calculates cost time and output to console
func Logger(ctx context.Context, request Request, responder Responder) bool {
	st := time.Now()
	handled := responder.Next(ctx, request, responder)
	cost := float32(time.Since(st)/time.Microsecond) / 1000.0
	log.Infof("%.3fms %v%s%v %s",
		cost,
		_greenColor,
		request.RawRequest().Method,
		_resetColor,
		request.RawRequest().RequestURI)
	return handled
}
