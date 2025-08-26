package bin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
)

var (
	ErrNotSeeker    = errors.New("not a Seeker")
	ErrInvalidLimit = errors.New("limit offset is invalid")
)

func ensureSeeker(r any) (io.Seeker, error) {
	rs, ok := r.(io.Seeker)
	if !ok {
		return nil, ErrNotSeeker
	}
	return rs, nil
}

// OffsetType represents a positive or negative offset within binary data.
// Positive offsets are interpreted to be offsets from the beginning.
// Negative offsets are interpreted to be offsets from the end.
type OffsetType interface {
	~int64 | ~int32 | ~int16 | ~int8
}

// Offset read/writes an OffsetType for later use in OffsetData.
func Offset[O OffsetType](offset *O) Mapper {
	if offset == nil {
		return nilMapping
	}
	return Any(func(r io.Reader, endian ByteOrder) error {
		return Int(offset).Read(r, endian)
	}, func(w io.Writer, endian ByteOrder) error {
		return Int(offset).Write(w, endian)
	})
}

// OffsetData reads/writes data at a given offset.
// Writes beyond the end offset will append data, appending padding zero bytes if necessary to reach the given offset.
//
// Note that seeks to offsets are not reset to the original after mapping.
// This is to limit the performance impact of random seeking within a binary, but it also means that offset reads/writes should be deferred to the last mapping executed for consistency.
//
// This is useful for many binary file formats, where a header section details the layout of the rest of the file with offsets and optionally sizes.
// In such a case, it may be necessary to separate header and offset reading, to ensure that all offsets have been read and sizes calculated.
func OffsetData[O OffsetType, S SizeType](offset *O, data *[]byte, size *S) Mapper {
	if offset == nil || data == nil {
		return nilMapping
	}
	var eofSize S = 0
	if size == nil {
		size = &eofSize
	}
	return Any(func(r io.Reader, endian ByteOrder) (rerr error) {
		seeker, err := ensureSeeker(r)
		if err != nil {
			return err
		}
		off := int64(*offset)
		if off < 0 {
			_, err = seeker.Seek(off, io.SeekEnd)
		} else {
			_, err = seeker.Seek(off, io.SeekStart)
		}
		if err != nil {
			return fmt.Errorf("failed to seek to designated offset %d: %w", off, err)
		}
		if *size > 0 {
			buf := make([]byte, *size)
			_, err = io.ReadFull(r, buf)
			if err != nil {
				return fmt.Errorf("failed to read expected size %d: %w", *size, err)
			}
			*data = buf
			return nil
		}
		remaining, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read remaining data from offset %d: %w", off, err)
		}
		*data = remaining
		return nil
	}, func(w io.Writer, endian ByteOrder) error {
		seeker, err := ensureSeeker(w)
		if err != nil {
			return err
		}
		off := int64(*offset)
		end, err := seeker.Seek(0, io.SeekEnd)
		if err != nil {
			return fmt.Errorf("failed to get output size: %w", err)
		}
		nilByte := []byte{0x0}
		if off > end {
			// Write zero padding until we get to offset
			for i := end; i < off; i++ {
				_, err := w.Write(nilByte)
				if err != nil {
					return fmt.Errorf("failed to write zero padding")
				}
			}
		} else {
			// Seek to offset to begin writing
			if off < 0 {
				_, err = seeker.Seek(off, io.SeekEnd)
			} else {
				_, err = seeker.Seek(off, io.SeekStart)
			}
			if err != nil {
				return fmt.Errorf("failed to seek to designated offset %d: %w", off, err)
			}
		}
		if *size > 0 {
			return FixedBytes(data, uint64(*size)).Write(w, endian)
		}
		return FixedBytes(data, uint64(len(*data))).Write(w, endian)
	})
}

func OffsetDataMapper[O OffsetType, S SizeType](offset *O, mapper Mapper, size *S) Mapper {
	return Any(
		func(r io.Reader, endian ByteOrder) error {
			var data []byte
			if err := OffsetData(offset, &data, size).Read(r, endian); err != nil {
				return err
			}
			return mapper.Read(bytes.NewReader(data), endian)
		},
		func(w io.Writer, endian ByteOrder) error {
			var buf bytes.Buffer
			if err := mapper.Write(&buf, endian); err != nil {
				return err
			}
			data := buf.Bytes()
			return OffsetData(offset, &data, size).Write(w, endian)
		},
	)
}

