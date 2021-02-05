package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/gopub/errors"
	"github.com/gopub/wine/session"
)

type Provider struct {
	c *redis.Client
}

var _ session.Provider = (*Provider)(nil)

func NewProvider(c *redis.Client) *Provider {
	p := new(Provider)
	p.c = c
	return p
}

func (p *Provider) Get(ctx context.Context, id string) (session.Session, error) {
	v, err := p.c.WithContext(ctx).Exists(id).Result()
	if err != nil {
		return nil, err
	}

	if v == 0 {
		return nil, errors.NotExist
	}

	return &Session{
		id: id,
		c:  p.c,
	}, nil
}

func (p *Provider) Create(ctx context.Context, id string, ttl time.Duration) (session.Session, error) {
	err := p.c.WithContext(ctx).HSet(id, "@", 0).Err()
	if err != nil {
		return nil, err
	}
	s := &Session{
		id: id,
		c:  p.c,
	}
	if err = p.c.Expire(id, ttl).Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func (p *Provider) Delete(ctx context.Context, id string) error {
	return p.c.WithContext(ctx).Del(id).Err()
}
