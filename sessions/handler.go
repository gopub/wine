package sessions

import (
	"context"
	"github.com/gopub/log"
	"github.com/gopub/utils"
	"github.com/gopub/wine"
	"net/http"
	"reflect"
	"time"
)

func InitSession(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
	sid := req.Parameters.String(KeySid)
	if len(sid) == 0 {
		var ok bool
		sid, ok = ctx.Value(wine.keyHTTP2ConnID).(string)
		if !ok || len(sid) == 0 { // http1.x
			sid = utils.UniqueID()
		}
	}

	ctx = context.WithValue(ctx, KeySid, sid)

	resp := next(ctx, req)

	switch v := resp.(type) {
	case wine.Response:
		v.Header().Set("sid", sid)
		expire := time.Now().Add(time.Minute * 60)
		cookie := &http.Cookie{
			Name:    "sid",
			Value:   sid,
			Expires: expire,
			Path:    "/",
		}

		return wine.ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
			http.SetCookie(w, cookie)
			v.Respond(ctx, w)
		})
	default:
		log.Warnf("Unable to write sid into header/cookie along with response type:%v", reflect.TypeOf(resp))
		return v
	}
}
