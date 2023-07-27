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
	m := Any(seq.Read, seq.Write)

	var (
		buf bytes.Buffer
	)
	assert.NoError(t, m.Write(&buf, binary.BigEndian))
	test.a, test.b = 0, 0
	assert.NoError(t, m.Read(&buf, binary.BigEndian))
	assert.Equal(t, int16(-5), test.a)
	assert.Equal(t, uint64(27), test.b)
}
