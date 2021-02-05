package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/gopub/wine/session"
)

type Session struct {
	id string
	c  *redis.Client
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Set(ctx context.Context, name string, value interface{}) error {
	return s.c.WithContext(ctx).HSet(s.id, name, value).Err()
}

func (s *Session) Get(ctx context.Context, name string, ptrValue interface{}) error {
	return s.c.WithContext(ctx).HGet(s.id, name).Scan(ptrValue)
}

func (s *Session) Delete(ctx context.Context, name string) error {
	return s.c.WithContext(ctx).HDel(s.id, name).Err()
}

func (s *Session) Clear() error {
	return s.c.Del(s.id).Err()
}

func (s *Session) SetTTL(ttl time.Duration) error {
	return s.c.Expire(s.id, ttl).Err()
}

var _ session.Session = (*Session)(nil)
