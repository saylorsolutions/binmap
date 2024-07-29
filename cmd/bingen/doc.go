/*
Package main provides the means of generating non-trivial binary serialization and deserialization code from a mapping DSL.

# Format Example

Here's an example for the ICO file type header.

	// This is a bingen mapping file for the ICO file header.
	// Source: https://en.wikipedia.org/wiki/ICO_(file_format)
	id icoHeader
	    field reserved pad 2 // This should always be zero
	    field imageType bf16
	        bf icon 1 // Indicates that this ICO contains an icon
	        bf cursor 2 // Cursors can also be conveyed with this file type.
	    field numImages l16
	    field dirEntries ary numImages struct iconDirEntry
	        field width u8
	        field height u8
	        field colors u8
	        field reserved pad 1 // This should always be zero
	        field colorPlanes u16 // Hotspot X coord with CUR format.
	        field bitsPerPixel u16 // Hotspot Y coord with CUR format.
	        field imageSize u32
	        field imageOffset u32
	    end
	    field imagePayloads data // Use iconDirEntry.imageOffset - 22 to get the correct starting offset in this slice.
	end

Note that line comments (with '//') can be used to document a serialization standard beyond the code.

For the most part binary fields follow this format.

	field NAME TYPE [ARG]

There are exceptions to this as shown with bitflag values (bf) above.
This is necessary because the format is intended to be unstructured, and sequentially parseable from start to finish.
Without field and bf interruption tokens, the format could have a lot of ambiguity.

The ARG value is only specified when needed, and may be used in different contexts.
For example, take a look at the array specification above.

	field dirEntries ary numImages struct iconDirEntry

This follows 'field NAME TYPE' and used ARG to specify the field holding the array length (for fixed length fields like 'fdata' can just specify the size).
For arrays, we must also specify the type of the elements, in this case a struct with its own fields.
The numImages field is specified with 'l16' to more clearly express its role in the format, but otherwise isn't treated any differently than 'u16'.

With that said, it should be clear that the format can tolerate whitespace, indentation, blank lines, etc.

# Basic Mapping Types

	id		This specifies the top level structure.
	i8		A signed, 8-bit integer field.
	i16		A signed, 16-bit integer field.
	i32		A signed, 32-bit integer field.
	i64		A signed, 64-bit integer field.
	u8		An unsigned, 8-bit integer field.
	u16		An unsigned, 16-bit integer field.
	u32		An unsigned, 32-bit integer field.
	u64		An unsigned, 64-bit integer field.
	l8		An unsigned, 8-bit integer field indicating length.
	l16		An unsigned, 16-bit integer field indicating length.
	l32		An unsigned, 32-bit integer field indicating length.
	l64		An unsigned, 64-bit integer field indicating length.
	vi		Variable length integer.
	uvi		Unsigned variable length integer.
	lv		Variable length integer indicating length.
	f32		A 32-bit floating-point field.
	f64		A 64-bit floating-point field.
	bf8		8-bit bitfield.
	bf16		16-bit bitfield.
	bf32		32-bit bitfield.
	bf64		64-bit bitfield.
	bf		A specific bitfield value, identified with a name and value.

# Structured Mapping Types

	ldata		A variable length byte array. Requires an ARG specifying a field holding the length of the byte array.
	fdata		A fixed length byte array. Requires an ARG specifying the size.
	data		A variable length byte array. This uses Remaining for mapping, so is likely only suitable for the end of a file.
	ntstr		A null-terminated string. In Go code, this will be represented as a normal string type without a null terminator.
	lstr		A variable length string. Requires an ARG specifying a field holding the length of the string.
	ary		An array value. Requires an ARG for the field holding the length, and an additional ARG for the element type (see example above).
	struct		A structured type with fields. This can make it easier to express repeated sets of values.
	field		A field in a struct.
	pad		Padding bytes that may be safely discarded. Requires an ARG specifying the number of padding bytes (see reserved fields above).
	end		Ends a struct definition, including the top level structure.
*/
package main
