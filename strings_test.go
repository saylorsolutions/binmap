package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFixedString(t *testing.T) {
	const (
		expected = "Hi\x00you"
	)
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	s := expected
	m := FixedString(&s, 8)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{'H', 'i', 0, 'y', 'o', 'u', 0, 0}, out)

	s = ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, expected, s)
}

func TestNullTermString(t *testing.T) {
	const (
		expected = "Hi"
	)
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	s := expected
	m := NullTermString(&s)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{'H', 'i', 0}, out)

	s = ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, "Hi", s)
}
