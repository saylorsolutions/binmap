# Binmap

I've found that using the stdlib `binary` interface to read and write data is a little cumbersome and tedious, since any operation can result in an error.
While this makes sense given the problem domain, the API leaves something to be desired.

I'd love to have a way to batch operations, so I don't have so much `if err != nil`.
If an error occurs at any point, then I'm able to fail fast and handle one error at the end.

I'd also like to work easily with `io.Reader`s rather than having to read everything into memory first to dissect it piecemeal.
While this *can* be accomplished with `binary.Read`, I still have the issue of too much error handling code cluttered around the code I *want* to write.

## Goals

* I'd like to have an easier to use interface for reading/writing binary data.
* I'd like to declare binary IO operations, execute them, and handle a single error at the end.
* I'd like to be able to reuse binary IO operations, and even pass them into more complex pipelines.
* I'd like to be able to declare dynamic behavior, like when the size of the next read is determined by the current field.
* I'd like to declare a read loop based on a read field value, and pass the loop construct to a larger pipeline.
* ~~Struct tag field binding would be fantastic, but reflection is... fraught. I'll see how this goes, and I'll probably take some hints from how the stdlib is handling this.~~
  * There's too much possibility of dynamic or dependent logic with a lot of binary payloads, and the number of edge cases for implementing this is more than I want to deal with.
  * I'm pretty happy with the API for mapping definition so far, and I'd rather simplify that than get into reflection with struct field tags. I feel like it's much more understandable (and thus maintainable) code.

## How it works

This package centers around the `Mapper` interface.
A mapper is anything that knows how to read and write binary data, and is easily adaptable to new data types with custom logic with the `Any` mapper.

Any given `Mapper` is expected to be short-lived, especially if the underlying data representation in Go code is changing often.
This mechanism makes heavy use of pointers, and even pointer-pointers in some cases, which means that there's a fair bit of indirection to make this work.
There are also a lot of generics in this code to both limit input types to what is readily supported, and to keep duplication to a minimum.

> Note that using different mapper procedures between software versions is effectively a **breaking change**, and should be handled like you would handle database migrations.
> There are certain patterns that make this easier to work with, explained below.

### Directly supported types

There are several primitive types that are directly supported.
Keep in mind that type restrictions mostly come from what [binary.Read and binary.Write support](https://pkg.go.dev/encoding/binary), and this package also inherits the design constraints of simplicity over speed mentioned in the encoding/binary docs.

* Integers with `Int`.
  * Note that `int` and `uint` are *not* supported because these are not necessarily of a known binary size at compile time.
* Floats with `Float`.
* Booleans with `Bool`.
* Bytes with `Byte`, and byte slices with `FixedBytes` and `LenBytes`.
* General slice mappers are provided with `Slice`, `LenSlice`, and `DynamicSlice`.
* Size types with `Size`, which are restricted to any known-size, unsigned integer.
* Strings, both with `FixedString` for fixed-width string fields, and null-terminated strings with `NullTermString`.
* More interesting types, such as `Map` for arbitrary maps, and even `DataTable` for persisting structs-of-arrays.
* As already mentioned, the `Any` mapper can be used to add arbitrary mapping logic for any type you'd like to express.
  * An `Any` mapper just needs a `ReadFunc` and `WriteFunc`.
  * This mapper function doesn't require a target because it's intended to be flexible, and the assumption is that a target would be available in a closure context.

## Common patterns

There are few assumptions made about or constraints applied to your data representation, but all persisted data must either be of a fixed size when persisted, or include an unambiguous delimiter (like a null terminator for a string).
This means that you are charged with managing things like binary format migrations and validation.
Binary serialization can get pretty complicated, depending on the data structures involved.
Fortunately, there are some commonly used patterns and library features that help manage this complexity.

### Mapper method

Expressing a mapper method that creates a consistent `Mapper` for your data in a struct, and then using that to expose read and write methods seems to work well in practice.

```golang
import (
	"encoding/binary"
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
	return u.mapper().Read(r, binary.BigEndian)
}

func (u *User) Write(w io.Writer) error {
	return u.mapper().Write(w, binary.BigEndian)
}
```

### Mapper Sequence

The previous pattern can be extended to map more fields with `MapSequence`.
This provides a tremendous level of flexibility, since the result of `MapSequence` is itself a `Mapper`.

```golang
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
```

### Mapper of Mappers

Once the previous patterns have been established, extensions may be made for additional types within your data.
Types included in your top-level structure can themselves have a mapper method that specifies how *they* are binary mapped.

> Note: That the use of `LenSlice` is an arbitrary choice, and not a requirement of embedding slices of types in other types.
> It's generally preferred to use `DynamicSlice` unless you're encoding the length of a slice as a field in your struct, or you always know the length of a slice ahead of time.

```golang
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
```

This makes reading a struct from a binary source incredibly trivial, with a single error to handle regardless of the mapping logic expressed.

```golang
func ReadUser(r io.Reader) (*User, error) {
	u := new(User)
	if err := u.Read(r); err != nil {
		return nil, err
	}
	return u, nil
}
```

### Validated read

Input validation is important, especially in cases where changes in persisted data could lead to changes to a struct's internal, unexposed state.
This can easily be added in the Read and Write methods added above to ensure that business rule constraints are encoded as part of the persistence logic.

```golang
var ErrNoContact = errors.New("all users must have a contact")

func (u *User) Read(r io.Reader) error {
  if err := u.mapper().Read(r, binary.BigEndian); err != nil {
	  return err
  }
  if len(u.contacts) == 0 {
	  return ErrNoContact
  }
}

func (u *User) Write(w io.Writer) error {
  if len(u.contacts) == 0 {
    return ErrNoContact
  }
  if err := u.mapper().Write(w, binary.BigEndian); err != nil {
    return err
  }
}
```

### Versioned mapping

As mentioned previously, versioned mapping can be very important if the binary representation is expected to change (often or not), since it's effectively a breaking change.
This can be handled pretty readily with a little forethought.

```golang
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
```
