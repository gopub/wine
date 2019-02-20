package mime

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

const charsetSuffix = "; charset=utf-8"

const (
	PlainContentType = Plain + charsetSuffix
	HTMLContentType  = HTML + charsetSuffix
)
