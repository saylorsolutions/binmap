package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
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
