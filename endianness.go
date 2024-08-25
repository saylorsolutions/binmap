package bin

import (
	"encoding/binary"
	"io"
)

// EndianIndicator is a function that indicates what byte order should be used with [MapEndian].
// Such a function will be called at the time of reading or writing, so it can make decisions as a result of previous reads.
type EndianIndicator func() binary.ByteOrder

// MapEndian will return a [Mapper] that overrides the endian value used with the result of the given [EndianIndicator].
func MapEndian(indicator EndianIndicator, wrapped Mapper) Mapper {
	if indicator == nil {
		panic("nil indicator")
	}
	if wrapped == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			endian = indicator()
			return wrapped.Read(r, endian)
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			endian = indicator()
			return wrapped.Write(w, endian)
		},
	}
}

// EndianInt interprets a previously read [AnyInt] typed field as an indicator of desired endianness.
//
//   - The bigVal value indicates that [binary.BigEndian] should be used.
//   - The littleVal value indicates that [binary.LittleEndian] should be used.
//   - If neither value matches the value of the field, then [binary.NativeEndian] is returned.
//
// The produced [EndianIndicator] will panic if the field pointer is nil.
func EndianInt[T AnyInt](field *T, bigVal T, littleVal T) EndianIndicator {
	if field == nil {
		panic("field is nil")
	}
	return func() binary.ByteOrder {
		switch *field {
		case bigVal:
			return binary.BigEndian
		case littleVal:
			return binary.LittleEndian
		default:
			return binary.NativeEndian
		}
	}
}

func OverrideBig() binary.ByteOrder {
	return binary.BigEndian
}

func OverrideLittle() binary.ByteOrder {
	return binary.LittleEndian
}

func OverrideNative() binary.ByteOrder {
	return binary.NativeEndian
}
