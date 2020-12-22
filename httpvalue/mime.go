package httpvalue

const (
	Plain    = "text/plain"
	HTML     = "text/html"
	XML2     = "text/xml"
	CSS      = "text/css"
	XML      = "application/xml"
	XHTML    = "application/xhtml+xml"
	Protobuf = "application/x-protobuf"

	FormData = "multipart/form-data"
	GIF      = "image/gif"
	JPEG     = "image/jpeg"
	PNG      = "image/png"
	WEBP     = "image/webp"
	ICON     = "image/x-icon"

	MPEG = "video/mpeg"

	FormURLEncoded = "application/x-www-form-urlencoded"
	OctetStream    = "application/octet-stream"
	JSON           = "application/json"
	PDF            = "application/pdf"
	MSWord         = "application/msword"
	GZIP           = "application/x-gzip"
	WASM           = "application/wasm"
)

const (
	CharsetUTF8 = "charset=utf-8"

	charsetSuffix = "; " + CharsetUTF8

	PlainUTF8 = Plain + charsetSuffix

	// Hope this style is better than HTMLUTF8, etc.
	HtmlUTF8 = HTML + charsetSuffix
	JsonUTF8 = JSON + charsetSuffix
	XmlUTF8  = XML + charsetSuffix
)

func IsMIMETextType(typ string) bool {
	switch typ {
	case Plain, HTML, CSS, XML, XML2, XHTML, JSON, PlainUTF8, HtmlUTF8, JsonUTF8, XmlUTF8:
		return true
	default:
		return false
	}
}
