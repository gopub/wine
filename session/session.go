package session

import (
	"context"
	"fmt"
	"time"

	"github.com/gopub/gox"
)

var defaultExpiration = time.Minute * 30
var minimumExpiration = time.Minute

const (
	keySession = "session"
	sidLength  = 40
)

type Session interface {
	ID() string
	Get(key string, ptrValue interface{}) error
	Set(key string, value interface{}) error
	Destroy() error
}

func GenerateSid() string {
	return gox.UniqueID()
}

func NewSession(id string) (Session, error) {
	return newSession(defaultStore, id, defaultExpiration)
}

func ContextWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, keySession, s)
}

func newSession(store Store, id string, expiration time.Duration) (Session, error) {
	logger := logger.With("id", id, "expiration", expiration)
	ok, err := store.Exists(id)
	if err != nil {
		return nil, fmt.Errorf("exists id=%s: %w", id, normalizeErr(err))
	}

	if ok {
		logger.Warn("Session already exists")
	}

	// Just save a key-val in order to create hmap in redis server
	if err := store.Set(id, "created_at", time.Now().Unix()); err != nil {
		return nil, fmt.Errorf("set id=%s: %w", id, normalizeErr(err))
	}

	logger.Info("New session")
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
	logger := logger.With("id", id, "expiration", expiration)
	ok, err := store.Exists(id)
	if err != nil {
		return nil, fmt.Errorf("exists id=%s: %w", id, normalizeErr(err))
	}

	if !ok {
		logger.Warn("Session doesn't exist")
		return nil, gox.ErrNotExist
	}

	logger.Info("Restored session")
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
		logger.Panic("Minimum expiration is 1 minute")
	}
	logger.Infof("Set default session expiration: %v", expiration)
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
		return fmt.Errorf("get key=%s: %w", key, normalizeErr(err))
	}

	if err := s.store.Expire(s.id, s.expiration); err != nil {
		return fmt.Errorf("set expiration id=%s: %w", s.id, normalizeErr(err))
	}
	return nil
}

func (s *session) Set(key string, value interface{}) error {
	if err := s.store.Set(s.id, key, value); err != nil {
		return fmt.Errorf("set key=%s: %w", key, err)
	}

	if err := s.store.Expire(s.id, s.expiration); err != nil {
		return fmt.Errorf("expire id=%s: %w", s.id, normalizeErr(err))
	}
	return nil
}

func (s *session) Destroy() error {
	err := normalizeErr(s.store.Delete(s.id))
	if err != nil {
		return fmt.Errorf("delete id=%s: %w", s.id, err)
	}
	logger.Infof("Destroyed session: id=%s", s.id)
	return err
}
