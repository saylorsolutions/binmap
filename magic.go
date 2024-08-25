package bin

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrMagicMismatch = errors.New("magic number mismatch")
)

func MagicNumber(num []byte) Mapper {
	if len(num) == 0 {
		return nilMapping
	}
	sz := uint64(len(num))
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			magic := make([]byte, len(num))
			if err := FixedBytes(&magic, sz).Read(r, endian); err != nil {
				return err
			}
			for i := 0; i < len(num); i++ {
				if magic[i] != num[i] {
					return ErrMagicMismatch
				}
			}
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			return FixedBytes(&num, sz).Write(w, endian)
		},
	}
}
