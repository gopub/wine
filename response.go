package wine

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/errors"
	"github.com/gopub/wine/ctxutil"
	"github.com/gopub/wine/httpvalue"
	iopkg "github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/respond"
)

type Response = respond.Response

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
	return respond.PlainText(status, text)
}

func CSS(status int, css string) Responder {
	return respond.CSS(status, css)
}

func JSON(status int, obj interface{}) Responder {
	return respond.JSON(status, obj)
}

func JSONError(obj interface{}, err error) Responder {
	if err != nil {
		return Error(err)
	}
	return JSON(http.StatusOK, obj)
}

func Protobuf(status int, message proto.Message) Responder {
	return respond.Protobuf(status, message)
}

func ProtobufError(message proto.Message, err error) Responder {
	if err != nil {
		return Error(err)
	}
	return Protobuf(http.StatusOK, message)
}

// StreamFile creates a application/octet-stream response
func StreamFile(r io.ReadCloser, name string) Responder {
	return respond.StreamFile(r, name)
}

// BytesFile creates a application/octet-stream response
func BytesFile(b []byte, name string) Responder {
	return respond.BytesFile(b, name)
}

// StaticFile serves static files
func StaticFile(req *http.Request, path string) Responder {
	return respond.StaticFile(req, path)
}

// Image serves image
func Image(contentType string, content []byte) Responder {
	return respond.Image(contentType, content)
}

func HTML(status int, html string) Responder {
	return respond.HTML(status, html)
}

func TemplateHTML(name string, params interface{}) Responder {
	return respond.Func(func(ctx context.Context, w http.ResponseWriter) {
		ctxutil.GetTemplateManager(ctx).Execute(w, name, params)
	})
}

// Handle creates a responder with raw http handler
func Handle(req *http.Request, h http.Handler) Responder {
	return respond.Handle(req, h)
}

var _ Responder = (*errors.Error)(nil)

func Error(err error) Responder {
	if err == nil {
		return OK
	}
	if s := errors.GetCode(err); httpvalue.IsValidStatus(s) {
		return Text(s, err.Error())
	}
	return Text(http.StatusInternalServerError, err.Error())
}

type Result struct {
	Status int
	Body   []byte
}

func CompressWriter(w http.ResponseWriter, encodings ...string) (http.ResponseWriter, error) {
	if _, ok := w.(*iopkg.CompressResponseWriter); ok {
		return nil, errors.New("cannot compress writer twice")
	}
	rw, _ := w.(*iopkg.ResponseWriter)
	if rw == nil {
		return nil, errors.New("invalid response writer")
	}
	if len(encodings) == 0 {
		return nil, errors.New("missing encodings")
	}
	var err error
	for _, encoding := range encodings {
		cw, er := iopkg.NewCompressResponseWriter(rw, encoding)
		if er == nil {
			return cw, nil
		}
		err = errors.Append(err, er)
	}
	return nil, err
}
