package main

import (
	"encoding/binary"
	"errors"
	bin "github.com/saylorsolutions/binmap"
	"io"
)

type Contact struct {
	email          string
	allowMarketing bool
}

func (c *Contact) mapper() bin.Mapper {
	m := bin.MapSequence(
		bin.FixedString(&c.email, 128),
		bin.Bool(&c.allowMarketing),
	)
	m = bin.ValidateRead(m, func(err error) error {
		if err != nil {
			return err
		}
		if len(c.email) == 0 {
			return errors.New("empty email")
		}
		return nil
	})
	m = bin.NormalizeWrite(m, func() error {
		if len(c.email) == 0 {
			return errors.New("empty email")
		}
		return nil
	})
	return m
}

type User struct {
	id           uint64
	username     string
	passwordHash []byte
	contacts     []Contact
}

func (u *User) mapper() bin.Mapper {
	return bin.MapSequence(
		bin.Int(&u.id),
		bin.NullTermString(&u.username),
		bin.DynamicSlice(&u.passwordHash, bin.Byte),
		bin.DynamicSlice(&u.contacts, func(c *Contact) bin.Mapper {
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
