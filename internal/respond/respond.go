package respond

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/conv"
	"github.com/gopub/log"
	"github.com/gopub/wine/mime"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

// Response holds all the http response information
// Value and headers except the status code can be modified before sent to the client
type Response struct {
	status int
	header http.Header
	value  interface{}
}

// Respond writes header and body to response writer w
func (r *Response) Respond(ctx context.Context, w http.ResponseWriter) {
	body, err := r.marshalBody()
	if err != nil {
		logger.Errorf("Cannot marshal body: %v", err)
		errResp := PlainText(http.StatusInternalServerError, err.Error())
		errResp.Respond(ctx, w)
		return
	}

	for k, v := range r.header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.status)
	if _, err = w.Write(body); err != nil {
		logger.Errorf("Cannot write body: %v", err)
	}
}

func (r *Response) marshalBody() ([]byte, error) {
	if r.value == nil {
		return nil, nil
	}
	ct := r.header.Get(mime.ContentType)
	switch {
	case strings.Contains(ct, mime.JSON):
		return json.Marshal(r.value)
	case strings.Contains(ct, mime.Protobuf):
		if m, ok := r.value.(proto.Message); ok {
			return proto.Marshal(m)
		}
		return nil, fmt.Errorf("value is %T instead of proto.message", r.value)
	case strings.Contains(ct, mime.Plain) || strings.Contains(ct, mime.HTML) ||
		strings.Contains(ct, mime.XML) || strings.Contains(ct, mime.XML2) || strings.Contains(ct, mime.CSS):
		s, err := conv.ToString(r.value)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	default:
		return nil, fmt.Errorf("%s is not supported", ct)
	}
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Value() interface{} {
	return r.value
}

func (r *Response) SetValue(v interface{}) {
	r.value = v
}

type Func func(ctx context.Context, w http.ResponseWriter)

func (f Func) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}
