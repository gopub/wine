package awssns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gopub/log/v2"
	"github.com/gopub/wine/exp/sms"
)

type SMS struct {
	client *sns.SNS
}

var _ sms.Sender = (*SMS)(nil)

func NewSMS() *SMS {
	sess := session.Must(session.NewSession())
	s := new(SMS)
	s.client = sns.New(sess)
	return s
}

func (s *SMS) Send(ctx context.Context, recipient, content string) (string, error) {
	logger := log.FromContext(ctx).With("recipient", recipient)
	input := &sns.PublishInput{
		Message:     aws.String(content),
		PhoneNumber: aws.String(recipient),
	}
	output, err := s.client.Publish(input)
	if err != nil {
		return "", fmt.Errorf("publish: %w", err)
	}
	logger.Debugf("%v %v", recipient, content)
	return *output.MessageId, nil
}
