package wine

import (
	"github.com/gopub/conv"
	"net/http"
	"strings"
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
	h.Header["Access-Control-Allow-Origin"] = origins
}

func (h *Header) AllowMethods(methods ...string) {
	// Combine multiple values separated by comma. Multiple lines style is also fine.
	h.Header.Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}

func (h *Header) AllowCredentials(b bool) {
	h.Header.Set("Access-Control-Allow-Credentials", conv.MustString(b))
}

func (h *Header) AllowHeaders(headers ...string) {
	h.Header.Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
}

func (h *Header) ExposeHeaders(headers ...string) {
	h.Header.Set("Access-Control-Expose-Headers", strings.Join(headers, ","))
}
