package wine

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gopub/log"
)

// BasicAuthUser is key for basic auth user
const BasicAuthUser = "basic_auth_user"

// BasicAuth returns a basic auth interceptor
func BasicAuth(userToPassword map[string]string, realm string) HandlerFunc {
	if len(userToPassword) == 0 {
		log.Panic("userToPassword is empty")
	}

	userToAuthInfo := make(map[string]string)
	for user, password := range userToPassword {
		if len(user) == 0 || len(password) == 0 {
			log.Panic("Empty user or password")
		}
		info := user + ":" + password
		userToAuthInfo[user] = "Basic " + base64.StdEncoding.EncodeToString([]byte(info))
	}

	authHeaderValue := "Basic realm=" + strconv.Quote(realm)
	return func(ctx context.Context, req *Request, next Invoker) Responsible {
		authValue := req.HTTPRequest.Header.Get("Authorization")
		for user, info := range userToAuthInfo {
			if info == authValue {
				ctx = context.WithValue(ctx, BasicAuthUser, user)
				return next(ctx, req)
			}
		}

		header := make(http.Header)
		header.Set("WWW-Authenticate", authHeaderValue)
		return NewResponse(http.StatusUnauthorized, header, nil)
	}
}
