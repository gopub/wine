package session

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gopub/errors"
	"github.com/patrickmn/go-cache"
)

type memorySession struct {
	id   string
	data sync.Map

	sharedCache *cache.Cache
}

func (m *memorySession) ID() string {
	return m.id
}

func (m *memorySession) Set(ctx context.Context, name string, value interface{}) error {
	m.data.Store(name, value)
	return nil
}

func (m *memorySession) Get(ctx context.Context, name string, ptrValue interface{}) error {
	v, ok := m.data.Load(name)
	if !ok {
		return fmt.Errorf("%s is missing", name)
	}

	if ptrValue == nil {
		return errors.New("ptrValue is nil")
	}

	pv := reflect.ValueOf(ptrValue)
	if pv.Kind() != reflect.Ptr {
		return errors.New("ptrValue is not pointer")
	}

	if pv.IsNil() {
		return errors.New("ptrValue is nil")
	}

	if pv.Type() != reflect.PtrTo(reflect.TypeOf(v)) {
		return fmt.Errorf("cannot assign %T to %T", v, ptrValue)
	}

	pv.Set(reflect.ValueOf(v))
	return nil
}

func (m *memorySession) Delete(ctx context.Context, name string) error {
	m.data.Delete(name)
	return nil
}

func (m *memorySession) Clear() error {
	m.data = sync.Map{}
	return nil
}

func (m *memorySession) Flush() error {
	return nil
}

var _ Session = (*memorySession)(nil)

type MemoryProvider struct {
	cache *cache.Cache
}

var _ Provider = (*MemoryProvider)(nil)

func NewMemoryProvider() *MemoryProvider {
	p := new(MemoryProvider)
	p.cache = cache.New(DefaultOptions().TTL, DefaultOptions().TTL*50)
	return p
}

func (m *MemoryProvider) Get(ctx context.Context, id string) (Session, error) {
	v, ok := m.cache.Get(id)
	if !ok {
		return nil, errors.NotExist
	}
	return v.(Session), nil
}

func (m *MemoryProvider) Create(ctx context.Context, id string, options *Options) (Session, error) {
	s := &memorySession{
		id:          id,
		sharedCache: m.cache,
	}
	m.cache.Set(id, s, options.TTL)
	return s, nil
}

func (m *MemoryProvider) Delete(ctx context.Context, id string) error {
	m.cache.Delete(id)
	return nil
}
