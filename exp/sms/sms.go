package sms

import "context"

type Sender interface {
	Send(ctx context.Context, recipient, content string) (string, error)
}
