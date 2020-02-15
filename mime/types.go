package mime

import (
	"net/http"
)

// gox
//const (
//	Text        = "text"
//	Multipart   = "multipart"
//	Application = "application"
//	Message     = "message"
//	Image       = "image"
//	Audio       = "audio"
//	Video       = "video"
//)

// subgox
const (
	Plain = "text/plain"
	HTML  = "text/html"
	XML2  = "text/xml"
	XML   = "application/xml"
	XHTML = "application/xhtml+xml"

	FormData = "multipart/form-data"
	GIF      = "image/gif"
	JPEG     = "image/jpeg"
	PNG      = "image/png"
	WEBP     = "image/webp"

	MPEG = "video/mpeg"

	FormURLEncoded = "application/x-www-form-urlencoded"
	OctetStream    = "application/octet-stream"
	JSON           = "application/json"
	PDF            = "application/pdf"
	MSWord         = "application/msword"
	GZIP           = "application/x-gzip"
)

const (
	ContentType        = "Content-Type"
	ContentDisposition = "Content-Disposition"
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
