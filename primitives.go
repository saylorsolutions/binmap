package bin

import (
	"encoding/binary"
	"io"
)

// Byte will map a single byte.
func Byte(b *byte) Mapper {
	if b == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, b)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, b)
		},
	}
}

// Bool will map a single boolean.
func Bool(b *bool) Mapper {
	if b == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, b)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, b)
		},
	}
}

type AnyInt interface {
	int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

// Int will map any integer, excluding int.
func Int[T AnyInt](i *T) Mapper {
	if i == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, i)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, i)
		},
	}
}

type AnyFloat interface {
	float32 | float64
}

// Float will map any floating point value.
func Float[T AnyFloat](f *T) Mapper {
	if f == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, f)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, f)
		},
	}
}
