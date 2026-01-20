package msgpck

import (
	"reflect"
	"strings"
	"sync"
)

// structInfo holds cached reflection info for a struct type
type structInfo struct {
	fields []fieldInfo
}

// fieldInfo holds info about a single struct field
type fieldInfo struct {
	index     []int  // field index path (supports embedded)
	name      []byte // msgpack field name
	omitempty bool
}

// structCache caches struct info to avoid repeated reflection
var structCache sync.Map // map[reflect.Type]*structInfo

// getStructInfo returns cached struct info, computing it if necessary
func getStructInfo(t reflect.Type) *structInfo {
	if cached, ok := structCache.Load(t); ok {
		return cached.(*structInfo)
	}

	info := buildStructInfo(t)
	structCache.Store(t, info)
	return info
}

// buildStructInfo builds struct info via reflection
func buildStructInfo(t reflect.Type) *structInfo {
	info := &structInfo{}
	buildStructFields(t, nil, info)
	return info
}

// buildStructFields recursively builds field info, handling embedded structs
func buildStructFields(t reflect.Type, index []int, info *structInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		fieldIndex := append(append([]int(nil), index...), i)

		// Handle embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			buildStructFields(field.Type, fieldIndex, info)
			continue
		}

		// Parse tag
		tag := field.Tag.Get("msgpack")
		if tag == "-" {
			continue // skip this field
		}

		name, opts := parseTag(tag)
		if name == "" {
			name = field.Name
		}

		fi := fieldInfo{
			index:     fieldIndex,
			name:      []byte(name),
			omitempty: strings.Contains(opts, "omitempty"),
		}
		info.fields = append(info.fields, fi)
	}
}

