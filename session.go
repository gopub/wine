package wine

import "time"

type Session interface {
	Id() string
	Get(key string, ptrVal interface{}) error
	Set(key string, val interface{}) error
	SetEx(key string, val interface{}, expires time.Duration) error
}
