package wine

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gopub/wine/internal/respond"
)

// NewBasicAuthHandler returns a basic auth interceptor
func NewBasicAuthHandler(userToPassword map[string]string, realm string) HandlerFunc {
	if len(userToPassword) == 0 {
		logger.Panic("userToPassword is empty")
	}

	userToAuthorization := make(map[string]string)
	for user, password := range userToPassword {
		if user == "" || password == "" {
			logger.Panic("Empty user or password")
		}
		info := user + ":" + password
		userToAuthorization[user] = "Basic " + base64.StdEncoding.EncodeToString([]byte(info))
	}

	return func(ctx context.Context, req *Request) Responder {
		a := req.Authorization()
		for user, auth := range userToAuthorization {
			if auth == a {
				ctx = WithUser(ctx, user)
				return Next(ctx, req)
			}
		}
		return RequireBasicAuth(realm)
	}
}

func RequireBasicAuth(realm string) Responder {
	return respond.Func(func(ctx context.Context, w http.ResponseWriter) {
		a := "Basic realm=" + strconv.Quote(realm)
		w.Header().Set("WWW-Authenticate", a)
		w.WriteHeader(http.StatusUnauthorized)
	})
}
