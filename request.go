package wine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gopub/mapper"
	"github.com/gopub/types"
	"github.com/gopub/wine/errors"
	iopkg "github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/path"
	"github.com/gopub/wine/mime"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	request     *http.Request
	params      types.M
	body        []byte
	contentType string
	sid         string
}

// Request returns original http request
func (r *Request) Request() *http.Request {
	return r.request
}

// Params returns request parameters
func (r *Request) Params() types.M {
	return r.params
}

// Body returns request body
func (r *Request) Body() []byte {
	return r.body
}

// ContentType returns request's content type
func (r *Request) ContentType() string {
	return r.contentType
}

func (r *Request) SessionID() string {
	return r.sid
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
	return path.Normalize(r.request.URL.Path)
}

func (r *Request) UnmarshalParams(i interface{}) error {
	// Unsafe assignment, so ignore error
	data, err := json.Marshal(r.params)
	if err == nil {
		_ = json.Unmarshal(data, i)
		// As all values in query will be parsed into string type
		// mapper.Assign can convert string to int automatically
		_ = mapper.Assign(i, r.params)
	}

	if r.ContentType() == mime.JSON {
		err := json.Unmarshal(r.Body(), i)
		return errors.Wrapf(err, "unmarshal json")
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

func parseRequest(r *http.Request, maxMem types.ByteUnit) (*Request, error) {
	params, body, err := iopkg.ReadRequest(r, maxMem)
	if err != nil {
		return nil, fmt.Errorf("read request: %w", err)
	}
	return &Request{
		request:     r,
		params:      params,
		body:        body,
		contentType: mime.GetContentType(r.Header),
	}, nil
}
