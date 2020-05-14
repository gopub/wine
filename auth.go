package wine

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gopub/log"
)

// NewBasicAuthHandler returns a basic auth interceptor
func NewBasicAuthHandler(userToPassword map[string]string, realm string) HandlerFunc {
	if len(userToPassword) == 0 {
		log.Panic("userToPassword is empty")
	}

	userToAuthInfo := make(map[string]string)
	for user, password := range userToPassword {
		if user == "" || password == "" {
			log.Panic("Empty user or password")
		}
		info := user + ":" + password
		userToAuthInfo[user] = "Basic " + base64.StdEncoding.EncodeToString([]byte(info))
	}

	return func(ctx context.Context, req *Request) Responder {
		a := req.Authorization()
		for user, info := range userToAuthInfo {
			if info == a {
				ctx = withBasicAuthUser(ctx, user)
				return Next(ctx)(ctx, req)
			}
		}
		return RequireBasicAuth(realm)
	}
}

func RequireBasicAuth(realm string) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		a := "Basic realm=" + strconv.Quote(realm)
		w.Header().Set("WWW-Authenticate", a)
		w.WriteHeader(http.StatusUnauthorized)
	})
}
