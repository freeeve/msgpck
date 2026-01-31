package msgpck

import "unsafe"

// bytesToString converts bytes to string based on zeroCopy flag.
func bytesToString(b []byte, zeroCopy bool) string {
	if zeroCopy {
		return unsafe.String(unsafe.SliceData(b), len(b))
	}
	return string(b)
}

// decodeAnyUint decodes uint format bytes to int64 (or uint64 for overflow).
func decodeAnyUint(d *Decoder, format byte) (any, error) {
	switch format {
	case formatUint8:
		v, err := d.readUint8()
		return int64(v), err
	case formatUint16:
		v, err := d.readUint16()
		return int64(v), err
	case formatUint32:
		v, err := d.readUint32()
		return int64(v), err
	default: // formatUint64
		v, err := d.readUint64()
		if v > 9223372036854775807 {
			return v, err
		}
		return int64(v), err
	}
}

// decodeAnyInt decodes int format bytes to int64.
func decodeAnyInt(d *Decoder, format byte) (any, error) {
	switch format {
	case formatInt8:
		v, err := d.readInt8()
		return int64(v), err
	case formatInt16:
		v, err := d.readInt16()
		return int64(v), err
	case formatInt32:
		v, err := d.readInt32()
		return int64(v), err
	default: // formatInt64
		return d.readInt64()
	}
}

// decodeAnyStr decodes str8/16/32 format bytes to string.
func decodeAnyStr(d *Decoder, format byte, zeroCopy bool) (string, error) {
	length, err := d.parseStringLenSwitch(format)
	if err != nil {
		return "", err
	}
	return decodeStringWithLen(d, length, zeroCopy)
}

// decodeAnyBin decodes bin8/16/32 format bytes to []byte.
func decodeAnyBin(d *Decoder, format byte, zeroCopy bool) ([]byte, error) {
	length, err := d.parseBinaryLen(format)
	if err != nil {
		return nil, err
	}
	return decodeBinaryWithLen(d, length, zeroCopy)
}

// decodeAnyArray decodes array16/32 format bytes to []any.
func decodeAnyArray(d *Decoder, format byte, zeroCopy bool) ([]any, error) {
	length, err := d.parseArrayLenSwitch(format)
	if err != nil {
		return nil, err
	}
	return decodeArrayAnyWithLen(d, length, zeroCopy)
}

// decodeAnyMap decodes map16/32 format bytes to map[string]any.
func decodeAnyMap(d *Decoder, format byte, zeroCopy bool) (map[string]any, error) {
	length, err := d.parseMapLenSwitch(format)
	if err != nil {
		return nil, err
	}
	return decodeMapStringAnyWithLen(d, length, zeroCopy)
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

	case formatUint8, formatUint16, formatUint32, formatUint64:
		return decodeAnyUint(d, format)

	case formatInt8, formatInt16, formatInt32, formatInt64:
		return decodeAnyInt(d, format)

	case formatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case formatFloat64:
		return d.readFloat64()

	case formatStr8, formatStr16, formatStr32:
		return decodeAnyStr(d, format, zeroCopy)

	case formatBin8, formatBin16, formatBin32:
		return decodeAnyBin(d, format, zeroCopy)

	case formatArray16, formatArray32:
		return decodeAnyArray(d, format, zeroCopy)

	case formatMap16, formatMap32:
		return decodeAnyMap(d, format, zeroCopy)

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
		key := bytesToString(keyBytes, zeroCopy)
		val, err := decodeAnyValue(d, zeroCopy)
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
	return m, nil
}
