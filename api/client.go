package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{client: client}
}

func (c *Client) Get(ctx context.Context, url string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodGet, url, params, result)
}

func (c *Client) Post(ctx context.Context, url string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPost, url, params, result)
}

func (c *Client) Put(ctx context.Context, url string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPut, url, params, result)
}

func (c *Client) Patch(ctx context.Context, url string, params interface{}, result interface{}) error {
	return c.call(ctx, http.MethodPatch, url, params, result)
}

func (c *Client) Delete(ctx context.Context, url string, result interface{}) error {
	return c.call(ctx, http.MethodDelete, url, nil, result)
}

func (c *Client) call(ctx context.Context, method string, url string, params interface{}, result interface{}) error {
	var body io.Reader
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return errors.Wrap(err, "marshal failed")
		}
		body = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "create request failed")
	}
	req = req.WithContext(ctx)
	req.Header = http.Header{}
	req.Header.Set(wine.ContentType, mime.JSON)
	return c.Do(req, result)
}

func (c *Client) Do(req *http.Request, result interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request failed")
	}
	return HandleResponse(resp, result)
}

func HandleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response body failed")
	}
	respObj := new(responseInfo)
	if err = json.Unmarshal(respData, respObj); err != nil {
		return errors.Wrap(err, "unmarshal response body failed")
	}
	if respObj.Error != nil {
		return gox.NewError(respObj.Error.Code, respObj.Error.Message)
	}

	if result == nil {
		return nil
	}

	jsonData, err := json.Marshal(respObj.Data)
	if err != nil {
		return errors.Wrap(err, "marshal response data failed")
	}

	if err = json.Unmarshal(jsonData, result); err != nil {
		return errors.Wrap(err, "unmarshal response data failed")
	}
	return nil
}
