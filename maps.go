package bin

import (
	"encoding/binary"
	"io"
)

type KeyMapper[K comparable] func(key *K) Mapper
type ValMapper[V any] func(val *V) Mapper

func Map[K comparable, V any](target *map[K]V, keyMapper KeyMapper[K], valMapper ValMapper[V]) Mapper {
	if target == nil {
		return nilMapping
	}
	return &mapper{
		read: func(r io.Reader, endian binary.ByteOrder) error {
			m := map[K]V{}
			var length uint32
			if err := Size(&length).Read(r, endian); err != nil {
				return err
			}
			i := uint32(0)
			for i < length {
				var (
					key K
					val V
				)
				err := keyMapper(&key).Read(r, endian)
				if err != nil {
					return err
				}
				err = valMapper(&val).Read(r, endian)
				if err != nil {
					return err
				}
				m[key] = val
				i++
			}
			*target = m
			return nil
		},
		write: func(w io.Writer, endian binary.ByteOrder) error {
			var length = uint32(len(*target))
			if err := Size(&length).Write(w, endian); err != nil {
				return err
			}
			for k, v := range *target {
				if err := keyMapper(&k).Write(w, endian); err != nil {
					return err
				}
				if err := valMapper(&v).Write(w, endian); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
