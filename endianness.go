package bin

import (
	"encoding/binary"
	"io"
)

// ByteOrder is an alias of binary.ByteOrder provided for reference consistency.
type ByteOrder = binary.ByteOrder

var (
	BigEndian    ByteOrder = binary.BigEndian    // BigEndian provided for reference consistency.
	LittleEndian ByteOrder = binary.LittleEndian // LittleEndian provided for reference consistency.
	NativeEndian ByteOrder = binary.NativeEndian // NativeEndian provided for reference consistency.
)

// EndianIndicator is a function that indicates what byte order should be used with [MapEndian].
// Such a function will be called at the time of reading or writing, so it can make decisions as a result of previous reads.
type EndianIndicator func() ByteOrder

// MapEndian will return a [Mapper] that overrides the endian value used with the result of the given [EndianIndicator].
func MapEndian(indicator EndianIndicator, wrapped Mapper) Mapper {
	if indicator == nil {
		panic("nil indicator")
	}
	if wrapped == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian ByteOrder) error {
			endian = indicator()
			return wrapped.Read(r, endian)
		},
		write: func(w io.Writer, endian ByteOrder) error {
			endian = indicator()
			return wrapped.Write(w, endian)
		},
	}
}

// EndianInt interprets a previously read [AnyInt] typed field as an indicator of desired endianness.
//
//   - The bigVal value indicates that [bin.BigEndian] should be used.
//   - The littleVal value indicates that [bin.LittleEndian] should be used.
//   - If neither value matches the value of the field, then [bin.NativeEndian] is returned.
//
// The produced [EndianIndicator] will panic if the field pointer is nil.
func EndianInt[T AnyInt](field *T, bigVal T, littleVal T) EndianIndicator {
	if field == nil {
		panic("field is nil")
	}
	return func() ByteOrder {
		switch *field {
		case bigVal:
			return BigEndian
		case littleVal:
			return LittleEndian
		default:
			return NativeEndian
		}
	}
}

func OverrideBig() ByteOrder {
	return BigEndian
}

func OverrideLittle() ByteOrder {
	return LittleEndian
}

func OverrideNative() ByteOrder {
	return NativeEndian
}
