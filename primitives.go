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

type AnyComplex interface {
	complex64 | complex128
}

func Complex[T AnyComplex](target *T) Mapper {
	return Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, target)
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, target)
		},
	)
}

var _ io.ByteReader = (*unbufferedByteReader)(nil)

type unbufferedByteReader struct {
	reader io.Reader
	buf    []byte
}

func (u *unbufferedByteReader) ReadByte() (byte, error) {
	if len(u.buf) == 0 {
		u.buf = make([]byte, 1)
	}
	_, err := u.reader.Read(u.buf)
	if err != nil {
		return 0, err
	}
	return u.buf[0], nil
}

// Varint encodes 16, 32, or 64-bit signed integers as a variable length integer.
// This is generally more efficient than reading/writing the full byte length.
func Varint(target *int64) Mapper {
	if target == nil {
		return nilMapping
	}
	return Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			ubr := &unbufferedByteReader{reader: r}
			val, err := binary.ReadVarint(ubr)
			if err != nil {
				return err
			}
			*target = val
			return nil
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutVarint(buf, *target)
			return binary.Write(w, endian, buf[:n])
		},
	)
}

// Uvarint encodes 16, 32, or 64-bit unsigned integers as a variable length integer.
// This is generally more efficient than reading/writing the full byte length.
func Uvarint(target *uint64) Mapper {
	if target == nil {
		return nilMapping
	}
	return Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			ubr := &unbufferedByteReader{reader: r}
			val, err := binary.ReadUvarint(ubr)
			if err != nil {
				return err
			}
			*target = val
			return nil
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, *target)
			return binary.Write(w, endian, buf[:n])
		},
	)
}
