package cli

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	testSettings()
}

func TestUsageError_Is(t *testing.T) {
	err := NewUsageError("test")
	assert.ErrorIs(t, err, &UsageError{})

	var ErrTesting = errors.New("test")
	err2 := NewUsageError("%w", ErrTesting)
	assert.ErrorIs(t, err2, &UsageError{})
	assert.ErrorIs(t, err2, ErrTesting)
}

func TestUsageError_Unwrap(t *testing.T) {
	var ErrTesting = errors.New("test")
	err := NewUsageError("%w", ErrTesting)
	var targetUsage = new(UsageError)
	assert.True(t, errors.As(err, &targetUsage))
}

func TestUsageError_Error(t *testing.T) {
	err := &UsageError{}
	assert.Equal(t, "usage error", err.Error(), "Default error output should be returned when there is no wrapping error")
	err2 := NewUsageError("test")
	assert.Equal(t, "usage error: test", err2.Error(), "The wrapped error's output should be returned when Error is called")
}

func ExampleNewUsageError() {
	testSettings()
	tlc := TopLevelCommandWithName("parent")
	cmd := tlc.AddCommand("another", "Another command!")
	cmd.AddUsageExample("[FLAGS]")
	cmd.Does(func(flags *Flags, out *Printer) error {
		return NewUsageError("test usage error")
	})
	// Error not handled for brevity
	_ = tlc.Exec([]string{"another"})

	// Output:
	// usage error: test usage error
	// Another command!
	//
	// USAGE: parent another [FLAGS]
	//
	// FLAGS:
	//   -h, --help   Prints this usage information
}
