package main

import (
	bin "github.com/saylorsolutions/binmap"
	"io"
)

type User struct {
	username string
}

func (u *User) mapper() bin.Mapper {
	return bin.NullTermString(&u.username)
}

func (u *User) Read(r io.Reader) error {
	return u.mapper().Read(r, bin.BigEndian)
}

func (u *User) Write(w io.Writer) error {
	return u.mapper().Write(w, bin.BigEndian)
}
