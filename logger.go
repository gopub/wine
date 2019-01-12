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

type logHandler struct {
	logger log.Logger
}

func Logger() Handler {
	l := &logHandler{}
	l.logger = log.Default().Derive()
	l.logger.SetFlags(log.LstdFlags ^ log.Lfunction ^ log.Lshortfile)
	return l
}

// Logger calculates cost time and output to console
func (l *logHandler) HandleRequest(ctx context.Context, req *Request, next Invoker) Responsible {
	st := time.Now()
	resp := next(ctx, req)
	cost := float32(time.Since(st)/time.Microsecond) / 1000.0
	//l.logger.Infof("[WINE] %s %v%s%v %s %v%.3fms%v",
	//	req.HTTPRequest.RemoteAddr,
	//	_greenColor,
	//	req.HTTPRequest.Method,
	//	_resetColor,
	//	req.HTTPRequest.RequestURI,
	//	_yellowColor,
	//	cost,
	//	_resetColor)

	l.logger.Infof("[WINE] %s %s %s %.3fms",
		req.HTTPRequest.RemoteAddr,
		req.HTTPRequest.Method,
		req.HTTPRequest.RequestURI,
		cost)
	return resp
}
