package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/mime"
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
	UseResultModel bool
}

var DefaultClient = NewClient(http.DefaultClient)

func NewClient(client *http.Client) *Client {
	return &Client{
		client:         client,
		header:         make(http.Header),
		UseResultModel: true,
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
	return ParseResult(resp, result, !c.UseResultModel)
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
