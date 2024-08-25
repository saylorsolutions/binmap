package bin

import (
	"encoding/binary"
	"io"
)

// Conditional allows specifying conditional logic for mapping.
//
// The condition function will be executed at the time of reading or writing, allowing the result of previous reads to be considered.
// If a falseMapper is not provided (or nil) and the condition returns false, then no mapping activity will take place.
// If a nil trueMapper is provided, then an error will be returned.
func Conditional(condition func() bool, trueMapper Mapper, falseMapper ...Mapper) Mapper {
	if trueMapper == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			if condition() {
				return trueMapper.Read(r, endian)
			} else if len(falseMapper) > 0 && falseMapper[0] != nil {
				return falseMapper[0].Read(r, endian)
			}
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			if condition() {
				return trueMapper.Write(w, endian)
			} else if len(falseMapper) > 0 && falseMapper[0] != nil {
				return falseMapper[0].Write(w, endian)
			}
			return nil
		},
	}
}
