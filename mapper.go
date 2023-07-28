package bin

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrNilReadWrite = errors.New("nil read source or write target")
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
	return errors.New("unimplemented")
}

func (m *mapper) Write(w io.Writer, endian binary.ByteOrder) error {
	if m.write != nil {
		return m.write(w, endian)
	}
	return errors.New("unimplemented")
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
