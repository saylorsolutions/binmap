package bin

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMagicNumber(t *testing.T) {
	var (
		magicNum = []byte{'a', 'b', 'c'}
		buf      bytes.Buffer
		endian   = BigEndian
	)

	m := MagicNumber(magicNum)
	assert.NoError(t, m.Write(&buf, endian))
	assert.Equal(t, 3, buf.Len())
	assert.Equal(t, magicNum, []byte{'a', 'b', 'c'})
	assert.Equal(t, magicNum, buf.Bytes())

	assert.NoError(t, m.Read(&buf, endian))
	buf.Write([]byte{'d', 'e', 'f'})
	assert.ErrorIs(t, m.Read(&buf, endian), ErrMagicMismatch)
	err := m.Read(&buf, endian)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrMagicMismatch))

	assert.ErrorIs(t, MagicNumber([]byte{}).Write(&buf, endian), ErrNilReadWrite, "Empty bytes should return nilMapper")
}
