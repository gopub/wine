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
	"github.com/gopub/wine/httpvalue"
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

	marshaledValue []byte
}

// Respond writes header and body to response writer w
func (r *Response) Respond(ctx context.Context, w http.ResponseWriter) {
	if r.marshaledValue == nil {
		data, err := r.marshal()
		if err != nil {
			resp := PlainText(http.StatusInternalServerError, err.Error())
			resp.Respond(ctx, w)
			return
		}
		r.marshaledValue = data
	}

	for k, v := range r.header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.status)
	if _, err := w.Write(r.marshaledValue); err != nil {
		logger.Errorf("Cannot write: %v", err)
	}
}

func (r *Response) marshal() ([]byte, error) {
	if r.value == nil {
		return nil, nil
	}
	ct := r.header.Get(httpvalue.ContentType)
	switch {
	case strings.Contains(ct, httpvalue.JSON):
		b, err := json.Marshal(r.value)
		if err != nil {
			return nil, fmt.Errorf("marshal json: %w", err)
		}
		return b, nil
	case strings.Contains(ct, httpvalue.Protobuf):
		if m, ok := r.value.(proto.Message); ok {
			b, err := proto.Marshal(m)
			if err != nil {
				return nil, fmt.Errorf("marshal protobuf: %w", err)
			}
			return b, nil
		}
		return nil, fmt.Errorf("value is %T instead of proto.message", r.value)
	case strings.Contains(ct, httpvalue.Plain) ||
		strings.Contains(ct, httpvalue.HTML) ||
		strings.Contains(ct, httpvalue.XML) ||
		strings.Contains(ct, httpvalue.XML2) ||
		strings.Contains(ct, httpvalue.CSS):
		s, err := conv.ToString(r.value)
		if err != nil {
			return nil, fmt.Errorf("convert to string: %w", err)
		}
		return []byte(s), nil
	default:
		b, err := conv.ToBytes(r.value)
		if err != nil {
			return nil, fmt.Errorf("convert to []byte: %w", err)
		}
		return b, nil
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

func (r *Response) ContentLength() int {
	if r.value == nil {
		return 0
	}
	if r.marshaledValue == nil {
		r.marshaledValue, _ = r.marshal()
	}
	return len(r.marshaledValue)
}

func (r *Response) SetValue(v interface{}) {
	r.value = v
}

type Func func(ctx context.Context, w http.ResponseWriter)

func (f Func) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}
