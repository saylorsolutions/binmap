package main

import (
	"encoding/binary"
	bin "github.com/saylorsolutions/binmap"
	"io"
)

type Contact struct {
	email          string
	allowMarketing bool
}

func (c *Contact) mapper() bin.Mapper {
	return bin.MapSequence(
		bin.FixedString(&c.email, 128),
		bin.Bool(&c.allowMarketing),
	)
}

type User struct {
	id           uint64
	username     string
	passwordHash []byte
	numContacts  uint16
	contacts     []Contact
}

func (u *User) mapper() bin.Mapper {
	return bin.MapSequence(
		bin.Int(&u.id),
		bin.NullTermString(&u.username),
		bin.DynamicSlice(&u.passwordHash, bin.Byte),
		bin.LenSlice(&u.contacts, &u.numContacts, func(c *Contact) bin.Mapper {
			return c.mapper()
		}),
	)
}

func (u *User) Read(r io.Reader) error {
	return u.mapper().Read(r, binary.BigEndian)
}

func (u *User) Write(w io.Writer) error {
	return u.mapper().Write(w, binary.BigEndian)
}
