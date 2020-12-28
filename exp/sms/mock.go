package sms

import (
	"context"

	"github.com/gopub/log"
	"github.com/gopub/wine"
)

type Mock struct {
}

func (m *Mock) Send(ctx context.Context, recipient, content string) (string, error) {
	sid := "mock-" + wine.NewUUID()
	log.FromContext(ctx).Debugf("Mocked %s %s %s", recipient, content, sid)
	return sid, nil
}
