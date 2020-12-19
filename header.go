package wine

import (
	"net/http"
	"strings"

	"github.com/gopub/conv"
	"github.com/gopub/wine/httpvalue"
)

type Header struct {
	http.Header
}

func NewHeader() *Header {
	return &Header{
		Header: make(http.Header),
	}
}

func (h *Header) WriteTo(rw http.ResponseWriter) {
	for k, v := range h.Header {
		rw.Header()[k] = v
	}
}

func (h *Header) Clone() *Header {
	c := &Header{
		Header: make(http.Header),
	}
	for k, v := range h.Header {
		c.Header[k] = v
	}
	return c
}

func (h *Header) AllowOrigins(origins ...string) {
	h.Header[httpvalue.ACLAllowOrigin] = origins
}

func (h *Header) AllowMethods(methods ...string) {
	// Combine multiple values separated by comma. Multiple lines style is also fine.
	h.Header.Set(httpvalue.ACLAllowMethods, strings.Join(methods, ","))
}

func (h *Header) AllowCredentials(b bool) {
	h.Header.Set(httpvalue.ACLAllowCredentials, conv.MustString(b))
}

func (h *Header) AllowHeaders(headers ...string) {
	h.Header.Set(httpvalue.ACLAllowHeaders, strings.Join(headers, ","))
}

func (h *Header) ExposeHeaders(headers ...string) {
	h.Header.Set(httpvalue.ACLExposeHeaders, strings.Join(headers, ","))
}
