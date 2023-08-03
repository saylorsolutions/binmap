package bin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	ErrNilReadWrite = errors.New("nil read source or write target")
	ErrPanic        = errors.New("panic during Read or Write")
)

// ReadFunc is a function that reads data from a binary source.
type ReadFunc func(r io.Reader, endian binary.ByteOrder) error

// WriteFunc is a function that writes data to a binary target.
type WriteFunc func(w io.Writer, endian binary.ByteOrder) error

// Mapper is any procedure that knows how to read from and write to binary data, given an endianness policy.
type Mapper interface {
	// Read data from a binary source.
	Read(r io.Reader, endian binary.ByteOrder) error
	// Write data to a binary target.
	Write(w io.Writer, endian binary.ByteOrder) error
}

type mapper struct {
	read  ReadFunc
	write WriteFunc
}

func (m *mapper) Read(r io.Reader, endian binary.ByteOrder) error {
	if m.read != nil {
		return m.read(r, endian)
	}
	return errors.New("unimplemented read")
}

func (m *mapper) Write(w io.Writer, endian binary.ByteOrder) error {
	if m.write != nil {
		return m.write(w, endian)
	}
	return errors.New("unimplemented write")
}

var nilMapping = &mapper{
	read: func(r io.Reader, endian binary.ByteOrder) error {
		return ErrNilReadWrite
	},
	write: func(w io.Writer, endian binary.ByteOrder) error {
		return ErrNilReadWrite
	},
}

// MapSequence creates a Mapper that uses each given Mapper in order.
func MapSequence(mappings ...Mapper) Mapper {
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			for _, m := range mappings {
				if err := m.Read(r, endian); err != nil {
					return err
				}
			}
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			for _, m := range mappings {
				if err := m.Write(w, endian); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// Any is provided to make it easy to create a custom Mapper for any given type.
func Any(read ReadFunc, write WriteFunc) Mapper {
	return &mapper{
		read:  read,
		write: write,
	}
}

// OverrideEndian will override the endian settings for a single operation.
// This is useful for UTF-16 strings which are often read/written little-endian.
func OverrideEndian(m Mapper, endian binary.ByteOrder) Mapper {
	return Any(
		func(r io.Reader, _ binary.ByteOrder) error {
			return m.Read(r, endian)
		},
		func(w io.Writer, _ binary.ByteOrder) error {
			return m.Write(w, endian)
		},
	)
}

type BeforeReadHandler = func() error
type AfterReadHandler = func(err error) error
type BeforeWriteHandler = func() error
type AfterWriteHandler = func(err error) error

var _ Mapper = (*EventHandler)(nil)

type EventHandler struct {
	mapper      Mapper
	BeforeRead  BeforeReadHandler
	AfterRead   AfterReadHandler
	BeforeWrite BeforeWriteHandler
	AfterWrite  AfterWriteHandler
}

func NewEventHandler(mapper Mapper, handler EventHandler) Mapper {
	if mapper == nil {
		return nilMapping
	}
	handler.mapper = mapper
	return &handler
}

func (h *EventHandler) Read(r io.Reader, endian binary.ByteOrder) (err error) {
	if h.BeforeRead != nil {
		if err = h.BeforeRead(); err != nil {
			return err
		}
	}
	if h.AfterRead != nil {
		defer func() {
			rerr := h.AfterRead(err)
			if err != nil {
				return
			}
			err = rerr
		}()
	}
	return h.mapper.Read(r, endian)
}

func (h *EventHandler) Write(w io.Writer, endian binary.ByteOrder) (err error) {
	if h.BeforeWrite != nil {
		if err = h.BeforeWrite(); err != nil {
			return err
		}
	}
	if h.AfterWrite != nil {
		defer func() {
			rerr := h.AfterWrite(err)
			if err != nil {
				return
			}
			err = rerr
		}()
	}
	return h.mapper.Write(w, endian)
}

// ValidateRead will run the validator function after reading with the mapper.
func ValidateRead(mapper Mapper, validator AfterReadHandler) Mapper {
	return NewEventHandler(mapper, EventHandler{
		AfterRead: validator,
	})
}

// NormalizeWrite will run the normalizer before writing with the mapper.
func NormalizeWrite(mapper Mapper, normalizer BeforeWriteHandler) Mapper {
	return NewEventHandler(mapper, EventHandler{
		BeforeWrite: normalizer,
	})
}

// Lock will manage locking and unlocking a sync.Mutex before/after a read/write.
func Lock(mapper Mapper, mux *sync.Mutex) Mapper {
	return NewEventHandler(mapper, EventHandler{
		BeforeRead: func() error {
			mux.Lock()
			return nil
		},
		AfterRead: func(err error) error {
			mux.Unlock()
			return err
		},
		BeforeWrite: func() error {
			mux.Lock()
			return nil
		},
		AfterWrite: func(err error) error {
			mux.Unlock()
			return err
		},
	})
}

// RWLock will manage locking and unlocking a sync.RWMutex before/after a read/write.
// Writing the mapper only requires read locking, while reading with the mapper requires write locking since state is being mutated.
func RWLock(mapper Mapper, mux *sync.RWMutex) Mapper {
	return NewEventHandler(mapper, EventHandler{
		BeforeRead: func() error {
			mux.Lock()
			return nil
		},
		AfterRead: func(err error) error {
			mux.Unlock()
			return err
		},
		BeforeWrite: func() error {
			mux.RLock()
			return nil
		},
		AfterWrite: func(err error) error {
			mux.RUnlock()
			return err
		},
	})
}

func _panicHandler(r any, panicHandler func(any) error) error {
	var rerr error
	if panicHandler != nil {
		rerr = panicHandler(r)
	}
	if rerr == nil {
		return ErrPanic
	}
	return fmt.Errorf("%w: %v", ErrPanic, rerr)
}

// OnPanic will recover a panic from a Read or Write operation, and return the error returned from panicHandler wrapped with ErrPanic.
// If no error is returned from panicHandler, then a plain ErrPanic error will be returned.
func OnPanic(mapper Mapper, panicHandler func(any) error) Mapper {
	return Any(
		func(r io.Reader, endian binary.ByteOrder) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = _panicHandler(r, panicHandler)
				}
			}()
			return mapper.Read(r, endian)
		},
		func(w io.Writer, endian binary.ByteOrder) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = _panicHandler(r, panicHandler)
				}
			}()
			return mapper.Write(w, endian)
		},
	)
}
