package mem

import (
	"context"

	"github.com/gopub/errors"
	"github.com/gopub/wine/session"
	"github.com/patrickmn/go-cache"
)

type Provider struct {
	cache *cache.Cache
}

var _ session.Provider = (*Provider)(nil)

func NewProvider() *Provider {
	p := new(Provider)
	p.cache = cache.New(session.DefaultOptions().TTL, session.DefaultOptions().TTL*50)
	return p
}

func (m *Provider) Get(ctx context.Context, id string) (session.Session, error) {
	v, ok := m.cache.Get(id)
	if !ok {
		return nil, errors.NotExist
	}
	return v.(*Session), nil
}

func (m *Provider) Create(ctx context.Context, id string, options *session.Options) (session.Session, error) {
	s := &Session{
		id:          id,
		sharedCache: m.cache,
		options:     options,
	}
	m.cache.Set(id, s, options.TTL)
	return s, nil
}

func (m *Provider) Delete(ctx context.Context, id string) error {
	m.cache.Delete(id)
	return nil
}
