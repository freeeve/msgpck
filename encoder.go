package msgpck

import (
	"encoding/binary"
	"math"
	"unsafe"
)

// Encoder writes MessagePack data to a byte buffer.
type Encoder struct {
	buf []byte
	pos int
}

// NewEncoder creates a new Encoder with the given initial capacity.
func NewEncoder(capacity int) *Encoder {
	return &Encoder{
		buf: make([]byte, 0, capacity),
	}
}

// NewEncoderBuffer creates an Encoder that writes to an existing buffer.
// The buffer will be grown as needed.
func NewEncoderBuffer(buf []byte) *Encoder {
	return &Encoder{
		buf: buf[:0], // reset length but keep capacity
	}
}

// Reset resets the encoder for reuse.
func (e *Encoder) Reset() {
	e.buf = e.buf[:0]
	e.pos = 0
}

// Bytes returns the encoded bytes.
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// Len returns the length of encoded data.
func (e *Encoder) Len() int {
	return len(e.buf)
}

// grow ensures there's space for n more bytes
func (e *Encoder) grow(n int) {
	if cap(e.buf)-len(e.buf) >= n {
		return
	}
	// Double capacity or add n, whichever is larger
	newCap := cap(e.buf) * 2
	if newCap < len(e.buf)+n {
		newCap = len(e.buf) + n
	}
	newBuf := make([]byte, len(e.buf), newCap)
	copy(newBuf, e.buf)
	e.buf = newBuf
}

// writeByte writes a single byte
func (e *Encoder) writeByte(b byte) {
	e.grow(1)
	e.buf = append(e.buf, b)
}

// writeBytes writes multiple bytes
func (e *Encoder) writeBytes(b []byte) {
	e.grow(len(b))
	e.buf = append(e.buf, b...)
}

// writeUint16 writes a big-endian uint16
func (e *Encoder) writeUint16(v uint16) {
	e.grow(2)
	e.buf = append(e.buf, 0, 0)
	binary.BigEndian.PutUint16(e.buf[len(e.buf)-2:], v)
}

