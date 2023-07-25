package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAny(t *testing.T) {
	test := struct {
		a int16
		b uint64
	}{
		a: -5,
		b: 27,
	}
	seq := MapSequence(Int(&test.a), Int(&test.b))
	m := Any(&test, seq.Read, seq.Write)

	var (
		buf bytes.Buffer
	)
	assert.NoError(t, m.Write(&buf, binary.BigEndian))
	test.a, test.b = 0, 0
	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Equal(t, int16(-5), test.a)
	assert.Equal(t, uint64(27), test.b)
}

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

func TestFixedString(t *testing.T) {
	const (
		expected = "Hi"
	)
	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	s := expected
	m := FixedString(&s, 5)
	assert.NoError(t, m.Write(&buf, endian))
	out := buf.Bytes()
	assert.Equal(t, []byte{'H', 'i', 0, 0, 0}, out)

	s = ""
	buf.Reset()
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, endian))
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
}

func TestLenBytes(t *testing.T) {
	data := []byte("Hello!")
	test := struct {
		len  uint16
		data []byte
	}{
		len:  uint16(len(data)),
		data: data,
	}
	m := LenBytes(&test.data, &test.len)

	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	assert.NoError(t, m.Write(&buf, endian))
	assert.Equal(t, 8, buf.Len())

	test.len, test.data = 0, nil
	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, uint16(6), test.len)
	assert.Equal(t, "Hello!", string(test.data))
}

func TestLenSlice(t *testing.T) {
	test := struct {
		len  uint8
		data []byte
	}{
		len:  6,
		data: []byte("Hello!"),
	}
	m := LenSlice(&test.data, &test.len, func(r *byte) Mapping { return Byte(r) })

	var (
		buf    bytes.Buffer
		endian = binary.BigEndian
	)
	assert.NoError(t, m.Write(&buf, endian))
	test.len, test.data = 0, nil

	assert.NoError(t, m.Read(&buf, endian))
	assert.Equal(t, uint8(6), test.len)
	assert.Equal(t, []byte("Hello!"), test.data)
}

func TestDynamicSlice(t *testing.T) {
	data := []int16{1, -2, 3}
	m := DynamicSlice(&data, func(b *int16) Mapping {
		return Int(b)
	})

	var (
		buf    bytes.Buffer
		endian = binary.LittleEndian
	)
	assert.NoError(t, m.Write(&buf, endian))
	data = nil

	assert.NoError(t, m.Read(&buf, endian))
	assert.Len(t, data, 3)
	assert.Equal(t, []int16{1, -2, 3}, data)
}
