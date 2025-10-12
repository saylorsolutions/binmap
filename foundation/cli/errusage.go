package cli

import (
	"errors"
	"fmt"
)

// UsageError is a special purpose error used to signal that usage information should be shown to the user.
// This is intended to be used as an error response for [Command] validation.
type UsageError struct {
	wrapped error
}

func (e *UsageError) Error() string {
	if e.wrapped == nil {
		return "usage error"
	}
	return "usage error: " + e.wrapped.Error()
}

func (e *UsageError) Is(err error) bool {
	var _err *UsageError
	return errors.As(err, &_err)
}

func (e *UsageError) Unwrap() error {
	return e.wrapped
}

// NewUsageError is used to create a [UsageError].
// The format and args parameters are passed to [fmt.Errorf] to create the underlying error.
func NewUsageError(format string, args ...any) error {
	return &UsageError{wrapped: fmt.Errorf(format, args...)}
}
