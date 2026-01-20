package msgpck

// Decode decodes a single MessagePack value.
// Returns a Value with zero-copy references to strings/binary data in the source.
func (d *Decoder) Decode() (Value, error) {
	format, err := d.readByte()
	if err != nil {
		return Value{}, err
	}
	return d.decodeValue(format)
}

// decodeValue decodes a value given its format byte
func (d *Decoder) decodeValue(format byte) (Value, error) {
	// Positive fixint: 0xxxxxxx
	if isPositiveFixint(format) {
		return Value{Type: TypeUint, Uint: uint64(format)}, nil
	}

	// Negative fixint: 111xxxxx
	if isNegativeFixint(format) {
		return Value{Type: TypeInt, Int: int64(int8(format))}, nil
	}

	// Fixmap: 1000xxxx
	if isFixmap(format) {
		return d.decodeMap(fixmapLen(format))
	}

	// Fixarray: 1001xxxx
	if isFixarray(format) {
		return d.decodeArray(fixarrayLen(format))
	}

	// Fixstr: 101xxxxx
	if isFixstr(format) {
		return d.decodeString(fixstrLen(format))
	}

	switch format {
	// Nil
	case formatNil:
		return Value{Type: TypeNil}, nil

	// Bool
	case formatFalse:
		return Value{Type: TypeBool, Bool: false}, nil
	case formatTrue:
		return Value{Type: TypeBool, Bool: true}, nil

	// Unsigned integers
	case formatUint8:
		v, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case formatUint16:
		v, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case formatUint32:
		v, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case formatUint64:
		v, err := d.readUint64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: v}, nil

	// Signed integers
	case formatInt8:
		v, err := d.readInt8()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case formatInt16:
		v, err := d.readInt16()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case formatInt32:
		v, err := d.readInt32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case formatInt64:
		v, err := d.readInt64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: v}, nil

	// Floats
	case formatFloat32:
		v, err := d.readFloat32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeFloat32, Float32: v}, nil

	case formatFloat64:
		v, err := d.readFloat64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeFloat64, Float64: v}, nil

	// Strings
	case formatStr8, formatStr16, formatStr32:
		length, err := d.parseStringLenSwitch(format)
		if err != nil {
			return Value{}, err
		}
		return d.decodeString(length)

	// Binary
	case formatBin8, formatBin16, formatBin32:
		length, err := d.parseBinaryLen(format)
		if err != nil {
			return Value{}, err
		}
		return d.decodeBinary(length)

	// Arrays
	case formatArray16, formatArray32:
		length, err := d.parseArrayLenSwitch(format)
		if err != nil {
			return Value{}, err
		}
		return d.decodeArray(length)

	// Maps
	case formatMap16, formatMap32:
		length, err := d.parseMapLenSwitch(format)
		if err != nil {
			return Value{}, err
		}
		return d.decodeMap(length)

	// Fixed ext
	case formatFixExt1:
		return d.decodeExt(1)
	case formatFixExt2:
		return d.decodeExt(2)
	case formatFixExt4:
		return d.decodeExt(4)
	case formatFixExt8:
		return d.decodeExt(8)
	case formatFixExt16:
		return d.decodeExt(16)

	// Variable ext
	case formatExt8:
		length, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return d.decodeExt(int(length))

	case formatExt16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeExt(int(length))

	case formatExt32:
		length, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return d.decodeExt(int(length))

	default:
		return Value{}, ErrInvalidFormat
	}
}

// decodeString decodes a string of known length (zero-copy)
func (d *Decoder) decodeString(length int) (Value, error) {
	if err := d.validateStringLen(length); err != nil {
		return Value{}, err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return Value{}, err
	}
	return Value{Type: TypeString, Bytes: bytes}, nil
}

// decodeBinary decodes binary data of known length (zero-copy)
func (d *Decoder) decodeBinary(length int) (Value, error) {
	if err := d.validateBinaryLen(length); err != nil {
		return Value{}, err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return Value{}, err
	}
	return Value{Type: TypeBinary, Bytes: bytes}, nil
}

// decodeArray decodes an array of known length
func (d *Decoder) decodeArray(length int) (Value, error) {
	if err := d.validateArrayLen(length); err != nil {
		return Value{}, err
	}
	if err := d.enterContainer(); err != nil {
		return Value{}, err
	}
	defer d.leaveContainer()

	arr := make([]Value, length)
	for i := 0; i < length; i++ {
		v, err := d.Decode()
		if err != nil {
			return Value{}, err
		}
		arr[i] = v
	}
	return Value{Type: TypeArray, Array: arr}, nil
}

// decodeMapKey decodes a map key (usually string, but handles other types).
func (d *Decoder) decodeMapKey() ([]byte, error) {
	keyFormat, err := d.readByte()
	if err != nil {
		return nil, err
	}

	// Try to parse as string (common case)
	keyLen, err := d.parseStringLen(keyFormat)
	if err == nil {
		if err := d.validateStringLen(keyLen); err != nil {
			return nil, err
		}
		return d.readBytes(keyLen)
	}

	// Non-string key - decode as value and extract bytes if string
	keyVal, err := d.decodeValue(keyFormat)
	if err != nil {
		return nil, err
	}
	if keyVal.Type == TypeString {
		return keyVal.Bytes, nil
	}
	// For non-string keys, return nil (limitation)
	return nil, nil
}

// decodeMap decodes a map of known length
func (d *Decoder) decodeMap(length int) (Value, error) {
	if err := d.validateMapLen(length); err != nil {
		return Value{}, err
	}
	if err := d.enterContainer(); err != nil {
		return Value{}, err
	}
	defer d.leaveContainer()

	kvs := make([]KV, length)
	for i := 0; i < length; i++ {
		// Read key
		key, err := d.decodeMapKey()
		if err != nil {
			return Value{}, err
		}

		// Read value
		val, err := d.Decode()
		if err != nil {
			return Value{}, err
		}

		kvs[i] = KV{Key: key, Value: val}
	}
	return Value{Type: TypeMap, Map: kvs}, nil
}

// decodeExt decodes extension data of known length
func (d *Decoder) decodeExt(length int) (Value, error) {
	if err := d.validateExtLen(length); err != nil {
		return Value{}, err
	}

	extType, err := d.readInt8()
	if err != nil {
		return Value{}, err
	}

	data, err := d.readBytes(length)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Type: TypeExt,
		Ext:  Ext{Type: extType, Data: data},
	}, nil
}
