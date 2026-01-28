package msgpck

import (
	"reflect"
	"sync"
	"time"
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

// decodeAnyUintValue decodes uint8/16/32/64 formats to int64 or uint64.
func (d *Decoder) decodeAnyUintValue(format byte) (any, error) {
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

// decodeAnyIntValue decodes int8/16/32/64 formats to int64.
func (d *Decoder) decodeAnyIntValue(format byte) (any, error) {
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

// decodeAnyValue decodes directly to any, skipping intermediate Value struct
func (d *Decoder) decodeAnyValue(format byte) (any, error) {
	// Positive fixint: 0xxxxxxx
	if isPositiveFixint(format) {
		return int64(format), nil
	}

	// Negative fixint: 111xxxxx
	if isNegativeFixint(format) {
		return int64(int8(format)), nil
	}

	// Fixmap: 1000xxxx
	if isFixmap(format) {
		return d.decodeMapAny(fixmapLen(format))
	}

	// Fixarray: 1001xxxx
	if isFixarray(format) {
		return d.decodeArrayAny(fixarrayLen(format))
	}

	// Fixstr: 101xxxxx
	if isFixstr(format) {
		return d.decodeStringAny(fixstrLen(format))
	}

	switch format {
	case formatNil:
		return nil, nil

	case formatFalse:
		return false, nil
	case formatTrue:
		return true, nil

	case formatUint8, formatUint16, formatUint32, formatUint64:
		return d.decodeAnyUintValue(format)

	case formatInt8, formatInt16, formatInt32, formatInt64:
		return d.decodeAnyIntValue(format)

	case formatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case formatFloat64:
		return d.readFloat64()

	case formatStr8, formatStr16, formatStr32:
		length, err := d.parseStringLenSwitch(format)
		if err != nil {
			return nil, err
		}
		return d.decodeStringAny(length)

	case formatBin8, formatBin16, formatBin32:
		length, err := d.parseBinaryLen(format)
		if err != nil {
			return nil, err
		}
		return d.decodeBinaryAny(length)

	case formatArray16, formatArray32:
		length, err := d.parseArrayLenSwitch(format)
		if err != nil {
			return nil, err
		}
		return d.decodeArrayAny(length)

	case formatMap16, formatMap32:
		length, err := d.parseMapLenSwitch(format)
		if err != nil {
			return nil, err
		}
		return d.decodeMapAny(length)

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

		// Decode key as string using parseStringLen
		keyLen, err := d.parseStringLen(keyFormat)
		if err != nil {
			// Only treat ErrTypeMismatch as invalid format (non-string key)
			if err == ErrTypeMismatch {
				return nil, ErrInvalidFormat
			}
			return nil, err
		}
		key, err := d.decodeStringAny(keyLen)
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

// decodeInto decodes msgpack data into v using reflection.
func (d *Decoder) decodeInto(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrNotPointer
	}
	return d.decodeReflect(rv.Elem())
}

func (d *Decoder) decodeReflect(rv reflect.Value) error {
	// Handle pointer types - allocate if nil
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decodeReflect(rv.Elem())
	}

	// Special case for time.Time
	if rv.Type() == timeType {
		t, err := d.DecodeTimestamp()
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(t))
		return nil
	}

	switch rv.Kind() {
	case reflect.Struct:
		return d.DecodeStruct(rv.Addr().Interface())

	case reflect.Map:
		return d.decodeMapReflect(rv)

	case reflect.Slice:
		return d.decodeSliceReflect(rv)

	case reflect.String:
		s, err := d.decodeStringValue()
		if err != nil {
			return err
		}
		rv.SetString(s)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := d.decodeIntValue()
		if err != nil {
			return err
		}
		rv.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := d.decodeUintValue()
		if err != nil {
			return err
		}
		rv.SetUint(u)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := d.decodeFloatValue()
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil

	case reflect.Bool:
		b, err := d.decodeBoolValue()
		if err != nil {
			return err
		}
		rv.SetBool(b)
		return nil

	case reflect.Interface:
		// Decode to any
		val, err := d.DecodeAny()
		if err != nil {
			return err
		}
		if val != nil {
			rv.Set(reflect.ValueOf(val))
		}
		return nil

	default:
		return ErrUnsupportedType
	}
}

var timeType = reflect.TypeOf(time.Time{})

func (d *Decoder) decodeStringValue() (string, error) {
	format, err := d.readByte()
	if err != nil {
		return "", err
	}
	length, err := d.parseStringLen(format)
	if err != nil {
		return "", err
	}
	if err := d.validateStringLen(length); err != nil {
		return "", err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (d *Decoder) decodeIntValue() (int64, error) {
	format, err := d.readByte()
	if err != nil {
		return 0, err
	}
	if isPositiveFixint(format) {
		return int64(format), nil
	}
	if isNegativeFixint(format) {
		return int64(int8(format)), nil
	}
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
	case formatInt64:
		return d.readInt64()
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
		return int64(v), err
	default:
		return 0, ErrTypeMismatch
	}
}

func (d *Decoder) decodeUintValue() (uint64, error) {
	format, err := d.readByte()
	if err != nil {
		return 0, err
	}
	if isPositiveFixint(format) {
		return uint64(format), nil
	}
	switch format {
	case formatUint8:
		v, err := d.readUint8()
		return uint64(v), err
	case formatUint16:
		v, err := d.readUint16()
		return uint64(v), err
	case formatUint32:
		v, err := d.readUint32()
		return uint64(v), err
	case formatUint64:
		return d.readUint64()
	case formatInt8:
		v, err := d.readInt8()
		return uint64(v), err
	case formatInt16:
		v, err := d.readInt16()
		return uint64(v), err
	case formatInt32:
		v, err := d.readInt32()
		return uint64(v), err
	case formatInt64:
		v, err := d.readInt64()
		return uint64(v), err
	default:
		return 0, ErrTypeMismatch
	}
}

func (d *Decoder) decodeFloatValue() (float64, error) {
	format, err := d.readByte()
	if err != nil {
		return 0, err
	}
	switch format {
	case formatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case formatFloat64:
		return d.readFloat64()
	default:
		return 0, ErrTypeMismatch
	}
}

func (d *Decoder) decodeBoolValue() (bool, error) {
	format, err := d.readByte()
	if err != nil {
		return false, err
	}
	switch format {
	case formatTrue:
		return true, nil
	case formatFalse:
		return false, nil
	default:
		return false, ErrTypeMismatch
	}
}

func (d *Decoder) decodeMapReflect(rv reflect.Value) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}
	if format == formatNil {
		return nil
	}
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return err
	}
	if err := d.validateMapLen(mapLen); err != nil {
		return err
	}
	if rv.IsNil() {
		rv.Set(reflect.MakeMapWithSize(rv.Type(), mapLen))
	}
	keyType := rv.Type().Key()
	valType := rv.Type().Elem()
	for i := 0; i < mapLen; i++ {
		key := reflect.New(keyType).Elem()
		if err := d.decodeReflect(key); err != nil {
			return err
		}
		val := reflect.New(valType).Elem()
		if err := d.decodeReflect(val); err != nil {
			return err
		}
		rv.SetMapIndex(key, val)
	}
	return nil
}

