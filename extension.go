package wine

import (
	"context"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/wine/ctxutil"
)

const (
	datePath     = "_wine/date"
	uptimePath   = "_wine/uptime"
	versionPath  = "_wine/version"
	endpointPath = "_wine/endpoints"
	echoPath     = "_wine/echo"
)

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

func checkAuth(ctx context.Context, req *Request) Responder {
	if ctxutil.GetUserID(ctx) <= 0 {
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
