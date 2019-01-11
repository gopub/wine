package wine

import (
	"net/http"
	"reflect"
)

func GetHTTP2Conn(w http.ResponseWriter) interface{} {
	if reflect.TypeOf(w).String() != "*http.http2responseWriter" {
		return nil
	}
	http2responseWriter := reflect.ValueOf(w).Elem()
	http2responseWriterState := http2responseWriter.FieldByName("rws").Elem()
	conn := http2responseWriterState.FieldByName("conn").Elem().FieldByName("conn")
	return conn
}
