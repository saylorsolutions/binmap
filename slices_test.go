package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
	m := LenSlice(&test.data, &test.len, func(r *byte) Mapper { return Byte(r) })

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
	m := DynamicSlice(&data, func(b *int16) Mapper {
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
