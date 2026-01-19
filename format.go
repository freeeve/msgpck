package msgpck

// MessagePack format bytes
// See: https://github.com/msgpack/msgpack/blob/master/spec.md
const (
	// Nil
	formatNil byte = 0xc0

	// Bool
	formatFalse byte = 0xc2
	formatTrue  byte = 0xc3

	// Int formats
	formatUint8  byte = 0xcc
	formatUint16 byte = 0xcd
	formatUint32 byte = 0xce
	formatUint64 byte = 0xcf
	formatInt8   byte = 0xd0
	formatInt16  byte = 0xd1
	formatInt32  byte = 0xd2
	formatInt64  byte = 0xd3

	// Float formats
	formatFloat32 byte = 0xca
	formatFloat64 byte = 0xcb

	// String formats
	formatStr8  byte = 0xd9
	formatStr16 byte = 0xda
	formatStr32 byte = 0xdb

	// Binary formats
	formatBin8  byte = 0xc4
	formatBin16 byte = 0xc5
	formatBin32 byte = 0xc6

	// Array formats
	formatArray16 byte = 0xdc
	formatArray32 byte = 0xdd

	// Map formats
	formatMap16 byte = 0xde
	formatMap32 byte = 0xdf

	// Ext formats
	formatFixExt1  byte = 0xd4
	formatFixExt2  byte = 0xd5
	formatFixExt4  byte = 0xd6
	formatFixExt8  byte = 0xd7
	formatFixExt16 byte = 0xd8
	formatExt8     byte = 0xc7
	formatExt16    byte = 0xc8
	formatExt32    byte = 0xc9
)

// Format masks and ranges
const (
	// Positive fixint: 0xxxxxxx (0x00 - 0x7f)
	posFixintMask byte = 0x80
	posFixintMax  byte = 0x7f

	// Negative fixint: 111xxxxx (0xe0 - 0xff)
	negFixintMin byte = 0xe0

	// Fixmap: 1000xxxx (0x80 - 0x8f)
	fixmapMask   byte = 0xf0
	fixmapPrefix byte = 0x80
	fixmapMax    byte = 0x0f

	// Fixarray: 1001xxxx (0x90 - 0x9f)
	fixarrayMask   byte = 0xf0
	fixarrayPrefix byte = 0x90
	fixarrayMax    byte = 0x0f

	// Fixstr: 101xxxxx (0xa0 - 0xbf)
	fixstrMask   byte = 0xe0
	fixstrPrefix byte = 0xa0
	fixstrMax    byte = 0x1f
)

// isPositiveFixint returns true if b is a positive fixint (0x00-0x7f)
func isPositiveFixint(b byte) bool {
	return b&posFixintMask == 0
}

// isNegativeFixint returns true if b is a negative fixint (0xe0-0xff)
func isNegativeFixint(b byte) bool {
	return b >= negFixintMin
}

// isFixmap returns true if b is a fixmap (0x80-0x8f)
func isFixmap(b byte) bool {
	return b&fixmapMask == fixmapPrefix
}

// isFixarray returns true if b is a fixarray (0x90-0x9f)
func isFixarray(b byte) bool {
	return b&fixarrayMask == fixarrayPrefix
}

// isFixstr returns true if b is a fixstr (0xa0-0xbf)
func isFixstr(b byte) bool {
	return b&fixstrMask == fixstrPrefix
}

// fixmapLen returns the length encoded in a fixmap byte
func fixmapLen(b byte) int {
	return int(b & fixmapMax)
}

// fixarrayLen returns the length encoded in a fixarray byte
func fixarrayLen(b byte) int {
	return int(b & fixarrayMax)
}

// fixstrLen returns the length encoded in a fixstr byte
func fixstrLen(b byte) int {
	return int(b & fixstrMax)
}
