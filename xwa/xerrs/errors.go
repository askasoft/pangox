package xerrs

import (
	"context"
	"errors"
)

func ContextCause(ctx context.Context, errs ...error) error {
	for _, err := range errs {
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	if err := context.Cause(ctx); err != nil {
		return err
	}
	return ctx.Err()
}

type ClientError struct {
	Err error
}

func NewClientError(err error) error {
	return &ClientError{Err: err}
}

func IsClientError(err error) bool {
	var ce *ClientError
	return errors.As(err, &ce)
}

func (ce *ClientError) Error() string {
	return ce.Err.Error()
}

func (ce *ClientError) Unwrap() error {
	return ce.Err
}

type FailedError struct {
	Err error
}

func NewFailedError(err error) error {
	return &FailedError{Err: err}
}

func IsFailedError(err error) bool {
	var fe *FailedError
	return errors.As(err, &fe)
}

func (fe *FailedError) Error() string {
	return fe.Err.Error()
}

func (fe *FailedError) Unwrap() error {
	return fe.Err
}
