package sessions

import (
	"github.com/go-redis/redis"
	"time"
)

var _ Store = (*RedisStore)(nil)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string) *RedisStore {
	c := &RedisStore{}
	c.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return c
}

func NewRedisStoreWithClient(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

func (s *RedisStore) Get(sid, key string, ptrValue interface{}) error {
	cmd := s.client.HGet(sid, key)
	return cmd.Scan(ptrValue)
}

func (s *RedisStore) Set(sid, key string, value interface{}) error {
	return s.client.HSet(sid, key, value).Err()
}

func (s *RedisStore) Delete(sid string) error {
	return s.client.Del(sid).Err()
}

func (s *RedisStore) SetExpiration(sid string, expiration time.Duration) error {
	return s.client.Expire(sid, expiration).Err()
}
