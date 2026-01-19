package msgpck

// MessagePack format bytes
// See: https://github.com/msgpack/msgpack/blob/master/spec.md
const (
	// Nil
	FormatNil byte = 0xc0

	// Bool
	FormatFalse byte = 0xc2
	FormatTrue  byte = 0xc3

	// Int formats
	FormatUint8  byte = 0xcc
	FormatUint16 byte = 0xcd
	FormatUint32 byte = 0xce
	FormatUint64 byte = 0xcf
	FormatInt8   byte = 0xd0
	FormatInt16  byte = 0xd1
	FormatInt32  byte = 0xd2
	FormatInt64  byte = 0xd3

	// Float formats
	FormatFloat32 byte = 0xca
	FormatFloat64 byte = 0xcb

	// String formats
	FormatStr8  byte = 0xd9
	FormatStr16 byte = 0xda
	FormatStr32 byte = 0xdb

	// Binary formats
	FormatBin8  byte = 0xc4
	FormatBin16 byte = 0xc5
	FormatBin32 byte = 0xc6

	// Array formats
	FormatArray16 byte = 0xdc
	FormatArray32 byte = 0xdd

	// Map formats
	FormatMap16 byte = 0xde
	FormatMap32 byte = 0xdf

	// Ext formats
	FormatFixExt1  byte = 0xd4
	FormatFixExt2  byte = 0xd5
	FormatFixExt4  byte = 0xd6
	FormatFixExt8  byte = 0xd7
	FormatFixExt16 byte = 0xd8
	FormatExt8     byte = 0xc7
	FormatExt16    byte = 0xc8
	FormatExt32    byte = 0xc9
)

// Format masks and ranges
const (
	// Positive fixint: 0xxxxxxx (0x00 - 0x7f)
	PosFixintMask byte = 0x80
	PosFixintMax  byte = 0x7f

	// Negative fixint: 111xxxxx (0xe0 - 0xff)
	NegFixintMin byte = 0xe0

	// Fixmap: 1000xxxx (0x80 - 0x8f)
	FixmapMask   byte = 0xf0
	FixmapPrefix byte = 0x80
	FixmapMax    byte = 0x0f

	// Fixarray: 1001xxxx (0x90 - 0x9f)
	FixarrayMask   byte = 0xf0
	FixarrayPrefix byte = 0x90
	FixarrayMax    byte = 0x0f

	// Fixstr: 101xxxxx (0xa0 - 0xbf)
	FixstrMask   byte = 0xe0
	FixstrPrefix byte = 0xa0
	FixstrMax    byte = 0x1f
)

// IsPositiveFixint returns true if b is a positive fixint (0x00-0x7f)
func IsPositiveFixint(b byte) bool {
	return b&PosFixintMask == 0
}

// IsNegativeFixint returns true if b is a negative fixint (0xe0-0xff)
func IsNegativeFixint(b byte) bool {
	return b >= NegFixintMin
}

// IsFixmap returns true if b is a fixmap (0x80-0x8f)
func IsFixmap(b byte) bool {
	return b&FixmapMask == FixmapPrefix
}

// IsFixarray returns true if b is a fixarray (0x90-0x9f)
func IsFixarray(b byte) bool {
	return b&FixarrayMask == FixarrayPrefix
}

// IsFixstr returns true if b is a fixstr (0xa0-0xbf)
func IsFixstr(b byte) bool {
	return b&FixstrMask == FixstrPrefix
}

// FixmapLen returns the length encoded in a fixmap byte
func FixmapLen(b byte) int {
	return int(b & FixmapMax)
}

// FixarrayLen returns the length encoded in a fixarray byte
func FixarrayLen(b byte) int {
	return int(b & FixarrayMax)
}

// FixstrLen returns the length encoded in a fixstr byte
func FixstrLen(b byte) int {
	return int(b & FixstrMax)
}
