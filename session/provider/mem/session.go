package mem

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gopub/errors"
	"github.com/gopub/wine/session"
	"github.com/patrickmn/go-cache"
)

type Session struct {
	id      string
	data    sync.Map
	options *session.Options

	sharedCache *cache.Cache
}

func (m *Session) ID() string {
	return m.id
}

func (m *Session) Set(ctx context.Context, name string, value interface{}) error {
	m.data.Store(name, value)
	return nil
}

func (m *Session) Get(ctx context.Context, name string, ptrValue interface{}) error {
	v, ok := m.data.Load(name)
	if !ok {
		return errors.NotExist
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

func (m *Session) Delete(ctx context.Context, name string) error {
	m.data.Delete(name)
	return nil
}

func (m *Session) Clear() error {
	m.data = sync.Map{}
	return nil
}

func (m *Session) Flush() error {
	return nil
}

func (m *Session) Options() *session.Options {
	return m.options
}

var _ session.Session = (*Session)(nil)
