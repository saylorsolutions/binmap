package bin

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"testing"
)

func TestEndianInt(t *testing.T) {
	assert.Panics(t, func() {
		EndianInt((*uint64)(nil), 1, 0)
	}, "A nil field pointer should result in a panic")

	var field uint64
	ind := EndianInt(&field, 1, 2)
	field = 1
	assert.Equal(t, BigEndian, ind())
	field = 2
	assert.Equal(t, LittleEndian, ind())
	field = 3
	assert.Equal(t, NativeEndian, ind())
}

func TestMapEndian(t *testing.T) {
	var (
		target uint16
		buf    bytes.Buffer
	)
	assertBig := Any(
		func(r io.Reader, endian ByteOrder) error {
			assert.Equal(t, BigEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian ByteOrder) error {
			assert.Equal(t, BigEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assertLittle := Any(
		func(r io.Reader, endian ByteOrder) error {
			assert.Equal(t, LittleEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian ByteOrder) error {
			assert.Equal(t, LittleEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assertNative := Any(
		func(r io.Reader, endian ByteOrder) error {
			assert.Equal(t, NativeEndian, endian)
			return Int(&target).Read(r, endian)
		},
		func(w io.Writer, endian ByteOrder) error {
			assert.Equal(t, NativeEndian, endian)
			return Int(&target).Write(w, endian)
		},
	)
	assert.Panics(t, func() {
		MapEndian(nil, assertBig)
	})
	assert.ErrorIs(t, MapEndian(OverrideNative, nil).Write(&buf, BigEndian), ErrNilReadWrite)

	t.Run("Override Big", func(t *testing.T) {
		buf.Reset()
		target = 0x0110
		m := MapEndian(OverrideBig, assertBig)
		assert.NoError(t, m.Write(&buf, LittleEndian))
		assert.Equal(t, uint16(0x0110), target)
		target = 0
		assert.NoError(t, m.Read(&buf, LittleEndian))
		assert.Equal(t, uint16(0x0110), target)
	})

	t.Run("Override Little", func(t *testing.T) {
		buf.Reset()
		target = 0x1001
		m := MapEndian(OverrideLittle, assertLittle)
		assert.NoError(t, m.Write(&buf, BigEndian))
		assert.Equal(t, uint16(0x1001), target)
		target = 0
		assert.NoError(t, m.Read(&buf, BigEndian))
		assert.Equal(t, uint16(0x1001), target)
	})

	t.Run("Override Native", func(t *testing.T) {
		buf.Reset()
		target = 0x2002
		var (
			testBytes = make([]byte, 2)
			endian    ByteOrder
		)
		NativeEndian.PutUint16(testBytes, target)
		switch testBytes[0] {
		case 0x20:
			// Native is big endian, set input endian to little
			endian = LittleEndian
		case 0x02:
			// Native is little endian, set input endian to little
			endian = BigEndian
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

func ExampleMapEndian() {
	var (
		buf  bytes.Buffer
		data uint16 = 0x1001
	)
	m := MapEndian(OverrideBig, Int(&data))

	// The specified byte order will be overridden for the Int mapper.
	if err := m.Write(&buf, LittleEndian); err != nil {
		log.Fatalln("Error writing to buffer")
	}
	fmt.Printf("First byte should be 0x10: 0x%x\n", buf.Bytes()[0])

	data = 0
	// Reading works with the same override.
	if err := m.Read(&buf, LittleEndian); err != nil {
		log.Fatalln("Error reading back from buffer")
	}
	fmt.Printf("Should have read 0x1001: 0x%x\n", data)

	// Output:
	// First byte should be 0x10: 0x10
	// Should have read 0x1001: 0x1001
}
