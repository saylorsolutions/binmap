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
