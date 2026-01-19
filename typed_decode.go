package msgpck

import "unsafe"

// UnmarshalMapStringAny decodes msgpack into map[string]any.
// Zero-copy strings when zeroCopy is true.
func UnmarshalMapStringAny(data []byte, zeroCopy bool) (map[string]any, error) {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)
	m, err := decodeMapStringAny(d, zeroCopy)
	decoderPool.Put(d)
	return m, err
}

func decodeMapStringAny(d *Decoder, zeroCopy bool) (map[string]any, error) {
	format, err := d.readByte()
	if err != nil {
		return nil, err
	}

	var mapLen int
	if isFixmap(format) {
		mapLen = fixmapLen(format)
	} else if format == formatMap16 {
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		mapLen = int(n)
	} else if format == formatMap32 {
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		mapLen = int(n)
	} else if format == formatNil {
		return nil, nil
	} else {
		return nil, ErrTypeMismatch
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return nil, err
	}

	m := make(map[string]any, mapLen)
	for i := 0; i < mapLen; i++ {
		// Read key
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return nil, err
		}
		var key string
		if zeroCopy {
			key = unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))
		} else {
			key = string(keyBytes)
		}

		// Read value
		val, err := decodeAnyValue(d, zeroCopy)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return m, nil
}

func decodeAnyValue(d *Decoder, zeroCopy bool) (any, error) {
	format, err := d.readByte()
	if err != nil {
		return nil, err
	}

	// Positive fixint
	if isPositiveFixint(format) {
		return int64(format), nil
	}

	// Negative fixint
	if isNegativeFixint(format) {
		return int64(int8(format)), nil
	}

	// Fixmap
	if isFixmap(format) {
		return decodeMapStringAnyWithLen(d, fixmapLen(format), zeroCopy)
	}

	// Fixarray
	if isFixarray(format) {
		return decodeArrayAnyWithLen(d, fixarrayLen(format), zeroCopy)
	}

	// Fixstr
	if isFixstr(format) {
		return decodeStringWithLen(d, fixstrLen(format), zeroCopy)
	}

	switch format {
	case formatNil:
		return nil, nil

	case formatFalse:
		return false, nil
	case formatTrue:
		return true, nil

	case formatUint8:
		v, err := d.readUint8()
		return int64(v), err
	case formatUint16:
		v, err := d.readUint16()
		return int64(v), err
	case formatUint32:
		v, err := d.readUint32()
		return int64(v), err
	case formatUint64:
		v, err := d.readUint64()
		if v > 9223372036854775807 {
			return v, err
		}
		return int64(v), err

	case formatInt8:
		v, err := d.readInt8()
		return int64(v), err
	case formatInt16:
		v, err := d.readInt16()
		return int64(v), err
	case formatInt32:
		v, err := d.readInt32()
		return int64(v), err
	case formatInt64:
		return d.readInt64()

	case formatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case formatFloat64:
		return d.readFloat64()

	case formatStr8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)
	case formatStr16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)
	case formatStr32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)

	case formatBin8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)
	case formatBin16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)
	case formatBin32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)

	case formatArray16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeArrayAnyWithLen(d, int(n), zeroCopy)
	case formatArray32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeArrayAnyWithLen(d, int(n), zeroCopy)

	case formatMap16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeMapStringAnyWithLen(d, int(n), zeroCopy)
	case formatMap32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeMapStringAnyWithLen(d, int(n), zeroCopy)

	default:
		return nil, ErrInvalidFormat
	}
}

