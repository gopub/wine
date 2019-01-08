package sessions

import (
	"github.com/gopub/log"
	"time"
)

type Store interface {
	Get(sid, key string, ptrValue interface{}) error
	Set(sid, key string, value interface{}) error
	Delete(sid string) error
	SetExpiration(sid string, expiration time.Duration) error
}

var defaultStore Store

func SetDefaultStore(s Store) {
	log.Debugf("Set default session store: %v", s)
	defaultStore = s
}

func DefaultStore() Store {
	return defaultStore
}
