package msgpck

import (
	"encoding/binary"
	"time"
)

// TimestampExtType is the extension type for msgpack timestamps.
const TimestampExtType int8 = -1

// EncodeTimestamp encodes a time.Time as a msgpack timestamp extension.
// Uses the most compact format that can represent the timestamp:
//   - Timestamp 32: seconds only (no nanoseconds), fits in uint32
//   - Timestamp 64: nanoseconds + seconds fits in 34-bit unsigned
//   - Timestamp 96: full range with 64-bit signed seconds
//
// The time is converted to UTC before encoding.
func (e *Encoder) EncodeTimestamp(t time.Time) {
	t = t.UTC()
	sec := t.Unix()
	nsec := uint32(t.Nanosecond())

	if nsec == 0 && sec >= 0 && sec <= 0xFFFFFFFF {
		// Timestamp 32: fixext 4, type -1, 4 bytes seconds
		e.writeByte(formatFixExt4)
		e.writeByte(0xff) // -1 as unsigned byte
		e.writeUint32(uint32(sec))
	} else if sec >= 0 && sec <= 0x3FFFFFFFF {
		// Timestamp 64: fixext 8, type -1, 8 bytes (30-bit nsec + 34-bit sec)
		// Upper 30 bits: nanoseconds, lower 34 bits: seconds
		val := (uint64(nsec) << 34) | uint64(sec)
		e.writeByte(formatFixExt8)
		e.writeByte(0xff) // -1 as unsigned byte
		e.writeUint64(val)
	} else {
		// Timestamp 96: ext 8, length 12, type -1, 4 bytes nsec + 8 bytes sec
		e.writeByte(formatExt8)
		e.writeByte(12)
		e.writeByte(0xff) // -1 as unsigned byte
		e.writeUint32(nsec)
		e.writeUint64(uint64(sec))
	}
}

// DecodeTimestamp decodes a msgpack timestamp extension to time.Time.
// Returns the time in UTC.
// Returns ErrTypeMismatch if the value is not a timestamp extension.
func (d *Decoder) DecodeTimestamp() (time.Time, error) {
	format, err := d.readByte()
	if err != nil {
		return time.Time{}, err
	}

	var dataLen int
	switch format {
	case formatFixExt4:
		dataLen = 4
	case formatFixExt8:
		dataLen = 8
	case formatExt8:
		n, err := d.readUint8()
		if err != nil {
			return time.Time{}, err
		}
		dataLen = int(n)
		if dataLen != 12 {
			return time.Time{}, ErrTypeMismatch
		}
	default:
		return time.Time{}, ErrTypeMismatch
	}

	// Read extension type
	extType, err := d.readInt8()
	if err != nil {
		return time.Time{}, err
	}
	if extType != TimestampExtType {
		return time.Time{}, ErrTypeMismatch
	}

	switch dataLen {
	case 4:
		// Timestamp 32: 4 bytes seconds as uint32
		sec, err := d.readUint32()
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(sec), 0).UTC(), nil

	case 8:
		// Timestamp 64: 8 bytes (30-bit nsec + 34-bit sec)
		val, err := d.readUint64()
		if err != nil {
			return time.Time{}, err
		}
		nsec := int64(val >> 34)
		sec := int64(val & 0x3FFFFFFFF)
		return time.Unix(sec, nsec).UTC(), nil

	case 12:
		// Timestamp 96: 4 bytes nsec + 8 bytes sec
		nsec, err := d.readUint32()
		if err != nil {
			return time.Time{}, err
		}
		sec, err := d.readInt64()
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(sec, int64(nsec)).UTC(), nil

	default:
		return time.Time{}, ErrTypeMismatch
	}
}

// MarshalTimestamp encodes a time.Time to msgpack timestamp bytes.
// Returns a copy of the encoded bytes (safe to retain).
// The time is converted to UTC before encoding.
func MarshalTimestamp(t time.Time) []byte {
	e := encoderPool.Get().(*Encoder)
	e.Reset()
	e.EncodeTimestamp(t)
	result := make([]byte, len(e.buf))
	copy(result, e.buf)
	encoderPool.Put(e)
	return result
}

// UnmarshalTimestamp decodes msgpack timestamp bytes to time.Time.
// Returns the time in UTC.
// If the input has no timezone info, it is treated as UTC.
func UnmarshalTimestamp(data []byte) (time.Time, error) {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)
	t, err := d.DecodeTimestamp()
	decoderPool.Put(d)
	return t, err
}

// IsTimestamp checks if an Ext value is a timestamp.
func IsTimestamp(ext Ext) bool {
	return ext.Type == TimestampExtType
}

// ExtToTimestamp converts an Ext value to time.Time.
// Returns ErrTypeMismatch if the extension type is not a timestamp.
// Returns the time in UTC.
func ExtToTimestamp(ext Ext) (time.Time, error) {
	if ext.Type != TimestampExtType {
		return time.Time{}, ErrTypeMismatch
	}

	switch len(ext.Data) {
	case 4:
		// Timestamp 32
		sec := binary.BigEndian.Uint32(ext.Data)
		return time.Unix(int64(sec), 0).UTC(), nil

	case 8:
		// Timestamp 64
		val := binary.BigEndian.Uint64(ext.Data)
		nsec := int64(val >> 34)
		sec := int64(val & 0x3FFFFFFFF)
		return time.Unix(sec, nsec).UTC(), nil

	case 12:
		// Timestamp 96
		nsec := binary.BigEndian.Uint32(ext.Data[:4])
		sec := int64(binary.BigEndian.Uint64(ext.Data[4:]))
		return time.Unix(sec, int64(nsec)).UTC(), nil

	default:
		return time.Time{}, ErrTypeMismatch
	}
}
