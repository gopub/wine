package sessions

import (
	"context"
	"errors"
	"github.com/gopub/log"
	"time"
)

var defaultExpiration = time.Minute * 30
var minimumExpiration = time.Minute

const keySession = "session"

type Session interface {
	ID() string
	Get(key string, ptrValue interface{}) error
	Set(key string, value interface{}) error
	Destroy() error
}

func NewSession(id string) (Session, error) {
	return newSession(defaultStore, id, defaultExpiration)
}

func ContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, keySession, s)
}

func newSession(store Store, id string, expiration time.Duration) (Session, error) {
	logger := log.With("id", id, "expiration", expiration)
	b, err := store.Exists(id)

	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if b {
		err = errors.New("session already exists")
		logger.Error(err)
		return nil, err
	}

	logger.Info("New session:", id)
	return &session{
		id:         id,
		store:      store,
		expiration: expiration,
	}, nil
}

func RestoreSession(id string) (Session, error) {
	return restoreSession(defaultStore, id, defaultExpiration)
}

func restoreSession(store Store, id string, expiration time.Duration) (Session, error) {
	logger := log.With("id", id, "expiration", expiration)
	if b, err := store.Exists(id); err != nil {
		logger.Error(err)
		return nil, err
	} else if !b {
		err := errors.New("session doesn't exist")
		logger.Error(err)
		return nil, err
	}

	return &session{
		id:         id,
		store:      store,
		expiration: expiration,
	}, nil
}

func GetSession(ctx context.Context) Session {
	s, _ := ctx.Value(keySession).(Session)
	return s
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

type session struct {
	id         string
	store      Store
	expiration time.Duration
}

func (s *session) ID() string {
	if s == nil {
		return ""
	}
	return s.id
}

func (s *session) Get(key string, ptrValue interface{}) error {
	if err := s.store.Get(s.id, key, ptrValue); err != nil {
		return err
	}
	return s.store.Expire(s.id, s.expiration)
}

func (s *session) Set(key string, value interface{}) error {
	if err := s.store.Set(s.id, key, value); err != nil {
		return err
	}
	return s.store.Expire(s.id, s.expiration)
}

func (s *session) Destroy() error {
	log.Debugf("Destroyed session:%s", s.id)
	return s.store.Delete(s.id)
}
