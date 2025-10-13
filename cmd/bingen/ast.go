package main

import (
	"regexp"
	"strconv"
	"strings"
)

// MapType specifies a type of data encoding for binary data.
// A stream of MapType (with encoded arguments where needed) can be used to define a procedure for reading binary data.
//
// This is most useful for reading more complex data where manually specifying a mapping would otherwise be tedious.
type MapType string

const (
	// ID is used to identify the top level structure.
	ID          MapType = "id"
	Magic       MapType = "magic"
	I8          MapType = "i8"
	I16         MapType = "i16"
	I32         MapType = "i32"
	I64         MapType = "i64"
	VarInt      MapType = "vi"
	UVarInt     MapType = "uvi"
	U8          MapType = "u8"
	U16         MapType = "u16"
	U32         MapType = "u32"
	U64         MapType = "u64"
	BitFlag8    MapType = "bf8"
	BitFlag16   MapType = "bf16"
	BitFlag32   MapType = "bf32"
	BitFlag64   MapType = "bf64"
	BitFlagEq8  MapType = "bfe8"
	BitFlagEq16 MapType = "bfe16"
	BitFlagEq32 MapType = "bfe32"
	BitFlagEq64 MapType = "bfe64"
	BitFlagVal  MapType = "bf"
	F32         MapType = "f32"
	F64         MapType = "f64"
	Len8        MapType = "l8"
	Len16       MapType = "l16"
	Len32       MapType = "l32"
	Len64       MapType = "l64"
	LenVar      MapType = "lv"
	LenData     MapType = "ldata"
	FixedData   MapType = "fdata"
	Data        MapType = "data"
	NullTermStr MapType = "ntstr"
	LenStr      MapType = "lstr"
	Array       MapType = "ary"
	Struct      MapType = "struct"
	StructField MapType = "field"
	StructPad   MapType = "pad" // StructPad defines the padding in bytes for a Struct.
	End         MapType = "end" // Used to end the innermost Struct or Union definition.
	Arg         MapType = "arg"
)

func (t MapType) String() string {
	return string(t)
}

func matchType(str string) MapType {
	mt := MapType(strings.TrimSpace(strings.ToLower(str)))
	switch mt {
	case ID:
		fallthrough
	case I8:
		fallthrough
	case I16:
		fallthrough
	case I32:
		fallthrough
	case I64:
		fallthrough
	case VarInt:
		fallthrough
	case UVarInt:
		fallthrough
	case U8:
		fallthrough
	case U16:
		fallthrough
	case U32:
		fallthrough
	case U64:
		fallthrough
	case BitFlag8:
		fallthrough
	case BitFlag16:
		fallthrough
	case BitFlag32:
		fallthrough
	case BitFlag64:
		fallthrough
	case BitFlagEq8:
		fallthrough
	case BitFlagEq16:
		fallthrough
	case BitFlagEq32:
		fallthrough
	case BitFlagEq64:
		fallthrough
	case BitFlagVal:
		fallthrough
	case F32:
		fallthrough
	case F64:
		fallthrough
	case Len8:
		fallthrough
	case Len16:
		fallthrough
	case Len32:
		fallthrough
	case Len64:
		fallthrough
	case LenVar:
		fallthrough
	case LenData:
		fallthrough
	case FixedData:
		fallthrough
	case Data:
		fallthrough
	case NullTermStr:
		fallthrough
	case LenStr:
		fallthrough
	case Array:
		fallthrough
	case Struct:
		fallthrough
	case StructField:
		fallthrough
	case StructPad:
		fallthrough
	case Magic:
		fallthrough
	case End:
		return mt
	default:
		return Arg
	}
}

