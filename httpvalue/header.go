package httpvalue

import (
	"fmt"
	"net/http"
	"strings"
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
	Location            = "Location"
	Cookies             = "Cookies"
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

func GetAcceptEncodings(h http.Header) []string {
	a := strings.Split(h.Get(AcceptEncoding), ",")
	for i, s := range a {
		a[i] = strings.TrimSpace(s)
	}

	// Remove empty strings
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] == "" {
			a = append(a[:i], a[i+1:]...)
		}
	}
	return a
}
