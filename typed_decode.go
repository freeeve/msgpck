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
	if IsFixmap(format) {
		mapLen = FixmapLen(format)
	} else if format == FormatMap16 {
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		mapLen = int(n)
	} else if format == FormatMap32 {
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		mapLen = int(n)
	} else if format == FormatNil {
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
	if IsPositiveFixint(format) {
		return int64(format), nil
	}

	// Negative fixint
	if IsNegativeFixint(format) {
		return int64(int8(format)), nil
	}

	// Fixmap
	if IsFixmap(format) {
		return decodeMapStringAnyWithLen(d, FixmapLen(format), zeroCopy)
	}

	// Fixarray
	if IsFixarray(format) {
		return decodeArrayAnyWithLen(d, FixarrayLen(format), zeroCopy)
	}

	// Fixstr
	if IsFixstr(format) {
		return decodeStringWithLen(d, FixstrLen(format), zeroCopy)
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
		return d.readInt64()

	case FormatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case FormatFloat64:
		return d.readFloat64()

	case FormatStr8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)
	case FormatStr16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)
	case FormatStr32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeStringWithLen(d, int(n), zeroCopy)

	case FormatBin8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)
	case FormatBin16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)
	case FormatBin32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeBinaryWithLen(d, int(n), zeroCopy)

	case FormatArray16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeArrayAnyWithLen(d, int(n), zeroCopy)
	case FormatArray32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		return decodeArrayAnyWithLen(d, int(n), zeroCopy)

	case FormatMap16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		return decodeMapStringAnyWithLen(d, int(n), zeroCopy)
	case FormatMap32:
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
	if IsFixmap(format) {
		mapLen = FixmapLen(format)
	} else if format == FormatMap16 {
		n, err := d.readUint16()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}
		mapLen = int(n)
	} else if format == FormatMap32 {
		n, err := d.readUint32()
		if err != nil {
			decoderPool.Put(d)
			return nil, err
		}
		mapLen = int(n)
	} else if format == FormatNil {
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
		if IsFixstr(valFormat) {
			valLen = FixstrLen(valFormat)
		} else if valFormat == FormatStr8 {
			n, err := d.readUint8()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == FormatStr16 {
			n, err := d.readUint16()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == FormatStr32 {
			n, err := d.readUint32()
			if err != nil {
				decoderPool.Put(d)
				return nil, err
			}
			valLen = int(n)
		} else if valFormat == FormatNil {
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

// Convenience functions with zeroCopy=false (safe default)

// UnmarshalMap decodes msgpack into map[string]any.
func UnmarshalMap(data []byte) (map[string]any, error) {
	return UnmarshalMapStringAny(data, false)
}

// UnmarshalMapZeroCopy decodes msgpack into map[string]any with zero-copy strings.
func UnmarshalMapZeroCopy(data []byte) (map[string]any, error) {
	return UnmarshalMapStringAny(data, true)
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

// DecodeStructFuncWithDecoder uses a pre-registered decoder for best performance.
// The decoder should be created with .ZeroCopy() for zero allocations.
func DecodeStructFuncWithDecoder[T any](data []byte, dec *StructDecoder[T], fn func(v *T) error) error {
	var v T
	if err := dec.Decode(data, &v); err != nil {
		return err
	}
	return fn(&v)
}
