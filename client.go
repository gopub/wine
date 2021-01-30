package wine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gopub/wine/urlutil"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/wine/httpvalue"
	iopkg "github.com/gopub/wine/internal/io"
)

type HeaderBuilder interface {
	Build(ctx context.Context, header http.Header) http.Header
}

type Client struct {
	client         *http.Client
	header         http.Header
	HeaderBuilder  HeaderBuilder
	RequestLogging bool
	Decoder        func(resp *http.Response, result interface{}) error

	getServerTime *ClientEndpoint
}

var DefaultClient = NewClient(http.DefaultClient)

func NewClient(client *http.Client) *Client {
	c := &Client{
		client:  client,
		header:  make(http.Header),
		Decoder: iopkg.DecodeResponse,
	}
	c.header.Set("User-Agent", "wine-client")
	return c
}

// HTTPClient returns raw http client
func (c *Client) HTTPClient() *http.Client {
	return c.client
}

func (c *Client) Endpoint(method, url string) (*ClientEndpoint, error) {
	return newClientEndpoint(c, method, url)
}

func (c *Client) MustEndpoint(method, url string) *ClientEndpoint {
	e, err := newClientEndpoint(c, method, url)
	if err != nil {
		panic(fmt.Sprintf("cannot create endpoint: %s %s", method, url))
	}
	return e
}

// Header returns shared http header
func (c *Client) Header() http.Header {
	return c.header
}

func (c *Client) SetAuthorization(credential string) {
	c.header.Set(httpvalue.Authorization, credential)
}

func (c *Client) SetBasicAuthorization(account, password string) {
	credential := []byte(account + ":" + password)
	c.header.Set(httpvalue.Authorization, "Basic "+base64.StdEncoding.EncodeToString(credential))
}

func (c *Client) Authorization() string {
	return c.header.Get(httpvalue.Authorization)
}

func (c *Client) SetDeviceID(id string) {
	c.header.Set(httpvalue.CustomDeviceID, id)
}

func (c *Client) DeviceID() string {
	return c.header.Get(httpvalue.CustomDeviceID)
}

func (c *Client) SetAppID(id string) {
	c.header.Set(httpvalue.CustomAppID, id)
}

func (c *Client) AppID() string {
	return c.header.Get(httpvalue.CustomAppID)
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
func (c *Client) Get(ctx context.Context, endpoint string, params, result interface{}) error {
	e, err := c.Endpoint(http.MethodGet, endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	return e.Call(ctx, params, result)
}

func (c *Client) Post(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	e, err := c.Endpoint(http.MethodPost, endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	return e.Call(ctx, params, result)
}

func (c *Client) Put(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	e, err := c.Endpoint(http.MethodPut, endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	return e.Call(ctx, params, result)
}

func (c *Client) Patch(ctx context.Context, endpoint string, params interface{}, result interface{}) error {
	e, err := c.Endpoint(http.MethodPatch, endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	return e.Call(ctx, params, result)
}

func (c *Client) Delete(ctx context.Context, endpoint string, query url.Values, result interface{}) error {
	e, err := c.Endpoint(http.MethodDelete, endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	return e.Call(ctx, query, result)
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
			err = errors.RequestTimeout(err.Error())
		} else {
			if tr, ok := err.(interface{ Timeout() bool }); ok && tr.Timeout() {
				err = errors.RequestTimeout(err.Error())
			} else {
				err = errors.Format(httpvalue.StatusTransportFailed, err.Error())
			}
		}
		return fmt.Errorf("cannot send request: %w", err)
	}
	return c.Decoder(resp, result)
}

func (c *Client) dumpRequest(req *http.Request) {
	logger := log.FromContext(req.Context())
	data, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Errorf("DumpRequestOut: %v", err)
		return
	}
	logger.Debugf("*** See the request as below ***\n%s", data)
}

// GetServerTime only works if server is powered by wine
func (c *Client) GetServerTime(ctx context.Context, serverURL string) (int64, error) {
	var res struct {
		Timestamp int64 `json:"timestamp"`
	}
	var err error
	if c.getServerTime == nil {
		c.getServerTime, err = c.Endpoint(http.MethodGet, urlutil.Join(serverURL, datePath))
		if err != nil {
			return 0, fmt.Errorf("create endpoint: %w", err)
		}
	}
	err = c.getServerTime.Call(ctx, nil, &res)
	return res.Timestamp, errors.Wrapf(err, "")
}

type ClientEndpoint struct {
	c      *Client
	method string
	url    *url.URL
	header http.Header
}

func newClientEndpoint(c *Client, method string, urlStr string) (*ClientEndpoint, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	method = strings.TrimSpace(method)
	method = strings.ToUpper(method)
	if method == "" {
		return nil, errors.New("method cannot be empty")
	}
	e := &ClientEndpoint{
		c:      c,
		method: method,
		url:    u,
		header: http.Header{},
	}
	return e, nil
}

func (c *ClientEndpoint) Header() http.Header {
	return c.header
}

func (c *ClientEndpoint) SetAuthorization(credential string) {
	c.header.Set(httpvalue.Authorization, credential)
}

func (c *ClientEndpoint) SetBasicAuthorization(account, password string) {
	credential := []byte(account + ":" + password)
	c.header.Set(httpvalue.Authorization, "Basic "+base64.StdEncoding.EncodeToString(credential))
}

func (c *ClientEndpoint) Call(ctx context.Context, input interface{}, output interface{}) error {
	input = conv.Indirect(input)
	var body io.Reader
	var contentType string
	switch iv := input.(type) {
	case url.Values:
		if c.method == http.MethodGet || c.method == http.MethodDelete {
			query := c.url.Query()
			for k, vl := range iv {
				for _, v := range vl {
					query.Add(k, v)
				}
			}
			c.url.RawQuery = query.Encode()
		} else if len(iv) > 0 {
			body = strings.NewReader(iv.Encode())
			contentType = httpvalue.FormURLEncoded
		}
	case nil:
		break
	default:
		contentType = httpvalue.JsonUTF8
		data, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("cannot marshal: %w", err)
		}
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequestWithContext(ctx, c.method, c.url.String(), body)
	if err != nil {
		return fmt.Errorf("create request %s %v: %w", c.method, c.url, err)
	}
	for k, v := range c.header {
		req.Header[k] = v
	}
	if contentType != "" {
		req.Header.Set(httpvalue.ContentType, contentType)
	}
	reqID := uuid.NewString()
	req.Header.Set(httpvalue.RequestID, reqID)
	err = c.c.Do(req, output)
	if err != nil {
		return fmt.Errorf("do request %s %s %v: %w", reqID, c.method, c.url, err)
	}
	return nil
}
