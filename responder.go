package wine

import (
	"context"
	"encoding/json"
	"github.com/gopub/log"
	"github.com/gopub/utils"
	"net/http"
	"strings"
)

type Responsible interface {
	Respond(ctx context.Context, w http.ResponseWriter)
}

type ResponsibleFunc func(ctx context.Context, w http.ResponseWriter)

func (f ResponsibleFunc) Respond(ctx context.Context, w http.ResponseWriter)  {
	f(ctx, w)
}

type Response interface {
	Responsible
	Status() int
	Header() http.Header
	Value() interface{}
	SetValue(v interface{})
}

type responseImpl struct {
	status int
	header http.Header
	value interface{}
}

func (r *responseImpl) Respond(ctx context.Context, w http.ResponseWriter)  {
	body, ok := r.value.([]byte)
	if !ok {
		body = r.getBytes()
	}
	w.Write(body)
	for k, v := range r.header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.status)
}

func (r *responseImpl) getBytes() []byte  {
	if body, ok := r.value.([]byte); ok {
		return body
	}

	contentType := r.header.Get(utils.ContentType)

	switch  {
	case strings.Contains(contentType, utils.MIMEJSON):
		if r.value != nil {
			body, err := json.Marshal(r.value)
			if err != nil {
				log.Error(err)
			} else {
				return body
			}
		}
	case strings.Contains(contentType, utils.MIMETEXT):
		fallthrough
	case strings.Contains(contentType, utils.MIMEHTML):
		if s, ok := r.value.(string); ok {
			return []byte(s)
		}
	default:
		log.Warn("unsupported Content-Type:", contentType)
	}

	return nil
}

func (r *responseImpl) Status() int {
	return r.status
}

func (r *responseImpl) Header() http.Header {
	return r.header
}

func (r *responseImpl) Value() interface{} {
	return r.value
}

func (r *responseImpl) SetValue(v interface{}) {
	r.value = v
}

func Status(status int) Response  {
	return Text(status, http.StatusText(status))
}

func Text(status int, text string) Response  {
	header := make(http.Header)
	header.Set(utils.ContentType, utils.MIMETEXT)
	return &responseImpl{
		status:status,
		header:header,
		value:text,
	}
}

func HTML(status int, html string) Response  {
	header := make(http.Header)
	header.Set(utils.ContentType, utils.MIMEHTML)
	return &responseImpl{
		status:status,
		header:header,
		value:html,
	}
}

func JSON(status int, value interface{}) Response  {
	header := make(http.Header)
	header.Set(utils.ContentType, utils.MIMEJSON)
	return &responseImpl{
		status:status,
		header:header,
		value:value,
	}
}

