package wine

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/types"
	"github.com/gopub/wine/httpvalue"
	iopkg "github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/router"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	request     *http.Request
	rawParams   *iopkg.RequestParams
	params      types.M
	body        []byte
	contentType string
	Model       interface{}

	uid       int64
	sensitive bool
}

// Request returns original http request
func (r *Request) Request() *http.Request {
	return r.request
}

// Body returns request parameters
func (r *Request) Params() types.M {
	return r.params
}

func (r *Request) setPathParams(p map[string]string) {
	r.rawParams.PathParams = types.M{}
	for k, v := range p {
		r.rawParams.PathParams[k] = v
		r.params[k] = v
	}
}

// Body returns request body
func (r *Request) Body() []byte {
	return r.body
}

// ContentType returns request's content type
func (r *Request) ContentType() string {
	return r.contentType
}

// Authorization returns request's Authorization in header
func (r *Request) Authorization() string {
	return r.request.Header.Get("Authorization")
}

// Bearer returns bearer token in header
func (r *Request) Bearer() string {
	s := r.Authorization()
	l := strings.Split(s, " ")
	if len(l) != 2 {
		return ""
	}
	if l[0] == "Bearer" {
		return l[1]
	}
	return ""
}

func (r *Request) BasicAccount() (user string, password string) {
	s := r.Authorization()
	l := strings.Split(s, " ")
	if len(l) != 2 {
		return
	}
	if l[0] != "Basic" {
		return
	}
	b, err := base64.StdEncoding.DecodeString(l[1])
	if err != nil {
		logger.Errorf("Decode base64 string %s: %v", l[1], err)
		return
	}
	userAndPass := strings.Split(string(b), ":")
	if len(userAndPass) != 2 {
		return
	}
	return userAndPass[0], userAndPass[1]
}

func (r *Request) NormalizedPath() string {
	return router.Normalize(r.request.URL.Path)
}

// bind: m represents the prototype of request.Model
func (r *Request) bind(m interface{}) error {
	if _, ok := m.(proto.Message); ok {
		pv := reflect.New(reflect.TypeOf(m).Elem())
		if err := proto.Unmarshal(r.body, pv.Interface().(proto.Message)); err != nil {
			return errors.BadRequest("cannot unmarshal protobuf message: %v", err)
		}
		r.Model = pv.Interface()
		return Validate(r.Model)
	}

	pv := reflect.New(reflect.TypeOf(m))
	err := conv.Assign(pv.Interface(), r.params)
	if err == nil {
		r.Model = pv.Elem().Interface()
		return Validate(r.Model)
	}

	err = errors.BadRequest("cannot assign: %v", err)
	kind := reflect.ValueOf(conv.Indirect(pv.Interface())).Kind()
	if kind == reflect.Struct || kind == reflect.Map || len(r.rawParams.BodyParams) != 0 {
		return err
	}

	if p := getSingleParam(r); p != nil && conv.Assign(pv.Interface(), p) == nil {
		r.Model = pv.Elem().Interface()
		return Validate(r.Model)
	}
	return err
}

func getSingleParam(r *Request) interface{} {
	var params = r.rawParams.PathParams
	if len(params) != 1 {
		params = r.rawParams.QueryParams
	}

	if len(params) == 1 {
		for _, val := range params {
			return val
		}
	}
	return nil
}

func (r *Request) IsWebsocket() bool {
	conn := strings.ToLower(r.Header("Connection"))
	if conn != "upgrade" {
		return false
	}
	return strings.EqualFold(r.Header("Upgrade"), "websocket")
}

func (r *Request) Header(key string) string {
	return r.request.Header.Get(key)
}

func (r *Request) SaveFormFile(name, dst string) error {
	f, _, err := r.request.FormFile(name)
	if err != nil {
		return fmt.Errorf("get form file: %w", err)
	}
	defer f.Close()

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, f)
	return fmt.Errorf("copy: %w", err)
}

func (r *Request) UserID() int64 {
	return r.uid
}

func (r *Request) SetUserID(id int64) {
	r.uid = id
}

func parseRequest(r *http.Request, maxMem types.ByteUnit) (*Request, error) {
	params, body, err := iopkg.ReadRequest(r, maxMem)
	if err != nil {
		return nil, fmt.Errorf("read request: %w", err)
	}
	return &Request{
		request:     r,
		rawParams:   params,
		params:      params.Combine(),
		body:        body,
		contentType: httpvalue.GetContentType(r.Header),
	}, nil
}
