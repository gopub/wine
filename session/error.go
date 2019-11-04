package session

import (
	"github.com/go-redis/redis"
	"github.com/gopub/gox"
)

type Error string

func (e Error) Error() string { return string(e) }

func normalizeErr(err error) error {
	if err == redis.Nil {
		return gox.ErrNotExist
	}
	return err
}
