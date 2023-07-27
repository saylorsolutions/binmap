package bin

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	initFieldCap = 16
)

var (
	ErrUnbalancedTable = errors.New("unbalanced data table")
)

// DataTable will construct a Mapper that orchestrates reading and writing a data table.
// This is very helpful for situations where the caller is using the array of structs to struct of arrays optimization, and wants to persist this table.
// Each FieldMapper will be used to read a single field element, making up a DataTable row, before returning to the first FieldMapper to start the next row.
// The length parameter will set during read, and read during write to ensure that all mapped fields are of the same length.
func DataTable(length *uint32, mappers ...FieldMapper) Mapper {
	if length == nil {
		return nilMapping
	}
	return Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			if err := Size(length).Read(r, endian); err != nil {
				return err
			}
			l := *length
			i := uint32(0)
			for i < l {
				for _, m := range mappers {
					if err := m.readNext(r, endian); err != nil {
						return err
					}
				}
				i++
			}
			for _, m := range mappers {
				m.apply()
			}
			return nil
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			l := *length
			for _, m := range mappers {
				if err := m.assertLen(l); err != nil {
					return err
				}
			}
			if err := Size(&l).Write(w, endian); err != nil {
				return err
			}

			i := uint32(0)
			for i < l {
				for _, m := range mappers {
					if err := m.writeNext(w, endian); err != nil {
						return err
					}
				}
				i++
			}
			return nil
		},
	)
}

// FieldMapper provides the logic necessary to read and write DataTable fields.
// Created with MapField.
type FieldMapper interface {
	readNext(r io.Reader, endian binary.ByteOrder) error
	apply()
	assertLen(uint32) error
	writeNext(w io.Writer, endian binary.ByteOrder) error
}

// MapField will associate a Mapper to each element in a target slice within a FieldMapper.
func MapField[T any](target *[]T, mapFn func(*T) Mapper) FieldMapper {
	return &fieldMapper[T]{
		fieldReader: &fieldReader[T]{
			target: target,
			fn:     mapFn,
			buf:    make([]T, 0, initFieldCap),
		},
		fieldWriter: &fieldWriter[T]{
			target: target,
			fn:     mapFn,
		},
	}
}

type fieldMapper[T any] struct {
	*fieldReader[T]
	*fieldWriter[T]
}

type fieldReader[T any] struct {
	target *[]T
	buf    []T
	fn     func(*T) Mapper
}

func (fr *fieldReader[T]) readNext(r io.Reader, endian binary.ByteOrder) error {
	var t T
	if err := fr.fn(&t).Read(r, endian); err != nil {
		return err
	}
	fr.buf = append(fr.buf, t)
	if len(fr.buf) == cap(fr.buf) {
		newBuf := make([]T, len(fr.buf), len(fr.buf)*2)
		copy(newBuf, fr.buf)
		fr.buf = newBuf
	}
	return nil
}

func (fr *fieldReader[T]) apply() {
	*fr.target = make([]T, len(fr.buf))
	copy(*fr.target, fr.buf)
	fr.buf = fr.buf[:0]
}

type fieldWriter[T any] struct {
	target *[]T
	fn     func(*T) Mapper
	wrPtr  uint32
}

func (fw *fieldWriter[T]) assertLen(length uint32) error {
	if uint32(len(*fw.target)) != length {
		return ErrUnbalancedTable
	}
	return nil
}

func (fw *fieldWriter[T]) next() *T {
	var t T
	t = (*fw.target)[fw.wrPtr]
	return &t
}

func (fw *fieldWriter[T]) writeNext(w io.Writer, endian binary.ByteOrder) error {
	if fw.wrPtr < uint32(len(*fw.target)) {
		if err := fw.fn(fw.next()).Write(w, endian); err != nil {
			return err
		}
		fw.wrPtr++
		return nil
	}
	return errors.New("at the end of the source buffer in writeNext")
}
