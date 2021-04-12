package twilio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gopub/environ"
	"github.com/gopub/errors"
	"github.com/gopub/types"
	"github.com/gopub/wine"
)

type SMSConfig struct {
	PhoneNumbers []string
	Account      string
	Token        string
}

func (c *SMSConfig) Validate() error {
	if len(c.PhoneNumbers) == 0 {
		return errors.New("missing phone numbers")
	}

	for i, pn := range c.PhoneNumbers {
		p, err := types.NewPhoneNumber(pn)
		if err != nil {
			return fmt.Errorf("invalid phone number %s: %w", pn, err)
		}
		c.PhoneNumbers[i] = p.String()
	}

	if c.Account == "" {
		return errors.New("missing account")
	}

	if c.Token == "" {
		return errors.New("missing token")
	}

	return nil
}

type sendResult struct {
	Sid          string `json:"sid"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Status       string `json:"status"`
}

type SMS struct {
	config   *SMSConfig
	numIndex int
	send     *wine.SMSEndpoint
}

func NewSMS(config *SMSConfig) (*SMS, error) {
	if config == nil {
		config = new(SMSConfig)
		config.PhoneNumbers = environ.MustStringSlice("twilio.numbers")
		config.Account = environ.MustString("twilio.account")
		config.Token = environ.MustString("twilio.token")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	s := &SMS{
		config:   config,
		numIndex: 0,
	}

	sendURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.config.Account)
	var err error
	s.send, err = wine.DefaultSMS.Endpoint(http.MethodPost, sendURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create http endpoint %s: %w", sendURL, err)
	}
	s.send.SetBasicAuthorization(s.config.Account, s.config.Token)
	return s, nil
}

func (s *SMS) Send(ctx context.Context, recipient, content string) (string, error) {
	pn, err := types.NewPhoneNumber(recipient)
	if err != nil {
		return "", fmt.Errorf("invalid recipient: %w", err)
	}
	s.numIndex = (s.numIndex + 1) % len(s.config.PhoneNumbers)
	form := url.Values{}
	form.Add("To", pn.String())
	form.Add("From", s.config.PhoneNumbers[s.numIndex])
	form.Add("Body", content)
	var result sendResult
	err = s.send.Call(ctx, form, &result)
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	switch result.Status {
	case "queued", "sent", "delivered":
		return result.Sid, nil
	default:
		return result.Sid, errors.Format(result.ErrorCode, result.ErrorMessage)
	}
}
