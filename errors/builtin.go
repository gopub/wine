package errors

import (
	"errors"
	"fmt"
)

func New(text string) error {
	return errors.New(text)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Cause(err error) error {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return err
}

func Wrapf(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	a = append(a, err)
	return fmt.Errorf(format+":%w", a...)
}