func goType(mt MapType) string {
	switch mt {
	case I8:
		return "int8"
	case I16:
		return "int16"
	case I32:
		return "int32"
	case I64:
		return "int64"
	case VarInt:
		return "int64"
	case UVarInt:
		return "uint64"
	case U8:
		return "uint8"
	case U16:
		return "uint16"
	case U32:
		return "uint32"
	case U64:
		return "uint64"
	case BitFlag8:
		return "uint8"
	case BitFlag16:
		return "uint16"
	case BitFlag32:
		return "uint32"
	case BitFlag64:
		return "uint64"
	case BitFlagEq8:
		return "uint8"
	case BitFlagEq16:
		return "uint16"
	case BitFlagEq32:
		return "uint32"
	case BitFlagEq64:
		return "uint64"
	case F32:
		return "float32"
	case F64:
		return "float64"
	case Len8:
		return "uint8"
	case Len16:
		return "uint16"
	case Len32:
		return "uint32"
	case Len64:
		return "uint64"
	case LenVar:
		return "uint64"
	case LenData:
		fallthrough
	case FixedData:
		fallthrough
	case Data:
		return "[]byte"
	case NullTermStr:
		fallthrough
	case LenStr:
		return "string"
	case StructPad:
		return "[]byte"
	case BitFlagVal:
		fallthrough
	case Array:
		fallthrough
	case Struct:
		fallthrough
	case StructField:
		fallthrough
	case Magic:
		fallthrough
	case End:
		return ""
	default:
		panic("No assigned Go type for MapType: " + mt.String())
	}
}

func elemMapper(mt MapType) string {
	switch mt {
	case Magic:
		return "bin.MagicNumber"
	case I8:
		fallthrough
	case I16:
		fallthrough
	case I32:
		fallthrough
	case I64:
		fallthrough
	case U8:
		fallthrough
	case U16:
		fallthrough
	case U32:
		fallthrough
	case U64:
		return "bin.Int"
	case VarInt:
		return "bin.Varint"
	case UVarInt:
		return "bin.Uvarint"
	case BitFlag8:
		fallthrough
	case BitFlag16:
		fallthrough
	case BitFlag32:
		fallthrough
	case BitFlag64:
		fallthrough
	case BitFlagEq8:
		fallthrough
	case BitFlagEq16:
		fallthrough
	case BitFlagEq32:
		fallthrough
	case BitFlagEq64:
		return "bin.Int"
	case F32:
		fallthrough
	case F64:
		return "bin.Float"
	case Len8:
		fallthrough
	case Len16:
		fallthrough
	case Len32:
		fallthrough
	case Len64:
		return "bin.Int"
	case LenVar:
		return "bin.Uvarint"
	case LenData:
		fallthrough
	case FixedData:
		fallthrough
	case Data:
		return "bin.Byte"
	case NullTermStr:
		return "bin.NullTermString"
	default:
		return "(Error: invalid element type)"
	}
}

var (
	stripRegex = regexp.MustCompile(`(\s|[^A-Za-z0-9_])`)
)

func identifier(id string) string {
	id = stripRegex.ReplaceAllString(id, "")
	if len(id) == 0 {
		panic("Empty/blank identifier")
	}
	idCap := ([]rune(id))[0]
	prefix := strings.ToTitle(string(idCap))
	return string(append([]rune(prefix), ([]rune(id))[1:]...))
}

type structMapping struct {
	TypeName string
	Fields   []fieldMapping
	BitFlags []bitFlags
	Package  string
}

func (s *structMapping) HasFieldNamed(name string) bool {
	name = identifier(name)
	for _, f := range s.Fields {
		if identifier(f.Name) == name {
			return true
		}
	}
	return false
}

type fieldMapping struct {
	Name          string
	FieldType     string
	BinType       string
	LenField      string
	FixedLen      uint64
	ElementMapper string
	MagicNumber   []byte
	Struct        structMapping
}

func (m *fieldMapping) MagicNumberValues() string {
	lenMagic := len(m.MagicNumber)
	if lenMagic == 0 {
		return ""
	}
	vals := make([]string, lenMagic)
	for i := 0; i < lenMagic; i++ {
		vals[i] = "0x" + strconv.FormatUint(uint64(m.MagicNumber[i]), 16)
	}
	return strings.Join(vals, ", ")
}

type bitFlags struct {
	Field         string
	FlagID        string
	FlagVal       string
	BitFlagEquals bool
}
