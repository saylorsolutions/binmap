package bin

import (
	"bufio"
	"bytes"
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

// Any is provided to make it easy to create a custom Mapper for any given type.
func Any[T any](target *T, read ReadFunc, write WriteFunc) Mapper {
	if target == nil {
		return nilMapping
	}
	return &mapper{
		read:  read,
		write: write,
	}
}

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

// FixedString will map a string with a max length that is known ahead of time.
// The target string will not contain any zero bytes if the encoded string is less than the space allowed.
func FixedString(s *string, length int) Mapper {
	if s == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			buf := make([]byte, length)
			if err := binary.Read(r, endian, buf); err != nil {
				return err
			}
			buf = bytes.TrimRightFunc(buf, func(r rune) bool {
				return r == 0
			})
			*s = string(buf)
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			bs := make([]byte, length)
			copy(bs, *s)
			return binary.Write(w, endian, bs)
		},
	}
}

// NullTermString will read and write null-byte terminated string.
// The string provided doesn't have to contain a null terminator, since one will be added on write.
func NullTermString(s *string) Mapper {
	if s == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			var (
				buf bytes.Buffer
				br  = bufio.NewReader(r)
			)
			for {
				b, err := br.ReadByte()
				if err != nil {
					return err
				}
				if b == 0 {
					*s = buf.String()
					return nil
				}
				if err := buf.WriteByte(b); err != nil {
					return err
				}
			}
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			bs := []byte(*s)
			if err := binary.Write(w, endian, bs); err != nil {
				return err
			}
			var null byte
			return binary.Write(w, endian, &null)
		},
	}
}

type SizeType interface {
	uint8 | uint16 | uint32 | uint64
}

// Size maps any value that can reasonably be used to express a size.
func Size[S SizeType](size *S) Mapper {
	if size == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			return binary.Read(r, endian, size)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return binary.Write(w, endian, size)
		},
	}
}

// FixedBytes maps a byte slice of a known length.
func FixedBytes[S SizeType](buf *[]byte, length S) Mapper {
	if buf == nil {
		return nilMapping
	}
	sz := uint64(length)
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			_buf := make([]byte, sz)
			if err := binary.Read(r, endian, _buf); err != nil {
				return err
			}
			*buf = _buf
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			out := make([]byte, sz)
			copy(out, *buf)
			return binary.Write(w, endian, out)
		},
	}
}

// LenBytes is used for situations where an arbitrarily sized byte slice is encoded after its length.
// This mapper will read the length, and then length number of bytes into a byte slice.
// The mapper will write the length and bytes in the same order.
func LenBytes[S SizeType](buf *[]byte, length *S) Mapper {
	if buf == nil {
		return nilMapping
	}
	if length == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			if err := Size(length).Read(r, endian); err != nil {
				return err
			}
			return FixedBytes(buf, *length).Read(r, endian)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			if err := Size(length).Write(w, endian); err != nil {
				return err
			}
			return FixedBytes(buf, *length).Write(w, endian)
		},
	}
}

// Slice will produce a mapper informed from the given function to use a slice of values.
// The slice length must be known ahead of time.
// The mapVal function will be used to create a Mapper that relates to the type returned from allocNext.
// The returned Mapper will orchestrate the array construction according to the given function.
func Slice[E any, S SizeType](target *[]E, count S, mapVal func(*E) Mapper) Mapper {
	if target == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			input := make([]E, count)
			i := S(0)
			for i < count {
				var e E
				m := mapVal(&e)
				if err := m.Read(r, endian); err != nil {
					return err
				}
				input[i] = e
				i++
			}
			*target = input
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			output := make([]E, count)
			copy(output, *target)
			for _, out := range output {
				if err := mapVal(&out).Write(w, endian); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// LenSlice is for situations where a slice is encoded with its length prepended.
// Otherwise, this behaves exactly like Slice.
func LenSlice[E any, S SizeType](target *[]E, count *S, mapVal func(*E) Mapper) Mapper {
	if target == nil {
		return nilMapping
	}
	if count == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			if err := Size(count).Read(r, endian); err != nil {
				return err
			}
			return Slice(target, *count, mapVal).Read(r, endian)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			if err := Size(count).Write(w, endian); err != nil {
				return err
			}
			return Slice(target, *count, mapVal).Write(w, endian)
		},
	}
}

// DynamicSlice tries to accomplish a happy medium between LenSlice and Slice.
// A uint32 will be used to store the size of the given slice, but it's not necessary to read this from a field, rather it will be discovered at write time.
// This means that the size will be available at read time by first reading the uint32 with LenSlice, without requiring a caller provided field.
// In a scenario where a slice in a struct is used, this makes it easier to read and write because the struct doesn't need to store the size in a field.
func DynamicSlice[E any](target *[]E, mapVal func(*E) Mapper) Mapper {
	if target == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			var length uint32
			return LenSlice(target, &length, mapVal).Read(r, endian)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			var length = uint32(len(*target))
			return LenSlice(target, &length, mapVal).Write(w, endian)
		},
	}
}
