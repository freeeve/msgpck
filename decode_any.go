package msgpck

import (
	"sync"
	"unsafe"
)

// Decoder pool for zero-alloc Unmarshal
var decoderPool = sync.Pool{
	New: func() any { return &Decoder{cfg: DefaultConfig()} },
}


// DecodeAny decodes a MessagePack value directly into a Go native type.
// Returns: nil, bool, int64, uint64, float64, string, []byte, []any, map[string]any
// This is optimized to avoid intermediate Value structs.
func (d *Decoder) DecodeAny() (any, error) {
	format, err := d.readByte()
	if err != nil {
		return nil, err
	}
	return d.decodeAnyValue(format)
}

// decodeAnyValue decodes directly to any, skipping intermediate Value struct
func (d *Decoder) decodeAnyValue(format byte) (any, error) {
	// Positive fixint: 0xxxxxxx
	if IsPositiveFixint(format) {
		return int64(format), nil
	}

	// Negative fixint: 111xxxxx
	if IsNegativeFixint(format) {
		return int64(int8(format)), nil
	}

	// Fixmap: 1000xxxx
	if IsFixmap(format) {
		return d.decodeMapAny(FixmapLen(format))
	}

	// Fixarray: 1001xxxx
	if IsFixarray(format) {
		return d.decodeArrayAny(FixarrayLen(format))
	}

	// Fixstr: 101xxxxx
	if IsFixstr(format) {
		return d.decodeStringAny(FixstrLen(format))
	}

	switch format {
	case FormatNil:
		return nil, nil

	case FormatFalse:
		return false, nil
	case FormatTrue:
		return true, nil

	case FormatUint8:
		v, err := d.readUint8()
		return int64(v), err
	case FormatUint16:
		v, err := d.readUint16()
		return int64(v), err
	case FormatUint32:
		v, err := d.readUint32()
		return int64(v), err
	case FormatUint64:
		v, err := d.readUint64()
		// Return as uint64 only if it overflows int64
		if v > 9223372036854775807 {
			return v, err
		}
		return int64(v), err

	case FormatInt8:
		v, err := d.readInt8()
		return int64(v), err
	case FormatInt16:
		v, err := d.readInt16()
		return int64(v), err
	case FormatInt32:
		v, err := d.readInt32()
		return int64(v), err
	case FormatInt64:
		v, err := d.readInt64()
		return v, err

	case FormatFloat32:
		v, err := d.readFloat32()
		return float64(v), err // promote to float64 for consistency
	case FormatFloat64:
		v, err := d.readFloat64()
		return v, err

	case FormatStr8:
		length, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return d.decodeStringAny(int(length))
	case FormatStr16:
		length, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return d.decodeStringAny(int(length))
	case FormatStr32:
		length, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return d.decodeStringAny(int(length))

	case FormatBin8:
		length, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return d.decodeBinaryAny(int(length))
	case FormatBin16:
		length, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return d.decodeBinaryAny(int(length))
	case FormatBin32:
		length, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return d.decodeBinaryAny(int(length))

	case FormatArray16:
		length, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return d.decodeArrayAny(int(length))
	case FormatArray32:
		length, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return d.decodeArrayAny(int(length))

	case FormatMap16:
		length, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return d.decodeMapAny(int(length))
	case FormatMap32:
		length, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return d.decodeMapAny(int(length))

	default:
		return nil, ErrInvalidFormat
	}
}

// decodeStringAny decodes a string directly (zero-copy to string using unsafe)
func (d *Decoder) decodeStringAny(length int) (string, error) {
	if err := d.validateStringLen(length); err != nil {
		return "", err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	// Zero-copy bytes to string using unsafe
	return unsafe.String(unsafe.SliceData(bytes), len(bytes)), nil
}


// decodeBinaryAny decodes binary data (returns copy for safety)
func (d *Decoder) decodeBinaryAny(length int) ([]byte, error) {
	if err := d.validateBinaryLen(length); err != nil {
		return nil, err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return nil, err
	}
	// Copy for safety since source buffer may be reused
	cp := make([]byte, length)
	copy(cp, bytes)
	return cp, nil
}

// decodeArrayAny decodes an array directly to []any
func (d *Decoder) decodeArrayAny(length int) ([]any, error) {
	if err := d.validateArrayLen(length); err != nil {
		return nil, err
	}
	if err := d.enterContainer(); err != nil {
		return nil, err
	}
	defer d.leaveContainer()

	arr := make([]any, length)
	for i := 0; i < length; i++ {
		v, err := d.DecodeAny()
		if err != nil {
			return nil, err
		}
		arr[i] = v
	}
	return arr, nil
}

// decodeMapAny decodes a map directly to map[string]any
func (d *Decoder) decodeMapAny(length int) (map[string]any, error) {
	if err := d.validateMapLen(length); err != nil {
		return nil, err
	}
	if err := d.enterContainer(); err != nil {
		return nil, err
	}
	defer d.leaveContainer()

	m := make(map[string]any, length)
	for i := 0; i < length; i++ {
		// Read key format
		keyFormat, err := d.readByte()
		if err != nil {
			return nil, err
		}

		// Decode key as string
		var key string
		if IsFixstr(keyFormat) {
			key, err = d.decodeStringAny(FixstrLen(keyFormat))
		} else if keyFormat == FormatStr8 {
			length, err := d.readUint8()
			if err != nil {
				return nil, err
			}
			key, err = d.decodeStringAny(int(length))
		} else if keyFormat == FormatStr16 {
			length, err := d.readUint16()
			if err != nil {
				return nil, err
			}
			key, err = d.decodeStringAny(int(length))
		} else if keyFormat == FormatStr32 {
			length, err := d.readUint32()
			if err != nil {
				return nil, err
			}
			key, err = d.decodeStringAny(int(length))
		} else {
			return nil, ErrInvalidFormat // non-string key
		}
		if err != nil {
			return nil, err
		}

		// Read value
		val, err := d.DecodeAny()
		if err != nil {
			return nil, err
		}

		m[key] = val
	}
	return m, nil
}

// valueToAny converts a Value to a Go native type (for compatibility)
func valueToAny(v *Value) any {
	switch v.Type {
	case TypeNil:
		return nil
	case TypeBool:
		return v.Bool
	case TypeInt:
		return v.Int
	case TypeUint:
		return v.Uint
	case TypeFloat32:
		return float64(v.Float32)
	case TypeFloat64:
		return v.Float64
	case TypeString:
		return unsafe.String(unsafe.SliceData(v.Bytes), len(v.Bytes))
	case TypeBinary:
		cp := make([]byte, len(v.Bytes))
		copy(cp, v.Bytes)
		return cp
	case TypeArray:
		arr := make([]any, len(v.Array))
		for i := range v.Array {
			arr[i] = valueToAny(&v.Array[i])
		}
		return arr
	case TypeMap:
		m := make(map[string]any, len(v.Map))
		for i := range v.Map {
			key := unsafe.String(unsafe.SliceData(v.Map[i].Key), len(v.Map[i].Key))
			m[key] = valueToAny(&v.Map[i].Value)
		}
		return m
	case TypeExt:
		return v.Ext
	default:
		return nil
	}
}

// Unmarshal decodes msgpack data into any.
// Uses pooled decoder for zero decoder allocation.
func Unmarshal(data []byte) (any, error) {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)
	result, err := d.DecodeAny()
	decoderPool.Put(d)
	return result, err
}

// UnmarshalWithConfig decodes msgpack data with custom config.
func UnmarshalWithConfig(data []byte, cfg Config) (any, error) {
	d := NewDecoderWithConfig(data, cfg)
	return d.DecodeAny()
}
