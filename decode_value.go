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
	if IsPositiveFixint(format) {
		return Value{Type: TypeUint, Uint: uint64(format)}, nil
	}

	// Negative fixint: 111xxxxx
	if IsNegativeFixint(format) {
		return Value{Type: TypeInt, Int: int64(int8(format))}, nil
	}

	// Fixmap: 1000xxxx
	if IsFixmap(format) {
		return d.decodeMap(FixmapLen(format))
	}

	// Fixarray: 1001xxxx
	if IsFixarray(format) {
		return d.decodeArray(FixarrayLen(format))
	}

	// Fixstr: 101xxxxx
	if IsFixstr(format) {
		return d.decodeString(FixstrLen(format))
	}

	switch format {
	// Nil
	case FormatNil:
		return Value{Type: TypeNil}, nil

	// Bool
	case FormatFalse:
		return Value{Type: TypeBool, Bool: false}, nil
	case FormatTrue:
		return Value{Type: TypeBool, Bool: true}, nil

	// Unsigned integers
	case FormatUint8:
		v, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case FormatUint16:
		v, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case FormatUint32:
		v, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: uint64(v)}, nil

	case FormatUint64:
		v, err := d.readUint64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeUint, Uint: v}, nil

	// Signed integers
	case FormatInt8:
		v, err := d.readInt8()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case FormatInt16:
		v, err := d.readInt16()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case FormatInt32:
		v, err := d.readInt32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: int64(v)}, nil

	case FormatInt64:
		v, err := d.readInt64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeInt, Int: v}, nil

	// Floats
	case FormatFloat32:
		v, err := d.readFloat32()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeFloat32, Float32: v}, nil

	case FormatFloat64:
		v, err := d.readFloat64()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeFloat64, Float64: v}, nil

	// Strings
	case FormatStr8:
		length, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return d.decodeString(int(length))

	case FormatStr16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeString(int(length))

	case FormatStr32:
		length, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return d.decodeString(int(length))

	// Binary
	case FormatBin8:
		length, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return d.decodeBinary(int(length))

	case FormatBin16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeBinary(int(length))

	case FormatBin32:
		length, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return d.decodeBinary(int(length))

	// Arrays
	case FormatArray16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeArray(int(length))

	case FormatArray32:
		length, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return d.decodeArray(int(length))

	// Maps
	case FormatMap16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeMap(int(length))

	case FormatMap32:
		length, err := d.readUint32()
		if err != nil {
			return Value{}, err
		}
		return d.decodeMap(int(length))

	// Fixed ext
	case FormatFixExt1:
		return d.decodeExt(1)
	case FormatFixExt2:
		return d.decodeExt(2)
	case FormatFixExt4:
		return d.decodeExt(4)
	case FormatFixExt8:
		return d.decodeExt(8)
	case FormatFixExt16:
		return d.decodeExt(16)

	// Variable ext
	case FormatExt8:
		length, err := d.readUint8()
		if err != nil {
			return Value{}, err
		}
		return d.decodeExt(int(length))

	case FormatExt16:
		length, err := d.readUint16()
		if err != nil {
			return Value{}, err
		}
		return d.decodeExt(int(length))

	case FormatExt32:
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
		keyFormat, err := d.readByte()
		if err != nil {
			return Value{}, err
		}

		// Key must be a string (common case) or other type
		var key []byte
		if IsFixstr(keyFormat) {
			keyLen := FixstrLen(keyFormat)
			if err := d.validateStringLen(keyLen); err != nil {
				return Value{}, err
			}
			key, err = d.readBytes(keyLen)
			if err != nil {
				return Value{}, err
			}
		} else {
			// Non-string key - decode as value and use its bytes
			keyVal, err := d.decodeValue(keyFormat)
			if err != nil {
				return Value{}, err
			}
			if keyVal.Type == TypeString {
				key = keyVal.Bytes
			} else {
				// For non-string keys, we need to store the original bytes
				// This is a limitation - we'll use nil for non-string keys
				key = nil
			}
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
