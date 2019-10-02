package session

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine"
)

func InitSession(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
	sid := req.Params().String(wine.SessionName)
	var session Session

	if len(sid) > 0 {
		session, _ = RestoreSession(sid)
	}

	if session == nil {
		if len(sid) != sidLength {
			if len(sid) > 0 {
				logger.Error("Invalid length of sid")
			}
			sid = gox.UniqueID()
		}
		var err error
		session, err = NewSession(sid)
		if err != nil {
			logger.Fatal("Cannot create session: %v", err)
			return wine.Status(http.StatusInternalServerError)
		}
	}

	if session == nil {
		logger.Fatal("Session is nil")
		return wine.Status(http.StatusInternalServerError)
	}

	ctx = context.WithValue(ctx, keySession, session)

	resp := next(ctx, req)

	switch v := resp.(type) {
	case wine.Response:
		v.Header().Set(wine.SessionName, sid)
		expire := time.Now().Add(time.Minute * 60)
		cookie := &http.Cookie{
			Name:    wine.SessionName,
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
