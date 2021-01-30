package session

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gopub/wine"
)

func NewHandler(provider Provider, options *Options) wine.HandlerFunc {
	if options == nil {
		options = DefaultOptions()
	}

	return func(ctx context.Context, req *wine.Request) wine.Responder {
		sid := req.Params().String(options.keyForID)
		var ses Session
		var err error
		if sid != "" {
			ses, err = provider.Get(ctx, sid)
		} else {
			sid = uuid.NewString()
		}

		if ses == nil {
			ses, err = provider.Create(ctx, sid, options)
			if err != nil {
				return wine.Error(err)
			}
		}

		ctx = withSession(ctx, ses)

		cookie := &http.Cookie{
			Name:     options.keyForID,
			Value:    sid,
			Expires:  time.Now().Add(options.TTL),
			Path:     options.CookiePath,
			HttpOnly: options.CookieHttpOnly,
		}

		resp := wine.Next(ctx, req)

		return wine.Handle(req.Request(), http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			http.SetCookie(writer, cookie)
			// Write to Header in case cookie is disabled by some browsers
			writer.Header().Set(options.keyForID, sid)
			resp.Respond(ctx, writer)
		}))
	}
}