// parseTag parses a struct tag like "name,omitempty"
func parseTag(tag string) (name, opts string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

// parseMapLen parses the map length from a format byte.
func (d *Decoder) parseMapLen(format byte) (int, error) {
	if isFixmap(format) {
		return fixmapLen(format), nil
	}
	switch format {
	case formatMap16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatMap32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseArrayLen parses the array length from a format byte.
func (d *Decoder) parseArrayLen(format byte) (int, error) {
	if isFixarray(format) {
		return fixarrayLen(format), nil
	}
	switch format {
	case formatArray16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatArray32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseStringLen parses the string length from a format byte.
func (d *Decoder) parseStringLen(format byte) (int, error) {
	if isFixstr(format) {
		return fixstrLen(format), nil
	}
	switch format {
	case formatStr8:
		n, err := d.readUint8()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatStr16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatStr32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseBinaryLen parses the binary length from a format byte.
func (d *Decoder) parseBinaryLen(format byte) (int, error) {
	switch format {
	case formatBin8:
		n, err := d.readUint8()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatBin16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatBin32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseStringLenSwitch parses string length for str8/str16/str32 formats only.
func (d *Decoder) parseStringLenSwitch(format byte) (int, error) {
	switch format {
	case formatStr8:
		n, err := d.readUint8()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatStr16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatStr32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseArrayLenSwitch parses array length for array16/array32 formats only.
func (d *Decoder) parseArrayLenSwitch(format byte) (int, error) {
	switch format {
	case formatArray16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatArray32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// parseMapLenSwitch parses map length for map16/map32 formats only.
func (d *Decoder) parseMapLenSwitch(format byte) (int, error) {
	switch format {
	case formatMap16:
		n, err := d.readUint16()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case formatMap32:
		n, err := d.readUint32()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// DecodeStruct decodes a msgpack map into a struct.
// v must be a pointer to a struct.
func (d *Decoder) DecodeStruct(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// Read format byte
	format, err := d.readByte()
	if err != nil {
		return err
	}

	// Handle nil
	if format == formatNil {
		return nil // nil decodes to zero value
	}

	// Determine map length
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return err
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return err
	}
	if err := d.enterContainer(); err != nil {
		return err
	}
	defer d.leaveContainer()

	// Get cached struct info
	info := getStructInfo(rv.Type())

	// Decode each key-value pair
	for i := 0; i < mapLen; i++ {
		// Read key
		keyFormat, err := d.readByte()
		if err != nil {
			return err
		}

		var key []byte
		if isFixstr(keyFormat) {
			keyLen := fixstrLen(keyFormat)
			if err := d.validateStringLen(keyLen); err != nil {
				return err
			}
			key, err = d.readBytes(keyLen)
			if err != nil {
				return err
			}
		} else if keyFormat == formatStr8 {
			n, err := d.readUint8()
			if err != nil {
				return err
			}
			if err := d.validateStringLen(int(n)); err != nil {
				return err
			}
			key, err = d.readBytes(int(n))
			if err != nil {
				return err
			}
		} else if keyFormat == formatStr16 {
			n, err := d.readUint16()
			if err != nil {
				return err
			}
			if err := d.validateStringLen(int(n)); err != nil {
				return err
			}
			key, err = d.readBytes(int(n))
			if err != nil {
				return err
			}
		} else if keyFormat == formatStr32 {
			n, err := d.readUint32()
			if err != nil {
				return err
			}
			if err := d.validateStringLen(int(n)); err != nil {
				return err
			}
			key, err = d.readBytes(int(n))
			if err != nil {
				return err
			}
		} else {
			// Skip non-string key
			_, err := d.decodeValue(keyFormat)
			if err != nil {
				return err
			}
			// Skip the value too
			_, err = d.Decode()
			if err != nil {
				return err
			}
			continue
		}

		// Find matching field
		var field *fieldInfo
		for j := range info.fields {
			if bytesEqual(info.fields[j].name, key) {
				field = &info.fields[j]
				break
			}
		}

		if field == nil {
			// Skip unknown field
			_, err := d.Decode()
			if err != nil {
				return err
			}
			continue
		}

		// Get the field value
		fv := rv.FieldByIndex(field.index)
		if !fv.CanSet() {
			// Skip non-settable field
			_, err := d.Decode()
			if err != nil {
				return err
			}
			continue
		}

		// Decode into the field
		if err := d.decodeIntoValue(fv); err != nil {
			return err
		}
	}

	return nil
}

// decodeIntoValue decodes a msgpack value into a reflect.Value
func (d *Decoder) decodeIntoValue(rv reflect.Value) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}

	// Handle nil
	if format == formatNil {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		if format == formatTrue {
			rv.SetBool(true)
		} else if format == formatFalse {
			rv.SetBool(false)
		} else {
			return ErrTypeMismatch
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := d.decodeValueInt(format)
		if err != nil {
			return err
		}
		rv.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := d.decodeValueUint(format)
		if err != nil {
			return err
		}
		rv.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := d.decodeValueFloat(format)
		if err != nil {
			return err
		}
		rv.SetFloat(v)

	case reflect.String:
		v, err := d.decodeValueString(format)
		if err != nil {
			return err
		}
		rv.SetString(string(v))

	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// []byte
			v, err := d.decodeValueBytes(format)
			if err != nil {
				return err
			}
			rv.SetBytes(v)
		} else {
			// Other slice
			return d.decodeIntoSlice(rv, format)
		}

	case reflect.Array:
		return d.decodeIntoArray(rv, format)

	case reflect.Map:
		return d.decodeIntoMap(rv, format)

	case reflect.Struct:
		return d.decodeIntoStruct(rv, format)

	case reflect.Ptr:
		// Allocate and decode
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return d.decodeIntoValue(rv.Elem())

	case reflect.Interface:
		// Decode to any
		val, err := d.decodeValue(format)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(valueToAny(&val)))

	default:
		return ErrUnsupportedType
	}

	return nil
}

// decodeValueInt decodes an integer value
func (d *Decoder) decodeValueInt(format byte) (int64, error) {
	if isPositiveFixint(format) {
		return int64(format), nil
	}
	if isNegativeFixint(format) {
		return int64(int8(format)), nil
	}

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
	case formatUint64:
		v, err := d.readUint64()
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
	default:
		return 0, ErrTypeMismatch
	}
}

// decodeValueUint decodes an unsigned integer value
func (d *Decoder) decodeValueUint(format byte) (uint64, error) {
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

// decodeValueFloat decodes a float value
func (d *Decoder) decodeValueFloat(format byte) (float64, error) {
	// First check if it's an integer (common in msgpack)
	if isPositiveFixint(format) {
		return float64(format), nil
	}
	if isNegativeFixint(format) {
		return float64(int8(format)), nil
	}

	switch format {
	case formatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case formatFloat64:
		return d.readFloat64()
	case formatUint8:
		v, err := d.readUint8()
		return float64(v), err
	case formatUint16:
		v, err := d.readUint16()
		return float64(v), err
	case formatUint32:
		v, err := d.readUint32()
		return float64(v), err
	case formatUint64:
		v, err := d.readUint64()
		return float64(v), err
	case formatInt8:
		v, err := d.readInt8()
		return float64(v), err
	case formatInt16:
		v, err := d.readInt16()
		return float64(v), err
	case formatInt32:
		v, err := d.readInt32()
		return float64(v), err
	case formatInt64:
		v, err := d.readInt64()
		return float64(v), err
	default:
		return 0, ErrTypeMismatch
	}
}

// decodeValueString decodes a string value (zero-copy)
func (d *Decoder) decodeValueString(format byte) ([]byte, error) {
	var length int

	if isFixstr(format) {
		length = fixstrLen(format)
	} else {
		switch format {
		case formatStr8:
			n, err := d.readUint8()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr16:
			n, err := d.readUint16()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr32:
			n, err := d.readUint32()
			if err != nil {
				return nil, err
			}
			length = int(n)
		default:
			return nil, ErrTypeMismatch
		}
	}

	if err := d.validateStringLen(length); err != nil {
		return nil, err
	}
	return d.readBytes(length)
}

// decodeValueBytes decodes binary or string data
func (d *Decoder) decodeValueBytes(format byte) ([]byte, error) {
	// Handle string formats (for []byte fields)
	if isFixstr(format) {
		length := fixstrLen(format)
		if err := d.validateStringLen(length); err != nil {
			return nil, err
		}
		return d.readBytes(length)
	}

	var length int
	switch format {
	case formatStr8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateStringLen(length); err != nil {
			return nil, err
		}
	case formatStr16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateStringLen(length); err != nil {
			return nil, err
		}
	case formatStr32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateStringLen(length); err != nil {
			return nil, err
		}
	case formatBin8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateBinaryLen(length); err != nil {
			return nil, err
		}
	case formatBin16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateBinaryLen(length); err != nil {
			return nil, err
		}
	case formatBin32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		length = int(n)
		if err := d.validateBinaryLen(length); err != nil {
			return nil, err
		}
	default:
		return nil, ErrTypeMismatch
	}

	return d.readBytes(length)
}

// decodeIntoSlice decodes an array into a slice
func (d *Decoder) decodeIntoSlice(rv reflect.Value, format byte) error {
	length, err := d.parseArrayLen(format)
	if err != nil {
		return err
	}

	if err := d.validateArrayLen(length); err != nil {
		return err
	}
	if err := d.enterContainer(); err != nil {
		return err
	}
	defer d.leaveContainer()

	slice := reflect.MakeSlice(rv.Type(), length, length)
	for i := 0; i < length; i++ {
		if err := d.decodeIntoValue(slice.Index(i)); err != nil {
			return err
		}
	}
	rv.Set(slice)
	return nil
}

// decodeIntoArray decodes into a fixed-size array
func (d *Decoder) decodeIntoArray(rv reflect.Value, format byte) error {
	length, err := d.parseArrayLen(format)
	if err != nil {
		return err
	}

	if err := d.validateArrayLen(length); err != nil {
		return err
	}
	if err := d.enterContainer(); err != nil {
		return err
	}
	defer d.leaveContainer()

	arrayLen := rv.Len()
	for i := 0; i < length; i++ {
		if i < arrayLen {
			if err := d.decodeIntoValue(rv.Index(i)); err != nil {
				return err
			}
		} else {
			// Skip extra elements
			if _, err := d.Decode(); err != nil {
				return err
			}
		}
	}
	return nil
}

// decodeIntoMap decodes into a map
func (d *Decoder) decodeIntoMap(rv reflect.Value, format byte) error {
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return err
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return err
	}
	if err := d.enterContainer(); err != nil {
		return err
	}
	defer d.leaveContainer()

	if rv.IsNil() {
		rv.Set(reflect.MakeMapWithSize(rv.Type(), mapLen))
	}

	keyType := rv.Type().Key()
	valType := rv.Type().Elem()

	for i := 0; i < mapLen; i++ {
		key := reflect.New(keyType).Elem()
		if err := d.decodeIntoValue(key); err != nil {
			return err
		}

		val := reflect.New(valType).Elem()
		if err := d.decodeIntoValue(val); err != nil {
			return err
		}

		rv.SetMapIndex(key, val)
	}
	return nil
}

// decodeIntoStruct decodes a map into a struct (nested)
func (d *Decoder) decodeIntoStruct(rv reflect.Value, format byte) error {
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return err
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return err
	}
	if err := d.enterContainer(); err != nil {
		return err
	}
	defer d.leaveContainer()

	info := getStructInfo(rv.Type())

	for i := 0; i < mapLen; i++ {
		// Read key
		keyBytes, err := d.decodeStringKey()
		if err != nil {
			// Skip non-string key and its value
			if _, err := d.Decode(); err != nil {
				return err
			}
			continue
		}

		// Find matching field
		var field *fieldInfo
		for j := range info.fields {
			if bytesEqual(info.fields[j].name, keyBytes) {
				field = &info.fields[j]
				break
			}
		}

		if field == nil {
			// Skip unknown field
			if _, err := d.Decode(); err != nil {
				return err
			}
			continue
		}

		fv := rv.FieldByIndex(field.index)
		if !fv.CanSet() {
			if _, err := d.Decode(); err != nil {
				return err
			}
			continue
		}

		if err := d.decodeIntoValue(fv); err != nil {
			return err
		}
	}

	return nil
}

// decodeStringKey reads a string key and returns it as []byte
func (d *Decoder) decodeStringKey() ([]byte, error) {
	format, err := d.readByte()
	if err != nil {
		return nil, err
	}

	if isFixstr(format) {
		length := fixstrLen(format)
		if err := d.validateStringLen(length); err != nil {
			return nil, err
		}
		return d.readBytes(length)
	}

	switch format {
	case formatStr8:
		n, err := d.readUint8()
		if err != nil {
			return nil, err
		}
		if err := d.validateStringLen(int(n)); err != nil {
			return nil, err
		}
		return d.readBytes(int(n))
	case formatStr16:
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		if err := d.validateStringLen(int(n)); err != nil {
			return nil, err
		}
		return d.readBytes(int(n))
	case formatStr32:
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		if err := d.validateStringLen(int(n)); err != nil {
			return nil, err
		}
		return d.readBytes(int(n))
	default:
		return nil, ErrTypeMismatch
	}
}

// UnmarshalStruct is a convenience function that decodes msgpack data into a struct.
func UnmarshalStruct(data []byte, v any) error {
	d := NewDecoder(data)
	return d.DecodeStruct(v)
}

// UnmarshalStructWithConfig decodes msgpack data into a struct with custom config.
func UnmarshalStructWithConfig(data []byte, v any, cfg Config) error {
	d := NewDecoderWithConfig(data, cfg)
	return d.DecodeStruct(v)
}
