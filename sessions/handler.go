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
	sid := req.Parameters.String("sid")
	var session Session
	if len(sid) == 0 {
		sid = wine.GetHTTP2ConnID(ctx)
		if len(sid) == 0 { // http1.x
			sid = utils.UniqueID()
			session, _ = NewSession(sid)
		}
	}

	if session == nil {
		session, _ = RestoreSession(sid)
	}

	if session == nil {
		return wine.Status(http.StatusInternalServerError)
	}

	ctx = context.WithValue(ctx, keySession, session)

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
