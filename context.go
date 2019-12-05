package wine

// ContextKey defines the key type of context
type ContextKey int

// Context keys
const (
	CKBasicAuthUser ContextKey = iota + 1
	CKHTTPResponseWriter
	CKTemplates
)
