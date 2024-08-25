package bin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestNilConditional(t *testing.T) {
	var buf bytes.Buffer
	m := Conditional(func() bool { return true }, nil)
	assert.ErrorIs(t, m.Read(&buf, binary.BigEndian), ErrNilReadWrite)
	assert.ErrorIs(t, m.Write(&buf, binary.BigEndian), ErrNilReadWrite)
}

func TestConditional(t *testing.T) {
	var (
		buf    bytes.Buffer
		endian = binary.LittleEndian
		ival   uint64
		iother uint64
	)
	trueCase := Conditional(func() bool { return true }, Int(&ival))
	falseCase := Conditional(func() bool { return false }, Int(&ival), Int(&iother))
	noopCase := Conditional(func() bool { return false }, Int(&ival))

	ival = 5
	assert.NoError(t, trueCase.Write(&buf, endian))
	assert.Equal(t, uint64(5), ival)
	ival = 10
	assert.NoError(t, trueCase.Read(&buf, endian))
	assert.Equal(t, uint64(5), ival)

	iother = 5
	assert.NoError(t, falseCase.Write(&buf, endian))
	assert.Equal(t, uint64(5), ival)
	assert.Equal(t, uint64(5), iother)
	ival = 10
	iother = 10
	assert.NoError(t, falseCase.Read(&buf, endian))
	assert.Equal(t, uint64(10), ival)
	assert.Equal(t, uint64(5), iother)

	bufLen := buf.Len()
	assert.NoError(t, noopCase.Write(&buf, endian))
	assert.Equal(t, bufLen, buf.Len())
	buf.Reset()
	assert.NoError(t, noopCase.Read(&buf, endian))
	assert.Equal(t, 0, buf.Len())
}

func ExampleConditional() {
	var (
		outputInt        = true
		val       uint16 = 5
		buf       bytes.Buffer
		endian    = binary.BigEndian
	)
	m := Conditional(
		func() bool {
			return outputInt
		},
		Int(&val),
		FixedPadding[uint64](2),
	)

	// Effective read/write
	_ = m.Write(&buf, endian)
	val = 10
	_ = m.Read(&buf, endian)
	fmt.Printf("Read back %d\n", val)
	if val != 5 {
		log.Fatalln("Should have read back 5")
	}

	// Disable this mapping.
	outputInt = false
	buf.Reset()
	val = 15
	_ = m.Write(&buf, endian)
	val = 0
	_ = m.Read(&buf, endian)
	fmt.Printf("Value is still %d\n", val)
	if val != 0 {
		log.Fatalln("Should have read back nothing")
	}

	// Output:
	// Read back 5
	// Value is still 0
}
