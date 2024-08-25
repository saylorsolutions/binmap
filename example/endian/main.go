package main

import (
	"bytes"
	"encoding/binary"
	bin "github.com/saylorsolutions/binmap"
	"io"
	"log"
)

type Header struct {
	Endianness byte
	DataSize   uint16
	Data       []byte
}

func (h *Header) mapper() bin.Mapper {
	return bin.MapSequence(
		bin.Byte(&h.Endianness),
		bin.MapEndian(bin.EndianInt(&h.Endianness, 1, 0), bin.Int(&h.DataSize)),
		bin.Any(
			// Not using LenBytes because it would duplicate the length in the output
			func(r io.Reader, endian binary.ByteOrder) error {
				l := h.DataSize
				return bin.FixedBytes(&h.Data, l).Read(r, endian)
			},
			func(w io.Writer, endian binary.ByteOrder) error {
				l := h.DataSize
				return bin.FixedBytes(&h.Data, l).Write(w, endian)
			},
		),
	)
}

func main() {
	h := &Header{
		Endianness: 1, // Big endian
		DataSize:   32,
		Data:       make([]byte, 32),
	}
	var buf bytes.Buffer

	// The size will still be overridden to big endian, even though we're passing little endian byte order.
	if err := h.mapper().Write(&buf, binary.LittleEndian); err != nil {
		log.Fatalln("Failed to write header and data to buffer:", err)
	}

	data := buf.Bytes()
	// Endianness byte, 2-byte size, and the 32-byte data slice.
	if len(data) != 35 {
		log.Println("Unexpected number of bytes written,", len(data))
	}
	// Offset 0 is the endianness flag, 1 should be zero for the big side of the uint16, making offset 2 the actual value.
	if data[2] != 32 {
		log.Printf("%#v\n", data)
		log.Fatalln("Should have written the data size as a big endian value")
	}
	log.Println("Data successfully written to buffer with big endian override")
}
