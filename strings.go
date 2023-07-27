package bin

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

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
