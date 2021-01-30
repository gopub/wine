package session

import (
	"context"
	"strings"
	"time"

	"github.com/gopub/environ"
)

type Options struct {
	Name           string
	TTL            time.Duration
	CookiePath     string
	CookieHttpOnly bool

	keyForID       string
	headerKeyForID string
}

var defaultOptions *Options

func DefaultOptions() *Options {
	if defaultOptions != nil {
		return defaultOptions
	}
	o := &Options{
		Name:           environ.String("wine.session.name", "wsession"),
		TTL:            environ.Duration("wine.session.ttl", 30*time.Minute),
		CookieHttpOnly: true,
		CookiePath:     "/",
	}
	o.Name = strings.ToLower(strings.TrimSpace(o.Name))
	if o.Name == "" {
		panic("Session name cannot be empty")
	}
	if o.TTL < time.Minute {
		panic("Session TTL cannot be less than 1 min")
	}
	o.keyForID = o.Name + "id"
	o.headerKeyForID = "X-" + strings.ToUpper(o.Name[0:1]) + o.Name[1:]

	defaultOptions = o
	return defaultOptions
}

type Session interface {
	ID() string
	Set(ctx context.Context, name string, value interface{}) error
	Get(ctx context.Context, name string, ptrValue interface{}) error
	Delete(ctx context.Context, name string) error
	Clear() error
	Flush() error
	Options() *Options
}

type contextKey int

const (
	keySession contextKey = iota + 1
)

func Get(ctx context.Context) Session {
	v, _ := ctx.Value(keySession).(Session)
	return v
}

func withSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, keySession, s)
}

type Provider interface {
	Get(ctx context.Context, id string) (Session, error)
	Create(ctx context.Context, id string, options *Options) (Session, error)
	Delete(ctx context.Context, id string) error
}