func (d *Decoder) decodeSliceReflect(rv reflect.Value) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}

	// Handle []byte specially - it's encoded as binary
	if rv.Type().Elem().Kind() == reflect.Uint8 {
		return d.decodeBinaryReflect(format, rv)
	}

	if format == formatNil {
		return nil
	}
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return err
	}
	slice := reflect.MakeSlice(rv.Type(), arrLen, arrLen)
	for i := 0; i < arrLen; i++ {
		if err := d.decodeReflect(slice.Index(i)); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

func (d *Decoder) decodeBinaryReflect(format byte, rv reflect.Value) error {
	if format == formatNil {
		return nil
	}
	length, err := d.parseBinaryLen(format)
	if err != nil {
		// Maybe it's a string encoded as str format
		length, err = d.parseStringLen(format)
		if err != nil {
			return err
		}
	}
	if err := d.validateBinaryLen(length); err != nil {
		return err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return err
	}
	cp := make([]byte, length)
	copy(cp, bytes)
	rv.SetBytes(cp)
	return nil
}

// Unmarshal decodes msgpack data into v.
// v must be a pointer to the target value.
//
// Supported target types:
//   - *struct: decoded from msgpack map
//   - *map[K]V: decoded from msgpack map
//   - *[]T: decoded from msgpack array
//   - *string, *int64, *float64, *bool, *[]byte: decoded from msgpack primitives
//   - *any: decoded to map[string]any, []any, or primitive types
//   - *time.Time: decoded from msgpack timestamp extension
func Unmarshal(data []byte, v any) error {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)
	err := d.decodeInto(v)
	decoderPool.Put(d)
	return err
}

// UnmarshalWithConfig decodes msgpack data into v with custom config.
func UnmarshalWithConfig(data []byte, v any, cfg Config) error {
	d := NewDecoderWithConfig(data, cfg)
	return d.decodeInto(v)
}
