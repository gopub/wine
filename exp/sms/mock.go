package sms

import (
	"context"

	"github.com/google/uuid"
	"github.com/gopub/log"
)

type Mock struct {
}

func (m *Mock) Send(ctx context.Context, recipient, content string) (string, error) {
	sid := "mock-" + uuid.New().String()
	log.FromContext(ctx).Debugf("Mocked %s %s %s", recipient, content, sid)
	return sid, nil
}
