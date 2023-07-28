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
	s1 := expected
	s2 := expected
	m := MapSequence(
		NullTermString(&s1),
		NullTermString(&s2),
	)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{'H', 'i', 0, 'H', 'i', 0}, out)

	s1, s2 = "", ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, "Hi", s1)
	assert.Equal(t, "Hi", s2)
}

func TestUni16FixedString(t *testing.T) {
	const (
		expected = "Hi\x00you"
	)
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	s := expected
	m := Uni16FixedString(&s, 8)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{0, 'H', 0, 'i', 0, 0, 0, 'y', 0, 'o', 0, 'u', 0, 0, 0, 0}, out)

	s = ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, expected, s)
}

func TestUni16NullTermString(t *testing.T) {
	const (
		expected = "Hi"
	)
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	s1 := expected
	s2 := expected
	m := MapSequence(
		Uni16NullTermString(&s1),
		Uni16NullTermString(&s2),
	)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{0, 'H', 0, 'i', 0, 0, 0, 'H', 0, 'i', 0, 0}, out)

	s1, s2 = "", ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, "Hi", s1)
	assert.Equal(t, "Hi", s2)
}
