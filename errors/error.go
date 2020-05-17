package errors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"net/http"
	"reflect"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	s := http.StatusText(e.Code)
	if s != "" {
		return s
	}
	return fmt.Sprintf("error: %d", e.Code)
}

func Format(code int, format string, a ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, a...),
	}
}

var rawErrType = reflect.TypeOf(errors.New(""))

func GetStatus(err error) int {
	if reflect.TypeOf(err) == rawErrType {
		return 0
	}

	if err == NotExist || err == sql.ErrNoRows {
		return http.StatusNotFound
	}

	if coder, ok := err.(interface{ Code() int }); ok {
		return coder.Code()
	}

	if coder, ok := err.(interface{ Status() int }); ok {
		return coder.Status()
	}

	if coder, ok := err.(interface{ StatusCode() int }); ok {
		return coder.StatusCode()
	}

	if v := reflect.ValueOf(err); v.Kind() == reflect.Int {
		n := int(v.Int())
		if n > 0 {
			return n
		}
		return 0
	}

	keys := []string{"status", "Status", "status_code", "StatusCode", "statusCode", "code", "Code"}
	i := indirect(err)
	k := reflect.ValueOf(i).Kind()
	if k != reflect.Struct && k != reflect.Map {
		return 0
	}

	b, jErr := json.Marshal(i)
	if jErr != nil {
		logger.Errorf("Marshal: %v", err)
		return 0
	}
	var m types.M
	jErr = json.Unmarshal(b, &m)
	if jErr != nil {
		logger.Errorf("Unmarshal: %v", err)
		return 0
	}

	for _, k := range keys {
		s := m.Int(k)
		if s > 0 {
			return s
		}
	}
	return 0
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
