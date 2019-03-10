package session

import (
	"errors"
	"github.com/go-redis/redis"
)

var (
	ErrNoValue = errors.New("session: no value")
)

func normalizeErr(err error) error {
	if err == redis.Nil {
		return ErrNoValue
	}

	return err
}
