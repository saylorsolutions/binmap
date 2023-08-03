package bin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestAny(t *testing.T) {
	test := struct {
		a int16
		b uint64
	}{}
	seq := MapSequence(Int(&test.a), Int(&test.b))
	seq = NormalizeWrite(seq, func() error {
		test.a, test.b = -5, 27
		return nil
	})
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

func TestOverrideEndian(t *testing.T) {
	const (
		expected = "Go"
	)
	s := expected
	m := ValidateRead(OverrideEndian(Uni16NullTermString(&s), binary.LittleEndian), func(err error) error {
		assert.Equal(t, expected, s)
		return err
	})
	var buf bytes.Buffer

	assert.NoError(t, m.Write(&buf, binary.BigEndian))
	out := buf.Bytes()
	assert.Equal(t, []byte{'G', 0, 'o', 0, 0, 0}, out)

	buf.Reset()
	buf.Write(out)
	s = ""

	assert.NoError(t, m.Read(&buf, binary.BigEndian))
}

func TestEventHandler_Read(t *testing.T) {
	data := struct {
		a uint16
		b bool
	}{}

	var (
		buf     bytes.Buffer
		ErrTest = errors.New("testing")
	)
	buf.Write([]byte{0, 5, 1})

	m := MapSequence(
		Int(&data.a),
		Bool(&data.b),
	)
	m = NewEventHandler(m, EventHandler{
		BeforeRead: func() error {
			assert.Equal(t, 3, buf.Len())
			assert.Equal(t, uint16(0), data.a)
			assert.Equal(t, false, data.b)
			return nil
		},
		AfterRead: func(err error) error {
			assert.NoError(t, err)
			assert.Equal(t, 0, buf.Len())
			assert.Equal(t, uint16(5), data.a)
			assert.Equal(t, true, data.b)
			return ErrTest
		},
	})
	assert.ErrorIs(t, m.Read(&buf, binary.BigEndian), ErrTest)
}

func TestEventHandler_Write(t *testing.T) {
	data := struct {
		a uint16
		b bool
	}{
		a: 5,
		b: true,
	}

	var (
		buf     bytes.Buffer
		ErrTest = errors.New("testing")
	)

	m := MapSequence(
		Int(&data.a),
		Bool(&data.b),
	)
	m = NewEventHandler(m, EventHandler{
		BeforeWrite: func() error {
			assert.Equal(t, 0, buf.Len())
			assert.Equal(t, uint16(5), data.a)
			assert.Equal(t, true, data.b)
			return nil
		},
		AfterWrite: func(err error) error {
			assert.NoError(t, err)
			assert.Equal(t, 3, buf.Len())
			assert.Equal(t, uint16(5), data.a)
			assert.Equal(t, true, data.b)
			return ErrTest
		},
	})
	assert.ErrorIs(t, m.Write(&buf, binary.LittleEndian), ErrTest)
}

func TestEventHandler_ReadWrite_Neg(t *testing.T) {
	var ErrTesting = errors.New("shouldn't be returned")
	mapper := Mapper(nilMapping)
	mapper = NewEventHandler(mapper, EventHandler{
		AfterRead: func(err error) error {
			return ErrTesting
		},
		AfterWrite: func(err error) error {
			return ErrTesting
		},
	})
	assert.ErrorIs(t, mapper.Read(nil, nil), ErrNilReadWrite, "The underlying error should be returned if it is non-nil")
	assert.ErrorIs(t, mapper.Write(nil, nil), ErrNilReadWrite, "The underlying error should be returned if it is non-nil")
}

func TestOnPanic(t *testing.T) {
	mapping := Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			panic("reading")
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			panic("writing")
		},
	)
	assert.Panics(t, func() {
		_ = mapping.Read(nil, nil)
	})
	assert.Panics(t, func() {
		_ = mapping.Write(nil, nil)
	})
	onPanic := OnPanic(mapping, func(a any) error {
		return nil
	})
	assert.NotPanics(t, func() {
		assert.ErrorIs(t, onPanic.Read(nil, nil), ErrPanic)
	})
	assert.NotPanics(t, func() {
		assert.ErrorIs(t, onPanic.Write(nil, nil), ErrPanic)
	})
	onPanic = OnPanic(mapping, func(a any) error {
		return errors.New("panic error")
	})
	assert.NotPanics(t, func() {
		assert.ErrorIs(t, onPanic.Read(nil, nil), ErrPanic)
	})
	assert.NotPanics(t, func() {
		assert.ErrorIs(t, onPanic.Write(nil, nil), ErrPanic)
	})
}
