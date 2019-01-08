package sessions

import (
	"context"
	"github.com/gopub/log"
	"time"
)

var defaultExpiration = time.Minute * 30
var minimumExpiration = time.Minute

const KeySid = "sid"

type Session interface {
	ID() string
	Get(key string, ptrValue interface{}) error
	Set(key string, value interface{}) error
	Destroy() error
}

func NewSession(store Store, id string, expiration time.Duration) Session {
	log.Debugf("New session:%s", id)
	return &session{
		id:         id,
		store:      store,
		expiration: expiration,
	}
}

func SetDefaultExpiration(expiration time.Duration) {
	if expiration < minimumExpiration {
		panic("Minimum expiration is 1 minute")
	}
	log.Debugf("Set default session expiration: %v", expiration)
	defaultExpiration = expiration
}

func DefaultExpiration() time.Duration {
	return defaultExpiration
}

func GetSession(sid string) Session {
	if defaultStore == nil {
		panic("DefaultStore is nil")
	}

	if len(sid) == 0 {
		panic("sid is empty")
	}
	return NewSession(defaultStore, sid, defaultExpiration)
}

func GetContextSession(ctx context.Context) Session {
	if sid, ok := ctx.Value(KeySid).(string); ok {
		return GetSession(sid)
	}
	return nil
}

type session struct {
	id         string
	store      Store
	expiration time.Duration
}

func (s *session) ID() string {
	return s.id
}

func (s *session) Get(key string, ptrValue interface{}) error {
	if err := s.store.Get(s.id, key, ptrValue); err != nil {
		return err
	}
	return s.store.SetExpiration(s.id, s.expiration)
}

func (s *session) Set(key string, value interface{}) error {
	if err := s.store.Set(s.id, key, value); err != nil {
		return err
	}
	return s.store.SetExpiration(s.id, s.expiration)
}

func (s *session) Destroy() error {
	log.Debugf("Destroyed session:%s", s.id)
	return s.store.Delete(s.id)
}
