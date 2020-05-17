package errors

import "net/http"

func BadRequest(format string, a ...interface{}) *Error {
	return Format(http.StatusBadRequest, format, a...)
}

func Unauthorized(format string, a ...interface{}) *Error {
	return Format(http.StatusUnauthorized, format, a...)
}

func PaymentRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusPaymentRequired, format, a...)
}

func Forbidden(format string, a ...interface{}) *Error {
	return Format(http.StatusPaymentRequired, format, a...)
}

func NotFound(format string, a ...interface{}) *Error {
	return Format(http.StatusNotFound, format, a...)
}

func MethodNotAllowed(format string, a ...interface{}) *Error {
	return Format(http.StatusMethodNotAllowed, format, a...)
}

func NotAcceptable(format string, a ...interface{}) *Error {
	return Format(http.StatusNotAcceptable, format, a...)
}

func ProxyAuthRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusProxyAuthRequired, format, a...)
}

func RequestTimeout(format string, a ...interface{}) *Error {
	return Format(http.StatusRequestTimeout, format, a...)
}

func Conflict(format string, a ...interface{}) *Error {
	return Format(http.StatusConflict, format, a...)
}

func LengthRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusLengthRequired, format, a...)
}

func PreconditionFailed(format string, a ...interface{}) *Error {
	return Format(http.StatusPreconditionFailed, format, a...)
}

func RequestEntityTooLarge(format string, a ...interface{}) *Error {
	return Format(http.StatusRequestEntityTooLarge, format, a...)
}

func RequestURITooLong(format string, a ...interface{}) *Error {
	return Format(http.StatusRequestURITooLong, format, a...)
}

func ExpectationFailed(format string, a ...interface{}) *Error {
	return Format(http.StatusExpectationFailed, format, a...)
}

func Teapot(format string, a ...interface{}) *Error {
	return Format(http.StatusTeapot, format, a...)
}

func MisdirectedRequest(format string, a ...interface{}) *Error {
	return Format(http.StatusMisdirectedRequest, format, a...)
}

func UnprocessableEntity(format string, a ...interface{}) *Error {
	return Format(http.StatusUnprocessableEntity, format, a...)
}

func Locked(format string, a ...interface{}) *Error {
	return Format(http.StatusLocked, format, a...)
}

func TooEarly(format string, a ...interface{}) *Error {
	return Format(http.StatusTooEarly, format, a...)
}

func UpgradeRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusUpgradeRequired, format, a...)
}

func PreconditionRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusPreconditionRequired, format, a...)
}

func TooManyRequests(format string, a ...interface{}) *Error {
	return Format(http.StatusTooManyRequests, format, a...)
}

func RequestHeaderFieldsTooLarge(format string, a ...interface{}) *Error {
	return Format(http.StatusRequestHeaderFieldsTooLarge, format, a...)
}

func UnavailableForLegalReasons(format string, a ...interface{}) *Error {
	return Format(http.StatusUnavailableForLegalReasons, format, a...)
}

func InternalServerError(format string, a ...interface{}) *Error {
	return Format(http.StatusInternalServerError, format, a...)
}

func NotImplemented(format string, a ...interface{}) *Error {
	return Format(http.StatusNotImplemented, format, a...)
}

func BadGateway(format string, a ...interface{}) *Error {
	return Format(http.StatusBadGateway, format, a...)
}

func ServiceUnavailable(format string, a ...interface{}) *Error {
	return Format(http.StatusServiceUnavailable, format, a...)
}

func GatewayTimeout(format string, a ...interface{}) *Error {
	return Format(http.StatusGatewayTimeout, format, a...)
}

func HTTPVersionNotSupported(format string, a ...interface{}) *Error {
	return Format(http.StatusHTTPVersionNotSupported, format, a...)
}

func VariantAlsoNegotiates(format string, a ...interface{}) *Error {
	return Format(http.StatusVariantAlsoNegotiates, format, a...)
}

func InsufficientStorage(format string, a ...interface{}) *Error {
	return Format(http.StatusInsufficientStorage, format, a...)
}

func LoopDetected(format string, a ...interface{}) *Error {
	return Format(http.StatusLoopDetected, format, a...)
}

func NotExtended(format string, a ...interface{}) *Error {
	return Format(http.StatusNotExtended, format, a...)
}

func NetworkAuthenticationRequired(format string, a ...interface{}) *Error {
	return Format(http.StatusNetworkAuthenticationRequired, format, a...)
}
