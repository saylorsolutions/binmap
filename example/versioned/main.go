package main

import (
	"encoding/binary"
	"errors"
	bin "github.com/saylorsolutions/binmap"
	"io"
)

type version = byte

const (
	v1 version = iota + 1
	v2
)

type User struct {
	username string
}

func (u *User) mapperV1() bin.Mapper {
	return bin.NullTermString(&u.username)
}

func (u *User) mapperV2() bin.Mapper {
	return bin.FixedString(&u.username, 32)
}

func (u *User) mapper() bin.Mapper {
	return bin.Any(
		func(r io.Reader, endian binary.ByteOrder) error {
			var v version
			if err := bin.Byte(&v).Read(r, endian); err != nil {
				return err
			}
			switch v {
			case v1:
				return u.mapperV1().Read(r, endian)
			case v2:
				return u.mapperV2().Read(r, endian)
			default:
				return errors.New("unknown version")
			}
		},
		func(w io.Writer, endian binary.ByteOrder) error {
			var v = v2
			return bin.MapSequence(
				bin.Byte(&v),
				u.mapperV2(),
			).Write(w, endian)
		},
	)
}

func (u *User) Read(r io.Reader) error {
	return u.mapper().Read(r, binary.BigEndian)
}

func (u *User) Write(w io.Writer) error {
	return u.mapper().Write(w, binary.BigEndian)
}
