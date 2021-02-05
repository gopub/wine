package session

import (
	"context"
	"github.com/gopub/errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gopub/wine"
)

func NewHandler(provider Provider, options *Options) wine.HandlerFunc {
	if options == nil {
		options = DefaultOptions()
	}

	cookieIDKey := options.Name + "id"
	headerIDKey := "X-" + strings.ToUpper(cookieIDKey[0:1]) + cookieIDKey[1:]

	return func(ctx context.Context, req *wine.Request) wine.Responder {
		sid := req.Params().String(cookieIDKey)
		var ses Session
		var err error
		if sid != "" {
			ses, err = provider.Get(ctx, sid)
			if err != nil && !errors.IsNotExist(err) {
				return wine.Error(err)
			}
		} else {
			sid = uuid.NewString()
		}

		if ses == nil {
			ses, err = provider.Create(ctx, sid, options.TTL)
		} else {
			err = ses.SetTTL(options.TTL)
		}

		if err != nil {
			return wine.Error(err)
		}

		ctx = withSession(ctx, ses)

		cookie := &http.Cookie{
			Name:     cookieIDKey,
			Value:    sid,
			Expires:  time.Now().Add(options.TTL),
			Path:     options.CookiePath,
			HttpOnly: options.CookieHttpOnly,
		}

		resp := wine.Next(ctx, req)

		return wine.Handle(req.Request(), http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.SetCookie(writer, cookie)
			// Write to Header in case cookie is disabled by some browsers
			writer.Header().Set(headerIDKey, sid)
			resp.Respond(ctx, writer)
		}))
	}
}
