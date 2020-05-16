package wine

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/conv"

	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/mime"
)

const (
	StatusTransportFailed = 600
)

type timeoutReporter interface {
	Timeout() bool
}

type HeaderBuilder interface {
	Build(ctx context.Context, header http.Header) http.Header
}

type Client struct {
	client         *http.Client
	header         http.Header
	HeaderBuilder  HeaderBuilder
	RequestLogging bool
}

var DefaultClient = NewClient(http.DefaultClient)

func NewClient(client *http.Client) *Client {
	return &Client{
		client: client,
		header: make(http.Header),
	}
}

// HTTPClient returns raw http client
func (c *Client) HTTPClient() *http.Client {
	return c.client
}

// Header returns shared http header
func (c *Client) Header() http.Header {
	return c.header
}

func (c *Client) injectHeader(req *http.Request) {
	for k, vs := range c.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if c.HeaderBuilder != nil {
		req.Header = c.HeaderBuilder.Build(req.Context(), req.Header)
	}
}

// Get executes http get request created with endpoint and query
func (c *Client) Get(ctx context.Context, endpoint string, query url.Values, result interface{}) error {
	if query == nil {
		return c.call(ctx, http.MethodGet, endpoint, nil, result)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	q := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return c.call(ctx, http.MethodGet, u.String(), nil, result)
}

// GetWithBody executes http get request created with endpoint and bodyParams which will be marshaled into json data
func (c *Client) GetWithBody(ctx context.Context, endpoint string, bodyParams interface{}, result interface{}) error {
	return c.call(ctx, http.MethodGet, endpoint, bodyParams, result)
}

func (c *Client) Post(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPost, endpoint, params, result)
}

func (c *Client) Put(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPut, endpoint, params, result)
}

func (c *Client) Patch(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPatch, endpoint, params, result)
}

func (c *Client) Delete(ctx context.Context, endpoint string, query url.Values, result interface{}) error {
	if query == nil {
		return c.call(ctx, http.MethodDelete, endpoint, nil, result)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	q := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return c.call(ctx, http.MethodDelete, u.String(), nil, result)
}

func (c *Client) call(ctx context.Context, method string, endpoint string, params interface{}, result interface{}) error {
	var body io.Reader
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return fmt.Errorf("create request: %s, %s, %w", method, endpoint, err)
	}
	if params != nil {
		req.Header.Set(mime.ContentType, mime.JSON)
	}
	req = req.WithContext(ctx)
	return c.Do(req, result)
}

// Do send http request 'req' and store response data into 'result'
func (c *Client) Do(req *http.Request, result interface{}) error {
	c.injectHeader(req)

	if c.RequestLogging {
		c.dumpRequest(req)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if err == context.DeadlineExceeded {
			err = types.NewError(http.StatusRequestTimeout, err.Error())
		} else {
			if tr, ok := err.(timeoutReporter); ok && tr.Timeout() {
				err = types.NewError(http.StatusRequestTimeout, err.Error())
			} else {
				err = types.NewError(StatusTransportFailed, err.Error())
			}
		}
		return fmt.Errorf("do request: %w", err)
	}
	return DecodeResponse(resp, result)
}

func (c *Client) dumpRequest(req *http.Request) {
	logger := log.FromContext(req.Context())
	data, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Errorf("DumpRequestOut: %v", err)
		return
	}
	logger.Debugf(string(data))
}

func DecodeResponse(resp *http.Response, result interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("read resp body: %v", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return types.NewError(resp.StatusCode, string(body))
	}
	if result == nil {
		return nil
	}
	ct := mime.GetContentType(resp.Header)
	switch {
	case strings.Contains(ct, mime.JSON):
		return json.Unmarshal(body, result)
	case strings.Contains(ct, mime.Protobuf):
		m, ok := result.(proto.Message)
		if !ok {
			return fmt.Errorf("expected proto.Message instead of %T", result)
		}
		return proto.Unmarshal(body, m)
	case strings.Contains(ct, mime.Plain):
		if len(body) == 0 {
			return errors.New("no data")
		}
		return assign(result, body)
	default:
		return errors.New("invalid result")
	}
}

func assign(dataModel interface{}, body []byte) error {
	v := reflect.ValueOf(dataModel)
	if v.Kind() != reflect.Ptr {
		log.Panicf("Argument dataModel %T is not pointer", dataModel)
	}

	elem := v.Elem()
	if !elem.CanSet() {
		log.Panicf("Argument dataModel %T cannot be set", dataModel)
	}

	if tu, ok := v.Interface().(encoding.TextUnmarshaler); ok {
		err := tu.UnmarshalText(body)
		if err != nil {
			return fmt.Errorf("unmarshal text: %w", err)
		}
		return nil
	}

	if bu, ok := v.Interface().(encoding.BinaryUnmarshaler); ok {
		err := bu.UnmarshalBinary(body)
		if err != nil {
			return fmt.Errorf("unmarshal binary: %w", err)
		}
		return nil
	}

	switch elem.Kind() {
	case reflect.String:
		elem.SetString(string(body))
	case reflect.Int64,
		reflect.Int32,
		reflect.Int,
		reflect.Int16,
		reflect.Int8:
		i, err := conv.ToInt64(body)
		if err != nil {
			return fmt.Errorf("parse int: %v", err)
		}
		elem.SetInt(i)
	case reflect.Uint64,
		reflect.Uint32,
		reflect.Uint,
		reflect.Uint16,
		reflect.Uint8:
		i, err := conv.ToUint64(body)
		if err != nil {
			return fmt.Errorf("parse uint: %w", err)
		}
		elem.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := conv.ToFloat64(body)
		if err != nil {
			return fmt.Errorf("parse float: %w", err)
		}
		elem.SetFloat(i)
	case reflect.Bool:
		i, err := conv.ToBool(body)
		if err != nil {
			return fmt.Errorf("parse bool: %w", err)
		}
		elem.SetBool(i)
	default:
		return fmt.Errorf("cannot assign to dataModel %T", dataModel)
	}
	return nil
}
