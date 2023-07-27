package main

import (
	"encoding/binary"
	bin "github.com/saylorsolutions/binmap"
	"io"
)

type User struct {
	id           uint64
	username     string
	passwordHash []byte
}

func (u *User) mapper() bin.Mapper {
	return bin.MapSequence(
		bin.Int(&u.id),
		bin.NullTermString(&u.username),
		bin.DynamicSlice(&u.passwordHash, bin.Byte),
	)
}

func (u *User) Read(r io.Reader) error {
	return u.mapper().Read(r, binary.BigEndian)
}

func (u *User) Write(w io.Writer) error {
	return u.mapper().Write(w, binary.BigEndian)
}
