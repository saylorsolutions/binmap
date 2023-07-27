package bin

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	data := map[uint8]bool{
		0: false,
		1: true,
		2: false,
		3: true,
	}

	m := Map(&data, Int[uint8], Bool)
	var buf bytes.Buffer
	assert.NoError(t, m.Write(&buf, binary.BigEndian))

	out := buf.Bytes()
	buf.Reset()
	assert.Equal(t, []byte{0x0, 0x0, 0x0, 0x4}, out[:4])

	data = map[uint8]bool{}
	buf.Write(out)
	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Len(t, data, 4)
	assert.Equal(t, data[0], false)
	assert.Equal(t, data[1], true)
	assert.Equal(t, data[2], false)
	assert.Equal(t, data[3], true)
}
