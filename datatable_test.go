package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataTable(t *testing.T) {
	a := []byte("H,tee")
	b := []byte("i hr!")

	var (
		buf    bytes.Buffer
		length = uint32(len(a))
	)

	m := DataTable(&length,
		MapField(&a, Byte),
		MapField(&b, Byte),
	)
	assert.NoError(t, m.Write(&buf, binary.BigEndian))

	written := buf.Bytes()
	expected := append([]byte{0, 0, 0, 5}, "Hi, there!"...)
	assert.Equal(t, expected, written)

	buf.Reset()
	buf.Write(written)
	a, b = nil, nil

	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Equal(t, "H,teei hr!", string(append(a, b...)))
}
