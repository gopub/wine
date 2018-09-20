package wine

import (
	"net/http"

	"github.com/gopub/types"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request interface {
	SetValue(key string, value interface{})
	Value(key string) interface{}
	RawRequest() *http.Request
	Parameters() types.M
	TemplateManager() *TemplateManager
}

// requestImpl is a wrapper of http.requestImpl, aims to provide more convenient interface
type requestImpl struct {
	req             *http.Request
	reqParams       types.M
	keyValues       types.M
	templateManager *TemplateManager
}

// Set sets key:value
func (c *requestImpl) SetValue(key string, value interface{}) {
	c.keyValues[key] = value
}

// Get returns value for key
func (c *requestImpl) Value(key string) interface{} {
	return c.keyValues[key]
}

// requestImpl returns request
func (c *requestImpl) RawRequest() *http.Request {
	return c.req
}

// Params returns request's parameters including queries, body
func (c *requestImpl) Parameters() types.M {
	return c.reqParams
}

// Params returns request's parameters including queries, body
func (c *requestImpl) TemplateManager() *TemplateManager {
	return c.templateManager
}