// writeUint32 writes a big-endian uint32
func (e *Encoder) writeUint32(v uint32) {
	e.grow(4)
	e.buf = append(e.buf, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(e.buf[len(e.buf)-4:], v)
}

// writeUint64 writes a big-endian uint64
func (e *Encoder) writeUint64(v uint64) {
	e.grow(8)
	e.buf = append(e.buf, 0, 0, 0, 0, 0, 0, 0, 0)
	binary.BigEndian.PutUint64(e.buf[len(e.buf)-8:], v)
}

// EncodeNil writes a nil value
func (e *Encoder) EncodeNil() {
	e.writeByte(FormatNil)
}

// EncodeBool writes a boolean value
func (e *Encoder) EncodeBool(v bool) {
	if v {
		e.writeByte(FormatTrue)
	} else {
		e.writeByte(FormatFalse)
	}
}

// EncodeInt writes an int64 using the most compact format
func (e *Encoder) EncodeInt(v int64) {
	if v >= 0 {
		e.EncodeUint(uint64(v))
		return
	}

	// Negative values
	if v >= -32 {
		// Negative fixint
		e.writeByte(byte(v))
	} else if v >= -128 {
		e.writeByte(FormatInt8)
		e.writeByte(byte(v))
	} else if v >= -32768 {
		e.writeByte(FormatInt16)
		e.writeUint16(uint16(v))
	} else if v >= -2147483648 {
		e.writeByte(FormatInt32)
		e.writeUint32(uint32(v))
	} else {
		e.writeByte(FormatInt64)
		e.writeUint64(uint64(v))
	}
}

// EncodeUint writes a uint64 using the most compact format
func (e *Encoder) EncodeUint(v uint64) {
	if v <= 127 {
		// Positive fixint
		e.writeByte(byte(v))
	} else if v <= 255 {
		e.writeByte(FormatUint8)
		e.writeByte(byte(v))
	} else if v <= 65535 {
		e.writeByte(FormatUint16)
		e.writeUint16(uint16(v))
	} else if v <= 4294967295 {
		e.writeByte(FormatUint32)
		e.writeUint32(uint32(v))
	} else {
		e.writeByte(FormatUint64)
		e.writeUint64(v)
	}
}

// EncodeFloat32 writes a float32 value
func (e *Encoder) EncodeFloat32(v float32) {
	e.writeByte(FormatFloat32)
	e.writeUint32(math.Float32bits(v))
}

// EncodeFloat64 writes a float64 value
func (e *Encoder) EncodeFloat64(v float64) {
	e.writeByte(FormatFloat64)
	e.writeUint64(math.Float64bits(v))
}

// EncodeString writes a string value (zero-copy using unsafe)
func (e *Encoder) EncodeString(v string) {
	length := len(v)

	if length <= 31 {
		e.writeByte(FixstrPrefix | byte(length))
	} else if length <= 255 {
		e.writeByte(FormatStr8)
		e.writeByte(byte(length))
	} else if length <= 65535 {
		e.writeByte(FormatStr16)
		e.writeUint16(uint16(length))
	} else {
		e.writeByte(FormatStr32)
		e.writeUint32(uint32(length))
	}
	// Zero-copy string to bytes using unsafe
	e.writeBytes(unsafe.Slice(unsafe.StringData(v), length))
}

// EncodeStringBytes writes a string from bytes
func (e *Encoder) EncodeStringBytes(v []byte) {
	length := len(v)

	if length <= 31 {
		// Fixstr
		e.writeByte(FixstrPrefix | byte(length))
	} else if length <= 255 {
		e.writeByte(FormatStr8)
		e.writeByte(byte(length))
	} else if length <= 65535 {
		e.writeByte(FormatStr16)
		e.writeUint16(uint16(length))
	} else {
		e.writeByte(FormatStr32)
		e.writeUint32(uint32(length))
	}
	e.writeBytes(v)
}

// EncodeBinary writes binary data
func (e *Encoder) EncodeBinary(v []byte) {
	length := len(v)

	if length <= 255 {
		e.writeByte(FormatBin8)
		e.writeByte(byte(length))
	} else if length <= 65535 {
		e.writeByte(FormatBin16)
		e.writeUint16(uint16(length))
	} else {
		e.writeByte(FormatBin32)
		e.writeUint32(uint32(length))
	}
	e.writeBytes(v)
}

// EncodeArrayHeader writes the header for an array of the given length.
// Call this, then encode each element.
func (e *Encoder) EncodeArrayHeader(length int) {
	if length <= 15 {
		// Fixarray
		e.writeByte(FixarrayPrefix | byte(length))
	} else if length <= 65535 {
		e.writeByte(FormatArray16)
		e.writeUint16(uint16(length))
	} else {
		e.writeByte(FormatArray32)
		e.writeUint32(uint32(length))
	}
}

// EncodeMapHeader writes the header for a map of the given length.
// Call this, then encode each key-value pair.
func (e *Encoder) EncodeMapHeader(length int) {
	if length <= 15 {
		// Fixmap
		e.writeByte(FixmapPrefix | byte(length))
	} else if length <= 65535 {
		e.writeByte(FormatMap16)
		e.writeUint16(uint16(length))
	} else {
		e.writeByte(FormatMap32)
		e.writeUint32(uint32(length))
	}
}

// EncodeExt writes extension data
func (e *Encoder) EncodeExt(extType int8, data []byte) {
	length := len(data)

	switch length {
	case 1:
		e.writeByte(FormatFixExt1)
	case 2:
		e.writeByte(FormatFixExt2)
	case 4:
		e.writeByte(FormatFixExt4)
	case 8:
		e.writeByte(FormatFixExt8)
	case 16:
		e.writeByte(FormatFixExt16)
	default:
		if length <= 255 {
			e.writeByte(FormatExt8)
			e.writeByte(byte(length))
		} else if length <= 65535 {
			e.writeByte(FormatExt16)
			e.writeUint16(uint16(length))
		} else {
			e.writeByte(FormatExt32)
			e.writeUint32(uint32(length))
		}
	}
	e.writeByte(byte(extType))
	e.writeBytes(data)
}

// EncodeValue writes a Value
func (e *Encoder) EncodeValue(v *Value) {
	switch v.Type {
	case TypeNil:
		e.EncodeNil()
	case TypeBool:
		e.EncodeBool(v.Bool)
	case TypeInt:
		e.EncodeInt(v.Int)
	case TypeUint:
		e.EncodeUint(v.Uint)
	case TypeFloat32:
		e.EncodeFloat32(v.Float32)
	case TypeFloat64:
		e.EncodeFloat64(v.Float64)
	case TypeString:
		e.EncodeStringBytes(v.Bytes)
	case TypeBinary:
		e.EncodeBinary(v.Bytes)
	case TypeArray:
		e.EncodeArrayHeader(len(v.Array))
		for i := range v.Array {
			e.EncodeValue(&v.Array[i])
		}
	case TypeMap:
		e.EncodeMapHeader(len(v.Map))
		for i := range v.Map {
			e.EncodeStringBytes(v.Map[i].Key)
			e.EncodeValue(&v.Map[i].Value)
		}
	case TypeExt:
		e.EncodeExt(v.Ext.Type, v.Ext.Data)
	}
}
