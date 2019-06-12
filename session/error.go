package session

import (
	"github.com/go-redis/redis"
)

const ErrNoValue = Error("session: no value")

type Error string

func (e Error) Error() string { return string(e) }

func normalizeErr(err error) error {
	if err == redis.Nil {
		return ErrNoValue
	}

	return err
}