func decodeStringWithLen(d *Decoder, length int, zeroCopy bool) (string, error) {
	if err := d.validateStringLen(length); err != nil {
		return "", err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	if zeroCopy {
		return unsafe.String(unsafe.SliceData(bytes), len(bytes)), nil
	}
	return string(bytes), nil
}

func decodeBinaryWithLen(d *Decoder, length int, zeroCopy bool) ([]byte, error) {
	if err := d.validateBinaryLen(length); err != nil {
		return nil, err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return nil, err
	}
	if zeroCopy {
		return bytes, nil
	}
	cp := make([]byte, length)
	copy(cp, bytes)
	return cp, nil
}

func decodeArrayAnyWithLen(d *Decoder, length int, zeroCopy bool) ([]any, error) {
	if err := d.validateArrayLen(length); err != nil {
		return nil, err
	}
	arr := make([]any, length)
	for i := 0; i < length; i++ {
		v, err := decodeAnyValue(d, zeroCopy)
		if err != nil {
			return nil, err
		}
		arr[i] = v
	}
	return arr, nil
}

func decodeMapStringAnyWithLen(d *Decoder, length int, zeroCopy bool) (map[string]any, error) {
	if err := d.validateMapLen(length); err != nil {
		return nil, err
	}
	m := make(map[string]any, length)
	for i := 0; i < length; i++ {
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return nil, err
		}
		var key string
		if zeroCopy {
			key = unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))
		} else {
			key = string(keyBytes)
		}
		val, err := decodeAnyValue(d, zeroCopy)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return m, nil
}

// UnmarshalMapStringString decodes msgpack into map[string]string.
// Much faster than map[string]any when you know the type.
func UnmarshalMapStringString(data []byte, zeroCopy bool) (map[string]string, error) {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)

	format, err := d.readByte()
	if err != nil {
		decoderPool.Put(d)
		return nil, err
	}

	var mapLen int
	if isFixmap(format) {
		mapLen = fixmapLen(format)
	} else if format == formatMap16 {
		n, err := d.readUint16()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}
		mapLen = int(n)
	} else if format == formatMap32 {
		n, err := d.readUint32()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}
		mapLen = int(n)
	} else if format == formatNil {
		decoderPool.Put(d)
		return nil, nil
	} else {
		decoderPool.Put(d)
		return nil, ErrTypeMismatch
	}

	if err := d.validateMapLen(mapLen); err != nil {
		decoderPool.Put(d)
		return nil, err
	}

	m := make(map[string]string, mapLen)
	for i := 0; i < mapLen; i++ {
		keyBytes, err := d.readStringBytes()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}

		valFormat, err := d.readByte()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}

		var valLen int
		if isFixstr(valFormat) {
			valLen = fixstrLen(valFormat)
		} else if valFormat == formatStr8 {
			n, err := d.readUint8()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == formatStr16 {
			n, err := d.readUint16()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == formatStr32 {
			n, err := d.readUint32()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == formatNil {
			// nil value - skip
			continue
		} else {
			decoderPool.Put(d)
			return nil, ErrTypeMismatch
		}

		if err := d.validateStringLen(valLen); err != nil {
			decoderPool.Put(d)
			return nil, err
		}
		valBytes, err := d.readBytes(valLen)
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}

		var key, val string
		if zeroCopy {
			key = unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))
			val = unsafe.String(unsafe.SliceData(valBytes), len(valBytes))
		} else {
			key = string(keyBytes)
			val = string(valBytes)
		}
		m[key] = val
	}

	decoderPool.Put(d)
	return m, nil
}

// Callback-based APIs for safe zero-copy usage.
// The callback guarantees the buffer stays valid for its duration.
// Strings are only valid within the callback - do not store them.

// DecodeMapFunc decodes msgpack into map[string]any and calls fn.
// Zero-copy strings: valid only within the callback.
// This is the safest way to use zero-copy when the buffer may be reused.
func DecodeMapFunc(data []byte, fn func(m map[string]any) error) error {
	m, err := UnmarshalMapStringAny(data, true)
	if err != nil {
		return err
	}
	return fn(m)
}

// DecodeStringMapFunc decodes msgpack into map[string]string and calls fn.
// Zero-copy strings: valid only within the callback.
func DecodeStringMapFunc(data []byte, fn func(m map[string]string) error) error {
	m, err := UnmarshalMapStringString(data, true)
	if err != nil {
		return err
	}
	return fn(m)
}

// DecodeStructFunc decodes msgpack into struct T and calls fn.
// Zero-copy strings: valid only within the callback.
// Uses cached zero-copy decoder for repeated types.
func DecodeStructFunc[T any](data []byte, fn func(v *T) error) error {
	dec := GetStructDecoderZeroCopy[T]()
	var v T
	if err := dec.Decode(data, &v); err != nil {
		return err
	}
	return fn(&v)
}
