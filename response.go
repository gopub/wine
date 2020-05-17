package wine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/types"
	"github.com/gopub/wine/errors"
	"github.com/gopub/wine/internal/respond"
)

// Responder interface is used by Wine server to write response to the client
type Responder interface {
	// Respond will be called to write status/body to http response writer
	Respond(ctx context.Context, w http.ResponseWriter)
}

func HandleResponder(r Responder) Handler {
	return HandlerFunc(func(ctx context.Context, req *Request) Responder {
		return r
	})
}

type ResponderFunc func(ctx context.Context, w http.ResponseWriter)

func (f ResponderFunc) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}

func (f ResponderFunc) HandleRequest(ctx context.Context, req *Request) Responder {
	return f
}

var OK = Status(http.StatusOK)

func Status(s int) Responder {
	return respond.Status(s)
}

func Redirect(location string, permanent bool) Responder {
	return respond.Redirect(location, permanent)
}

func Text(status int, text string, args ...interface{}) Responder {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	return respond.Text(status, text)
}

func JSON(status int, value interface{}) Responder {
	return respond.JSON(status, value)
}

func Protobuf(status int, message proto.Message) Responder {
	return respond.Protobuf(status, message)
}

// StreamFile creates a application/octet-stream response
func StreamFile(r io.Reader, name string) Responder {
	return respond.StreamFile(r, name)
}

// File creates a application/octet-stream response
func File(b []byte, name string) Responder {
	return respond.File(b, name)
}

// StaticFile serves static files
func StaticFile(req *http.Request, path string) Responder {
	return respond.StaticFile(req, path)
}

func HTML(status int, html string) Responder {
	return respond.HTML(status, html)
}

func TemplateHTML(name string, params interface{}) Responder {
	return respond.Func(func(ctx context.Context, w http.ResponseWriter) {
		getTemplateManager(ctx).Execute(w, name, params)
	})
}

// Handle creates a responder with raw http handler
func Handle(req *http.Request, h http.Handler) Responder {
	return respond.Handle(req, h)
}

func Error(err error) Responder {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}

	if reflect.TypeOf(errors.New("")) == reflect.TypeOf(err) {
		return Text(http.StatusInternalServerError, err.Error())
	}

	if err == errors.NotExists || err == sql.ErrNoRows {
		return Text(http.StatusNotFound, err.Error())
	}

	if s := extractStatus(err); s > 0 {
		return Text(s, err.Error())
	}
	return Text(http.StatusInternalServerError, err.Error())
}

func extractStatus(err error) int {
	if v := reflect.ValueOf(err); v.Kind() == reflect.Int {
		n := int(v.Int())
		if n > 0 {
			return n
		}
		return 0
	}

	keys := []string{"status", "Status", "status_code", "StatusCode", "statusCode", "code", "Code"}
	i := indirect(err)
	k := reflect.ValueOf(i).Kind()
	if k != reflect.Struct && k != reflect.Map {
		return 0
	}

	b, jErr := json.Marshal(i)
	if jErr != nil {
		logger.Errorf("Marshal: %v", err)
		return 0
	}
	var m types.M
	jErr = json.Unmarshal(b, &m)
	if jErr != nil {
		logger.Errorf("Unmarshal: %v", err)
		return 0
	}

	for _, k := range keys {
		s := m.Int(k)
		if s > 0 {
			return s
		}
	}
	return http.StatusInternalServerError
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
