package bin

import (
	"encoding/binary"
	"io"
)

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
