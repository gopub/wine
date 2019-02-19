package mime

import "strings"

// types
//const (
//	Text        = "text"
//	Multipart   = "multipart"
//	Application = "application"
//	Message     = "message"
//	Image       = "image"
//	Audio       = "audio"
//	Video       = "video"
//)

// subtypes
const (
	Plain = "text/plain"
	HTML  = "text/html"
	XML   = "text/xml"
	XML2  = "application/xml"
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

func TypeOfFileExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case "txt", "md":
		return Plain
	case "html", "htm":
		return HTML
	case "json":
		return JSON
	case "zip":
		return GZIP
	case "jpg", "jpeg":
		return JPEG
	case "png":
		return PNG
	case "webp":
		return WEBP
	case "pdf":
		return PDF
	case "word":
		return MSWord
	default:
		return ""
	}
}
