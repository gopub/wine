package session

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gopub/gox"
	"reflect"
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

	pv := reflect.ValueOf(ptrValue)
	if pv.Kind() != reflect.Ptr {
		panic("ptrValue must be a pointer")
	}

	if b, ok := ptrValue.(*[]byte); ok {
		data, err := cmd.Bytes()
		if err != nil {
			return err
		}
		*b = data
		return nil
	}

	switch v := pv.Elem(); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := cmd.Int64()
		if err != nil {
			return err
		}
		v.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := cmd.Uint64()
		if err != nil {
			return err
		}
		v.SetUint(i)
		return nil
	case reflect.Float32, reflect.Float64:
		i, err := cmd.Float64()
		if err != nil {
			return err
		}
		v.SetFloat(i)
		return nil
	case reflect.String:
		i, err := cmd.Result()
		if err != nil {
			return err
		}
		v.SetString(i)
		return nil
	default:
		data, _ := cmd.Bytes()
		return gox.JSONUnmarshal(data, ptrValue)
	}
}

func (s *RedisStore) Set(sid, key string, value interface{}) error {
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.String:
		err := s.client.HSet(sid, key, fmt.Sprint(value)).Err()
		if err != nil {
			return err
		}
		return nil
	}

	if _, ok := value.([]byte); ok {
		err := s.client.HSet(sid, key, value).Err()
		if err != nil {
			return err
		}
		return nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = s.client.HSet(sid, key, b).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisStore) Exists(sid string) (bool, error) {
	n, err := s.client.Exists(sid).Result()
	return n > 0, err
}

func (s *RedisStore) Delete(sid string) error {
	return s.client.Del(sid).Err()
}

func (s *RedisStore) ExistsKey(sid, key string) (bool, error) {
	cmd := s.client.HExists(sid, key)
	return cmd.Result()
}

func (s *RedisStore) DeleteKey(sid, key string) error {
	return s.client.HDel(sid, key).Err()
}

func (s *RedisStore) Expire(sid string, expiration time.Duration) error {
	return s.client.Expire(sid, expiration).Err()
}
