package httpvalue

import (
	"fmt"
	"net/http"
)

const (
	Authorization       = "Authorization"
	AcceptEncoding      = "Accept-Encoding"
	ACLAllowCredentials = "Access-Control-Allow-Credentials"
	ACLAllowHeaders     = "Access-Control-Allow-Headers"
	ACLAllowMethods     = "Access-Control-Allow-Methods"
	ACLAllowOrigin      = "Access-Control-Allow-Origin"
	ACLExposeHeaders    = "Access-Control-Expose-Headers"
	ContentType         = "Content-Type"
	ContentDisposition  = "Content-Disposition"
	ContentEncoding     = "Content-Encoding"
)

func GetContentType(h http.Header) string {
	t := h.Get(ContentType)
	for i, ch := range t {
		if ch == ' ' || ch == ';' {
			t = t[:i]
			break
		}
	}
	return t
}

// FileAttachment returns value for Content-Disposition
// e.g. Content-Disposition: attachment; filename=test.txt
func FileAttachment(filename string) string {
	return fmt.Sprintf(`attachment; filename="%s"`, filename)
}
