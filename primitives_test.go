package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestByte(t *testing.T) {
	var (
		buf bytes.Buffer
		a   byte
		b   byte
	)
	buf.Write([]byte{0x01, 0x02})
	m := MapSequence(Byte(&a), Byte(&b))
	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Equal(t, byte(0x01), a)
	assert.Equal(t, byte(0x02), b)

	a, b = b, a
	buf.Reset()
	assert.NoError(t, m.Write(&buf, binary.BigEndian))
	assert.Equal(t, []byte{0x02, 0x01}, buf.Bytes())
}

func TestBool(t *testing.T) {
	var (
		buf bytes.Buffer
		a   bool
		b   bool
	)
	buf.Write([]byte{0x01, 0x01})
	m := MapSequence(Bool(&a), Bool(&b))
	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Equal(t, true, a)
	assert.Equal(t, true, b)

	a, b = false, false
	buf.Reset()
	assert.NoError(t, m.Write(&buf, binary.BigEndian))
	assert.Equal(t, []byte{0x00, 0x00}, buf.Bytes())
}

func TestFloat(t *testing.T) {
	var (
		buf    bytes.Buffer
		endian = binary.LittleEndian
	)
	f1 := float32(0.5)
	f2 := 1.5
	m := MapSequence(Float(&f1), Float(&f2))
	assert.NoError(t, m.Write(&buf, endian))

	f1, f2 = 0, 0
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, float32(0.5), f1)
	assert.Equal(t, 1.5, f2)
}

func TestComplex(t *testing.T) {
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	c1 := complex(float32(3.14), float32(1))
	c2 := complex(4.13, 5)

	m := MapSequence(
		Complex(&c1),
		Complex(&c2),
	)

	assert.NoError(t, m.Write(&buf, endian))
	c1, c2 = 0, 0
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, complex(float32(3.14), float32(1)), c1)
	assert.Equal(t, complex(4.13, 5), c2)
}

func TestVarint(t *testing.T) {
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	v1 := int64(257)
	v2 := int64(258)

	m := MapSequence(
		Varint(&v1),
		Varint(&v2),
	)

	assert.NoError(t, m.Write(&buf, endian))
	v1, v2 = 0, 0
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, int64(257), v1)
	assert.Equal(t, int64(258), v2)
}

func TestUvarint(t *testing.T) {
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	v1 := uint64(257)
	v2 := uint64(258)

	m := MapSequence(
		Uvarint(&v1),
		Uvarint(&v2),
	)

	assert.NoError(t, m.Write(&buf, endian))
	v1, v2 = 0, 0
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, uint64(257), v1)
	assert.Equal(t, uint64(258), v2)
}
