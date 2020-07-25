package wine

import (
	"net/http"

	"github.com/gopub/conv"
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

func (h *Header) ExposeHeaders(headers ...string) {
	h.Header["Access-Control-Expose-Headers"] = headers
}

func (h *Header) AllowMethods(methods ...string) {
	h.Header["Access-Control-Allow-Methods"] = methods
}

func (h *Header) AllowCredentials(b bool) {
	h.Header.Set("Access-Control-Allow-Credentials", conv.MustString(b))
}