// OffsetSection is a convenience structure provided as a mapping target for sections of a binary referenced by offset.
//
// The generic OffsetType and SizeType dictates the size of the values when read and written.
// This means that narrowing conversions can be lossy if the correct datatype is not used.
//
// It's usually necessary to set a section size.
// Not setting a size - or explicitly setting it to zero - will result in reading the remainder of the binary into OffsetSection.Data.
// This is likely valid behavior for at most one section.
type OffsetSection[O OffsetType, S SizeType] struct {
	Offset int64
	Size   uint64
	Data   []byte
	Mapper Mapper
}

func (s *OffsetSection[O, S]) HeaderOffset() Mapper {
	return Any(func(r io.Reader, endian ByteOrder) error {
		var off = O(s.Offset)
		if err := Offset(&off).Read(r, endian); err != nil {
			return err
		}
		s.Offset = int64(off)
		return nil
	}, func(w io.Writer, endian ByteOrder) error {
		var off = O(s.Offset)
		return Offset(&off).Write(w, endian)
	})
}

func (s *OffsetSection[O, S]) HeaderSize() Mapper {
	return Any(func(r io.Reader, endian ByteOrder) error {
		var sz = S(s.Size)
		if err := Size(&sz).Read(r, endian); err != nil {
			return err
		}
		s.Size = uint64(sz)
		return nil
	}, func(w io.Writer, endian ByteOrder) error {
		var sz = S(s.Size)
		return Size(&sz).Write(w, endian)
	})
}

func (s *OffsetSection[O, S]) HeaderSizeThenOffset() Mapper {
	return MapSequence(s.HeaderSize(), s.HeaderOffset())
}

func (s *OffsetSection[O, S]) HeaderOffsetThenSize() Mapper {
	return MapSequence(s.HeaderOffset(), s.HeaderSize())
}

func (s *OffsetSection[O, S]) SectionData() Mapper {
	return Any(
		func(r io.Reader, endian ByteOrder) error {
			if s.Mapper != nil {
				return OffsetDataMapper(&s.Offset, s.Mapper, &s.Size).Read(r, endian)
			}
			return OffsetData(&s.Offset, &s.Data, &s.Size).Read(r, endian)
		},
		func(w io.Writer, endian ByteOrder) error {
			if s.Mapper != nil {
				return OffsetDataMapper(&s.Offset, s.Mapper, &s.Size).Write(w, endian)
			}
			return OffsetData(&s.Offset, &s.Data, &s.Size).Write(w, endian)
		},
	)
}

// LimitSizeToOffset sets the size of this OffsetSection such that it cannot overlap the readLimit offset.
// This is used by SizeAdjacentSections to ensure that section reads/writes cannot overlap.
//
// An error is returned if readLimit <= OffsetSection.Offset.
func (s *OffsetSection[O, S]) LimitSizeToOffset(readLimit int64) error {
	if s.Offset >= readLimit {
		return fmt.Errorf("%w: current offset %d is >= limit %d", ErrInvalidLimit, s.Offset, readLimit)
	}
	s.Size = uint64(readLimit - s.Offset)
	return nil
}

// AdjacentSection is a generic interface implemented by OffsetSection to allow populating adjacent section sizes using different OffsetType values.
type AdjacentSection interface {
	getOffset() int64
	LimitSizeToOffset(readLimit int64) error
}

func (s *OffsetSection[O, S]) getOffset() int64 {
	return s.Offset
}

var _ sort.Interface = (AdjacentSectionList)(nil)

// AdjacentSectionList is used by SizeAdjacentSections to set the size of many adjacent OffsetSection.
type AdjacentSectionList []AdjacentSection

func (s AdjacentSectionList) Len() int {
	return len(s)
}

func (s AdjacentSectionList) Less(i, j int) bool {
	return s[i].getOffset() < s[j].getOffset()
}

func (s AdjacentSectionList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SizeAdjacentSections is used to set OffsetSection sizes when they are adjacent to each other.
//
// Sections will be sorted before comparing offsets.
// An error will be returned if two sections have the same offset.
func SizeAdjacentSections(sections ...AdjacentSection) error {
	if len(sections) <= 1 {
		return nil
	}
	sorted := AdjacentSectionList(sections)
	sort.Sort(sorted)
	for i := 1; i < len(sorted); i++ {
		a := i - 1
		b := i
		if err := sorted[a].LimitSizeToOffset(sorted[b].getOffset()); err != nil {
			return err
		}
	}
	return nil
}
