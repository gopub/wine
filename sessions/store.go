package sessions

import (
	"time"
)

type Store interface {
	Get(sid, key string, ptrValue interface{}) error
	Set(sid, key string, value interface{}) error
	DeleteKey(sid, key string) error
	ExistsKey(sid, key string) (bool, error)
	Exists(sid string) (bool, error)
	Delete(sid string) error
	Expire(sid string, expiration time.Duration) error
}

var defaultStore Store

func SetDefaultStore(s Store) {
	logger.Debugf("Set default session store: %v", s)
	defaultStore = s
}

func DefaultStore() Store {
	return defaultStore
}
