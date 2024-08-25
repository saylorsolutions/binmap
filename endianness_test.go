package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestEndianInt(t *testing.T) {
	assert.Panics(t, func() {
		EndianInt((*uint64)(nil), 1, 0)
	}, "A nil field pointer should result in a panic")

	var field uint64
	ind := EndianInt(&field, 1, 2)
	field = 1
	assert.Equal(t, binary.BigEndian, ind())
	field = 2
	assert.Equal(t, binary.LittleEndian, ind())
	field = 3
	assert.Equal(t, binary.NativeEndian, ind())
}

func TestMapEndian(t *testing.T) {
	var (
		target uint16
		buf    bytes.Buffer
	)
	assertBig := Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			assert.Equal(t, binary.BigEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			assert.Equal(t, binary.BigEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assertLittle := Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			assert.Equal(t, binary.LittleEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			assert.Equal(t, binary.LittleEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assertNative := Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			assert.Equal(t, binary.NativeEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			assert.Equal(t, binary.NativeEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assert.Panics(t, func() {
		MapEndian(nil, assertBig)
	})
	assert.ErrorIs(t, MapEndian(OverrideNative, nil).Write(&buf, binary.BigEndian), ErrNilReadWrite)

	t.Run("Override Big", func(t *testing.T) {
		buf.Reset()
		target = 0x0110
		m := MapEndian(OverrideBig, assertBig)
		assert.NoError(t, m.Write(&buf, binary.LittleEndian))
		assert.Equal(t, uint16(0x0110), target)
		target = 0
		assert.NoError(t, m.Read(&buf, binary.LittleEndian))
		assert.Equal(t, uint16(0x0110), target)
	})

	t.Run("Override Little", func(t *testing.T) {
		buf.Reset()
		target = 0x1001
		m := MapEndian(OverrideLittle, assertLittle)
		assert.NoError(t, m.Write(&buf, binary.BigEndian))
		assert.Equal(t, uint16(0x1001), target)
		target = 0
		assert.NoError(t, m.Read(&buf, binary.BigEndian))
		assert.Equal(t, uint16(0x1001), target)
	})

	t.Run("Override Native", func(t *testing.T) {
		buf.Reset()
		target = 0x2002
		var (
			testBytes = make([]byte, 2)
			endian    binary.ByteOrder
		)
		binary.NativeEndian.PutUint16(testBytes, target)
		switch testBytes[0] {
		case 0x20:
			// Native is big endian, set input endian to little
			endian = binary.LittleEndian
		case 0x02:
			// Native is little endian, set input endian to little
			endian = binary.BigEndian
		default:
			t.Fatal("Unable to determine native endianness")
		}

		m := MapEndian(OverrideNative, assertNative)
		assert.NoError(t, m.Write(&buf, endian))
		assert.Equal(t, uint16(0x2002), target)
		target = 0
		assert.NoError(t, m.Read(&buf, endian))
		assert.Equal(t, uint16(0x2002), target)
	})
}
