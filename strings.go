package bin

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// FixedString will map a string with a max length that is known ahead of time.
// The target string will not contain any trailing zero bytes if the encoded string is less than the space allowed.
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
// The string should not contain a null terminator, one will be added on write.
func NullTermString(s *string) Mapper {
	if s == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			var (
				buf bytes.Buffer
				ubr = &unbufferedByteReader{reader: r}
			)
			for {
				b, err := ubr.ReadByte()
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
			bs := append([]byte(*s), 0)
			return binary.Write(w, endian, bs)
		},
	}
}

// Uni16NullTermString is the same as NullTermString, except that it works with UTF-16 strings.
func Uni16NullTermString(s *string) Mapper {
	if s == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			var (
				buf   bytes.Buffer
				wchar uint16
				u8s   = make([]byte, 4)
			)
			for {
				if err := binary.Read(r, endian, &wchar); err != nil {
					return err
				}
				if wchar == 0 {
					*s = buf.String()
					return nil
				}
				_rune := utf16.Decode([]uint16{wchar})
				n := utf8.EncodeRune(u8s, _rune[0])
				buf.Write(u8s[:n])
			}
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			var u16str []uint16
			for _, ru := range []rune(*s) {
				u16str = utf16.AppendRune(u16str, ru)
			}
			u16str = append(u16str, 0)
			if err := binary.Write(w, endian, u16str); err != nil {
				return err
			}
			return nil
		},
	}
}

// Uni16FixedString is the same as FixedString, except that it works with UTF-16 strings.
func Uni16FixedString(s *string, wcharlen int) Mapper {
	if s == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			var (
				buf = make([]uint16, wcharlen)
			)
			if err := binary.Read(r, endian, buf); err != nil {
				return err
			}
			val := string(utf16.Decode(buf))
			*s = strings.TrimRightFunc(val, func(r rune) bool {
				return r == 0
			})
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			var buf []uint16
			runes := []rune(*s)
			for i := 0; i < wcharlen && i < len(runes); i++ {
				buf = utf16.AppendRune(buf, runes[i])
			}
			out := make([]uint16, wcharlen)
			copy(out, buf)
			if err := binary.Write(w, endian, out); err != nil {
				return err
			}
			return nil
		},
	}
}
