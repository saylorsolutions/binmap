package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNoTokens      = errors.New("no tokens found")
	ErrUnexpected    = errors.New("unexpected token")
	ErrEOF           = errors.New("unexpected end of input")
	ErrRefBeforeRead = errors.New("field referenced before reading")
	ErrInvalidMagic  = errors.New("invalid magic number constant")
	magicNumberRegex = regexp.MustCompile(`^[0-9a-fA-F]+$`)
)

func tokenIs(t *token, typ MapType, others ...MapType) bool {
	if t == nil {
		return false
	}
	if t.typ == typ {
		return true
	}
	for _, other := range others {
		if t.typ == other {
			return true
		}
	}
	return false
}

func tokenNot(t *token, types ...MapType) bool {
	if t == nil {
		return true
	}
	for _, typ := range types {
		if t.typ == typ {
			return false
		}
	}
	return true
}

func unexpectedToken(got *token, wanted ...MapType) error {
	var oneOf strings.Builder
	for i, want := range wanted {
		if i > 0 {
			oneOf.WriteString(", ")
		}
		oneOf.WriteString(string(want))
	}
	if len(wanted) > 0 {
		return fmt.Errorf("%w: got %s (%s) at line %d col %d, wanted one of %s", ErrUnexpected,
			string(got.typ), got.tok,
			got.line, got.col,
			oneOf.String(),
		)
	} else {
		return fmt.Errorf("%w: got %s (%s) at line %d col %d", ErrUnexpected,
			string(got.typ), got.tok,
			got.line, got.col,
		)
	}
}

func refBeforeRead(fieldName string) error {
	return fmt.Errorf("%w: %s", ErrRefBeforeRead, fieldName)
}

func parseMapping(r io.Reader) (*structMapping, error) {
	tokens, err := lex(r)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, ErrNoTokens
	}
	next := tokenIterator(tokens)
	// I know there's at least one
	id, _ := next()
	if tokenNot(id, ID) {
		return nil, unexpectedToken(id, ID)
	}
	name, ok := next()
	if !ok {
		return nil, ErrEOF
	}
	if tokenNot(name, Arg) {
		return nil, unexpectedToken(name, Arg)
	}
	topLevelStruct := &structMapping{
		TypeName: name.tok,
	}
	if err := parseStructFields(topLevelStruct, next); err != nil {
		return nil, err
	}
	return topLevelStruct, nil
}

func parseStructFields(s *structMapping, next iterator) error {
	for {
		tok, ok := next()
		if !ok {
			return ErrEOF
		}
		switch tok.typ {
		case StructField:
			if err := parseField(s, next); err != nil {
				return err
			}
		case End:
			return nil
		default:
			return unexpectedToken(tok, StructField, End)
		}
	}
}

func parseField(s *structMapping, next iterator) error {
	name, ok := next()
	if !ok {
		return ErrEOF
	}
	if tokenNot(name, Arg) {
		return unexpectedToken(name, Arg)
	}
	field := fieldMapping{
		Name: name.tok,
	}
	if err := mapField(s, &field, next); err != nil {
		return err
	}
	s.Fields = append(s.Fields, field)
	return nil
}

func mapField(s *structMapping, field *fieldMapping, next iterator) error {
	mapping, ok := next()
	if !ok {
		return ErrEOF
	}
	field.FieldType = goType(mapping.typ)
	field.ElementMapper = elemMapper(mapping.typ)
	field.BinType = mapping.typ.String()
	switch {
	case tokenIs(mapping, I8, I16, I32, I64, U8, U16, U32, U64, VarInt, UVarInt, F32, F64, Len8, Len16, Len32, Len64, LenVar, Data, NullTermStr):
	case tokenIs(mapping, LenData, LenStr):
		lenField, ok := next()
		if !ok {
			return ErrEOF
		}
		if tokenNot(lenField, Arg) {
			return unexpectedToken(lenField, Arg)
		}
		if !s.HasFieldNamed(lenField.tok) {
			return refBeforeRead(lenField.tok)
		}
		field.LenField = lenField.tok
	case tokenIs(mapping, Magic):
		magicVal, ok := next()
		if !ok {
			return ErrEOF
		}
		if tokenNot(magicVal, Arg) {
			return unexpectedToken(magicVal, Arg)
		}
		if !strings.HasPrefix(magicVal.tok, "0x") {
			return fmt.Errorf("%w: expected leading '0x' for magic number constant '%s'", ErrInvalidMagic, magicVal.tok)
		}
		valueStr := strings.TrimPrefix(magicVal.tok, "0x")
		if !magicNumberRegex.MatchString(valueStr) {
			return fmt.Errorf("%w: magic number constant '%s' has invalid characters", ErrInvalidMagic, magicVal.tok)
		}
		valueRunes := []rune(valueStr)
		if len(valueRunes)%2 != 0 {
			return fmt.Errorf("%w: magic number constant '%s' must have an even number of hex digits", ErrInvalidMagic, magicVal.tok)
		}
		value := make([]byte, len(valueRunes)/2)
		for i := 0; i < len(valueRunes); i += 2 {
			str := string(valueRunes[i : i+2])
			ival, err := strconv.ParseUint(str, 16, 8)
			if err != nil {
				return fmt.Errorf("%w: unable to parse magic number part '%s' as hex number: %v", ErrInvalidMagic, str, err)
			}
			value[i/2] = byte(ival & 0xff)
		}
		field.MagicNumber = value
	case tokenIs(mapping, FixedData, StructPad):
		fixedLen, ok := next()
		if !ok {
			return ErrEOF
		}
		if tokenNot(fixedLen, Arg) {
			return unexpectedToken(fixedLen, Arg)
		}
		size, err := strconv.ParseInt(fixedLen.tok, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid fixed field size string '%s'", fixedLen.tok)
		}
		field.FixedLen = uint64(size)
	case tokenIs(mapping, BitFlag8, BitFlag16, BitFlag32, BitFlag64):
		for {
			// Try to resolve bit flag values.
			bf, ok := next()
			if !ok {
				return ErrEOF
			}
			if tokenNot(bf, BitFlagVal) {
				next(-1)
				break
			}
			name, ok := next()
			if !ok {
				return ErrEOF
			}
			if tokenNot(name, Arg) {
				return unexpectedToken(name, Arg)
			}
			val, ok := next()
			if !ok {
				return ErrEOF
			}
			if tokenNot(val, Arg) {
				return unexpectedToken(val, Arg)
			}
			s.BitFlags = append(s.BitFlags, bitFlags{
				Field:   field.Name,
				FlagID:  name.tok,
				FlagVal: val.tok,
			})
		}
	case tokenIs(mapping, Array):
		aryLen, ok := next()
		if !ok {
			return ErrEOF
		}
		if !s.HasFieldNamed(aryLen.tok) {
			return refBeforeRead(aryLen.tok)
		}
		field.LenField = aryLen.tok
		if err := mapField(s, field, next); err != nil {
			return err
		}
		field.BinType = Array.String()
		field.FieldType = "[]" + field.FieldType
	case tokenIs(mapping, Struct):
		structName, ok := next()
		if !ok {
			return ErrEOF
		}
		if tokenNot(structName, Arg) {
			return unexpectedToken(structName, Arg)
		}
		field.FieldType = identifier(structName.tok)
		field.Struct = structMapping{
			TypeName: identifier(structName.tok),
		}
		if err := parseStructFields(&field.Struct, next); err != nil {
			return err
		}
		field.ElementMapper = identifier(structName.tok + "Mapper")
	default:
		return unexpectedToken(mapping)
	}
	return nil
}
