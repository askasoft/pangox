package xerrs

import (
	"context"
	"errors"

	"github.com/askasoft/pango/tbs"
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

func AsClientError(err error) (ce *ClientError, ok bool) {
	ok = errors.As(err, &ce)
	return
}

func IsClientError(err error) bool {
	_, ok := AsClientError(err)
	return ok
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

func AsFailedError(err error) (fe *FailedError, ok bool) {
	ok = errors.As(err, &fe)
	return
}

func IsFailedError(err error) bool {
	_, ok := AsFailedError(err)
	return ok
}

func (fe *FailedError) Error() string {
	return fe.Err.Error()
}

func (fe *FailedError) Unwrap() error {
	return fe.Err
}

type SkippedError struct {
	Err error
}

func NewSkippedError(err error) error {
	return &SkippedError{Err: err}
}

func AsSkippedError(err error) (se *SkippedError, ok bool) {
	ok = errors.As(err, &se)
	return
}

func IsSkippedError(err error) bool {
	_, ok := AsSkippedError(err)
	return ok
}

func (fe *SkippedError) Error() string {
	return fe.Err.Error()
}

func (fe *SkippedError) Unwrap() error {
	return fe.Err
}

type LocaleError struct {
	name string
	vars []any
}

func NewLocaleError(name string, vars ...any) *LocaleError {
	return &LocaleError{name, vars}
}

func AsLocaleError(err error) (le *LocaleError, ok bool) {
	ok = errors.As(err, &le)
	return
}

func IsLocaleError(err error) bool {
	_, ok := AsLocaleError(err)
	return ok
}

func (le *LocaleError) Error() string {
	return le.LocaleError("")
}

func (le *LocaleError) LocaleError(loc string) string {
	err := tbs.Format(loc, le.name, le.vars...)
	if err == "" {
		err = le.name
	}
	return err
}
