package msgpck

import (
	"reflect"
	"sync"
	"unsafe"
)

// StructDecoder is a pre-registered decoder for a specific struct type.
// Create once, reuse many times for best performance.
type StructDecoder[T any] struct {
	fields   []structField
	zeroCopy bool // if true, strings point into input buffer (no alloc)
}

type structField struct {
	name       string       // msgpack field name
	offset     uintptr      // field offset in struct
	kind       reflect.Kind // field type
	elem       reflect.Type // for slices/maps: element type
	structType reflect.Type // for struct fields: the struct type
}

// parseStructTagName extracts the field name from a msgpack struct tag.
// Handles "name,omitempty" style tags by extracting just the name part.
func parseStructTagName(tag, defaultName string) string {
	if tag == "" || tag == "-" {
		if tag == "-" {
			return "-"
		}
		return defaultName
	}
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			return tag[:i]
		}
	}
	return tag
}

// buildNestedFieldMap builds a field name to StructField map for nested struct decoding.
func buildNestedFieldMap(structType reflect.Type) map[string]reflect.StructField {
	fieldMap := make(map[string]reflect.StructField, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		if !f.IsExported() {
			continue
		}
		name := parseStructTagName(f.Tag.Get("msgpack"), f.Name)
		if name == "-" {
			continue
		}
		fieldMap[name] = f
	}
	return fieldMap
}

// newStructDecoder creates a reusable decoder for type T.
// Caches field information for fast repeated decoding.
// Use .ZeroCopy() for zero-allocation decoding when input buffer outlives result.
func newStructDecoder[T any]() *StructDecoder[T] {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	dec := &StructDecoder[T]{
		fields: make([]structField, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		name := parseStructTagName(f.Tag.Get("msgpack"), f.Name)
		if name == "-" {
			continue
		}

		sf := structField{
			name:   name,
			offset: f.Offset,
			kind:   f.Type.Kind(),
		}
		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Map {
			sf.elem = f.Type.Elem()
		} else if f.Type.Kind() == reflect.Struct {
			sf.structType = f.Type
		}
		dec.fields = append(dec.fields, sf)
	}

	return dec
}

// ZeroCopy returns a zero-copy version of this decoder.
// Strings will point directly into the input buffer - no allocations!
// WARNING: Decoded strings are only valid while the input buffer exists.
// Use this when the input buffer outlives the decoded struct (e.g., from database).
func (sd *StructDecoder[T]) ZeroCopy() *StructDecoder[T] {
	sd.zeroCopy = true
	return sd
}

// Decode decodes msgpack data into the struct.
// This is the fast path - single pass, no intermediate allocations.
func (sd *StructDecoder[T]) Decode(data []byte, dst *T) error {
	d := decoderPool.Get().(*Decoder)
	d.Reset(data)
	err := sd.decodeInto(d, unsafe.Pointer(dst))
	decoderPool.Put(d)
	return err
}

// DecodeWith decodes msgpack data using a provided decoder.
// Use this to avoid pool overhead when you manage your own decoder.
// The decoder will be reset with the provided data before decoding.
func (sd *StructDecoder[T]) DecodeWith(d *Decoder, data []byte, dst *T) error {
	d.Reset(data)
	return sd.decodeInto(d, unsafe.Pointer(dst))
}

func (sd *StructDecoder[T]) decodeInto(d *Decoder, ptr unsafe.Pointer) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}

	if format == formatNil {
		return nil
	}

	// Inline map length parsing for performance
	var mapLen int
	if isFixmap(format) {
		mapLen = fixmapLen(format)
	} else if format == formatMap16 {
		n, err := d.readUint16()
		if err != nil {
			return err
		}
		mapLen = int(n)
	} else if format == formatMap32 {
		n, err := d.readUint32()
		if err != nil {
			return err
		}
		mapLen = int(n)
	} else {
		return ErrTypeMismatch
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return err
	}

	// Single pass: read each key-value and match to registered fields
	for i := 0; i < mapLen; i++ {
		// Read key
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return err
		}
		key := unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))

		// Find matching field via linear scan (fastest for typical struct sizes)
		var field *structField
		for i := range sd.fields {
			if sd.fields[i].name == key {
				field = &sd.fields[i]
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

		// Decode directly into struct field
		fieldPtr := unsafe.Add(ptr, field.offset)
		if err := sd.decodeField(d, fieldPtr, field); err != nil {
			return err
		}
	}

	return nil
}

func (sd *StructDecoder[T]) decodeField(d *Decoder, ptr unsafe.Pointer, field *structField) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}

	if format == formatNil {
		return nil // leave as zero value
	}

	switch field.kind {
	case reflect.String:
		s, err := sd.decodeString(d, format)
		if err != nil {
			return err
		}
		*(*string)(ptr) = s

	case reflect.Int:
		v, err := decodeInt(d, format)
		if err != nil {
			return err
		}
		*(*int)(ptr) = int(v)

	case reflect.Int64:
		v, err := decodeInt(d, format)
		if err != nil {
			return err
		}
		*(*int64)(ptr) = v

	case reflect.Int32:
		v, err := decodeInt(d, format)
		if err != nil {
			return err
		}
		*(*int32)(ptr) = int32(v)

	case reflect.Int16:
		v, err := decodeInt(d, format)
		if err != nil {
			return err
		}
		*(*int16)(ptr) = int16(v)

	case reflect.Int8:
		v, err := decodeInt(d, format)
		if err != nil {
			return err
		}
		*(*int8)(ptr) = int8(v)

	case reflect.Uint:
		v, err := decodeUint(d, format)
		if err != nil {
			return err
		}
		*(*uint)(ptr) = uint(v)

	case reflect.Uint64:
		v, err := decodeUint(d, format)
		if err != nil {
			return err
		}
		*(*uint64)(ptr) = v

	case reflect.Uint32:
		v, err := decodeUint(d, format)
		if err != nil {
			return err
		}
		*(*uint32)(ptr) = uint32(v)

	case reflect.Uint16:
		v, err := decodeUint(d, format)
		if err != nil {
			return err
		}
		*(*uint16)(ptr) = uint16(v)

	case reflect.Uint8:
		v, err := decodeUint(d, format)
		if err != nil {
			return err
		}
		*(*uint8)(ptr) = uint8(v)

	case reflect.Float64:
		v, err := decodeFloat(d, format)
		if err != nil {
			return err
		}
		*(*float64)(ptr) = v

	case reflect.Float32:
		v, err := decodeFloat(d, format)
		if err != nil {
			return err
		}
		*(*float32)(ptr) = float32(v)

	case reflect.Bool:
		if format == formatTrue {
			*(*bool)(ptr) = true
		} else if format == formatFalse {
			*(*bool)(ptr) = false
		} else {
			return ErrTypeMismatch
		}

	case reflect.Slice:
		switch field.elem.Kind() {
		case reflect.String:
			arr, err := sd.decodeStringSlice(d, format)
			if err != nil {
				return err
			}
			*(*[]string)(ptr) = arr
		case reflect.Uint8:
			b, err := sd.decodeBytes(d, format)
			if err != nil {
				return err
			}
			*(*[]byte)(ptr) = b
		case reflect.Int64:
			arr, err := sd.decodeInt64Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]int64)(ptr) = arr
		case reflect.Int:
			arr, err := sd.decodeIntSlice(d, format)
			if err != nil {
				return err
			}
			*(*[]int)(ptr) = arr
		case reflect.Int32:
			arr, err := sd.decodeInt32Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]int32)(ptr) = arr
		case reflect.Int16:
			arr, err := sd.decodeInt16Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]int16)(ptr) = arr
		case reflect.Int8:
			arr, err := sd.decodeInt8Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]int8)(ptr) = arr
		case reflect.Uint64:
			arr, err := sd.decodeUint64Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]uint64)(ptr) = arr
		case reflect.Uint:
			arr, err := sd.decodeUintSlice(d, format)
			if err != nil {
				return err
			}
			*(*[]uint)(ptr) = arr
		case reflect.Uint32:
			arr, err := sd.decodeUint32Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]uint32)(ptr) = arr
		case reflect.Uint16:
			arr, err := sd.decodeUint16Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]uint16)(ptr) = arr
		case reflect.Float64:
			arr, err := sd.decodeFloat64Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]float64)(ptr) = arr
		case reflect.Float32:
			arr, err := sd.decodeFloat32Slice(d, format)
			if err != nil {
				return err
			}
			*(*[]float32)(ptr) = arr
		default:
			d.pos--
			if _, err := d.Decode(); err != nil {
				return err
			}
		}

	case reflect.Map:
		switch field.elem.Kind() {
		case reflect.String:
			m, err := sd.decodeStringMap(d, format)
			if err != nil {
				return err
			}
			*(*map[string]string)(ptr) = m
		case reflect.Interface:
			// map[string]any - use the generic map decoder
			m, err := sd.decodeMapStringAny(d, format)
			if err != nil {
				return err
			}
			*(*map[string]any)(ptr) = m
		case reflect.Map:
			// Nested map like map[string]map[string][]byte
			m, err := sd.decodeNestedMap(d, format, field.elem)
			if err != nil {
				return err
			}
			reflect.NewAt(reflect.MapOf(reflect.TypeOf(""), field.elem), ptr).Elem().Set(reflect.ValueOf(m))
		default:
			d.pos--
			if _, err := d.Decode(); err != nil {
				return err
			}
		}

	case reflect.Struct:
		d.pos--
		if err := sd.decodeNestedStruct(d, ptr, field.structType); err != nil {
			return err
		}

	default:
		d.pos--
		if _, err := d.Decode(); err != nil {
			return err
		}
	}

	return nil
}

func (sd *StructDecoder[T]) decodeString(d *Decoder, format byte) (string, error) {
	// Inline string length parsing for performance
	var length int
	if isFixstr(format) {
		length = fixstrLen(format)
	} else {
		switch format {
		case formatStr8:
			n, err := d.readUint8()
			if err != nil {
				return "", err
			}
			length = int(n)
		case formatStr16:
			n, err := d.readUint16()
			if err != nil {
				return "", err
			}
			length = int(n)
		case formatStr32:
			n, err := d.readUint32()
			if err != nil {
				return "", err
			}
			length = int(n)
		default:
			return "", ErrTypeMismatch
		}
	}
	if err := d.validateStringLen(length); err != nil {
		return "", err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	if sd.zeroCopy {
		// Zero-copy: string points into input buffer
		return unsafe.String(unsafe.SliceData(bytes), len(bytes)), nil
	}
	// Copy string for safety
	return string(bytes), nil
}

func (sd *StructDecoder[T]) decodeStringSlice(d *Decoder, format byte) ([]string, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}

	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}

	result := make([]string, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		result[i], err = sd.decodeString(d, f)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeInt64Slice(d *Decoder, format byte) ([]int64, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]int64, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		result[i], err = d.decodeInt(f)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeIntSlice(d *Decoder, format byte) ([]int, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]int, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeInt(f)
		if err != nil {
			return nil, err
		}
		result[i] = int(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeInt32Slice(d *Decoder, format byte) ([]int32, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]int32, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeInt(f)
		if err != nil {
			return nil, err
		}
		result[i] = int32(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeInt16Slice(d *Decoder, format byte) ([]int16, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]int16, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeInt(f)
		if err != nil {
			return nil, err
		}
		result[i] = int16(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeInt8Slice(d *Decoder, format byte) ([]int8, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]int8, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeInt(f)
		if err != nil {
			return nil, err
		}
		result[i] = int8(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeUint64Slice(d *Decoder, format byte) ([]uint64, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]uint64, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		result[i], err = d.decodeUint(f)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeUintSlice(d *Decoder, format byte) ([]uint, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]uint, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeUint(f)
		if err != nil {
			return nil, err
		}
		result[i] = uint(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeUint32Slice(d *Decoder, format byte) ([]uint32, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]uint32, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeUint(f)
		if err != nil {
			return nil, err
		}
		result[i] = uint32(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeUint16Slice(d *Decoder, format byte) ([]uint16, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]uint16, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeUint(f)
		if err != nil {
			return nil, err
		}
		result[i] = uint16(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeFloat64Slice(d *Decoder, format byte) ([]float64, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]float64, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		result[i], err = d.decodeFloat(f)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeFloat32Slice(d *Decoder, format byte) ([]float32, error) {
	arrLen, err := d.parseArrayLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateArrayLen(arrLen); err != nil {
		return nil, err
	}
	result := make([]float32, arrLen)
	for i := 0; i < arrLen; i++ {
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		v, err := d.decodeFloat(f)
		if err != nil {
			return nil, err
		}
		result[i] = float32(v)
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeStringMap(d *Decoder, format byte) (map[string]string, error) {
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return nil, err
	}

	if err := d.validateMapLen(mapLen); err != nil {
		return nil, err
	}

	result := make(map[string]string, mapLen)
	for i := 0; i < mapLen; i++ {
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return nil, err
		}
		var key string
		if sd.zeroCopy {
			key = unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))
		} else {
			key = string(keyBytes)
		}

		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		val, err := sd.decodeString(d, f)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func (sd *StructDecoder[T]) decodeMapStringAny(d *Decoder, format byte) (map[string]any, error) {
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateMapLen(mapLen); err != nil {
		return nil, err
	}
	return decodeMapStringAnyWithLen(d, mapLen, sd.zeroCopy)
}

func (sd *StructDecoder[T]) decodeNestedMap(d *Decoder, format byte, elemType reflect.Type) (any, error) {
	mapLen, err := d.parseMapLen(format)
	if err != nil {
		return nil, err
	}
	if err := d.validateMapLen(mapLen); err != nil {
		return nil, err
	}

	// Create a new map of the appropriate type
	mapType := reflect.MapOf(reflect.TypeOf(""), elemType)
	result := reflect.MakeMapWithSize(mapType, mapLen)

	for i := 0; i < mapLen; i++ {
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return nil, err
		}
		var key string
		if sd.zeroCopy {
			key = unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))
		} else {
			key = string(keyBytes)
		}

		// Decode the nested map value
		f, err := d.readByte()
		if err != nil {
			return nil, err
		}
		nestedLen, err := d.parseMapLen(f)
		if err != nil {
			return nil, err
		}

		// Create nested map
		nestedMap := reflect.MakeMapWithSize(elemType, nestedLen)
		nestedElemType := elemType.Elem()

		for j := 0; j < nestedLen; j++ {
			nestedKeyBytes, err := d.readStringBytes()
			if err != nil {
				return nil, err
			}
			var nestedKey string
			if sd.zeroCopy {
				nestedKey = unsafe.String(unsafe.SliceData(nestedKeyBytes), len(nestedKeyBytes))
			} else {
				nestedKey = string(nestedKeyBytes)
			}

			// Decode the value based on type
			val, err := sd.decodeNestedMapValue(d, nestedElemType)
			if err != nil {
				return nil, err
			}
			nestedMap.SetMapIndex(reflect.ValueOf(nestedKey), reflect.ValueOf(val))
		}

		result.SetMapIndex(reflect.ValueOf(key), nestedMap)
	}
	return result.Interface(), nil
}

func (sd *StructDecoder[T]) decodeNestedMapValue(d *Decoder, elemType reflect.Type) (any, error) {
	f, err := d.readByte()
	if err != nil {
		return nil, err
	}

	switch elemType.Kind() {
	case reflect.Slice:
		if elemType.Elem().Kind() == reflect.Uint8 {
			// []byte
			b, err := sd.decodeBytes(d, f)
			return b, err
		}
	case reflect.String:
		return sd.decodeString(d, f)
	case reflect.Int64:
		return d.decodeInt(f)
	case reflect.Interface:
		d.pos--
		return decodeAnyValue(d, sd.zeroCopy)
	}

	// Fall back to generic decode
	d.pos--
	v, err := d.Decode()
	if err != nil {
		return nil, err
	}
	return convertValue(v), nil
}

func convertValue(v Value) any {
	switch v.Type {
	case TypeString:
		return string(v.Bytes)
	case TypeBinary:
		return v.Bytes
	case TypeArray:
		result := make([]any, len(v.Array))
		for i, item := range v.Array {
			result[i] = convertValue(item)
		}
		return result
	case TypeMap:
		result := make(map[string]any, len(v.Map))
		for _, kv := range v.Map {
			result[string(kv.Key)] = convertValue(kv.Value)
		}
		return result
	case TypeBool:
		return v.Bool
	case TypeFloat32:
		return v.Float32
	case TypeFloat64:
		return v.Float64
	case TypeUint:
		return v.Uint
	case TypeInt:
		return v.Int
	default:
		return nil
	}
}

// Helper functions for decoding primitives
func decodeInt(d *Decoder, format byte) (int64, error) {
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

func decodeUint(d *Decoder, format byte) (uint64, error) {
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
	default:
		return 0, ErrTypeMismatch
	}
}

func decodeFloat(d *Decoder, format byte) (float64, error) {
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

func (sd *StructDecoder[T]) decodeBytes(d *Decoder, format byte) ([]byte, error) {
	var length int
	if isFixstr(format) {
		length = fixstrLen(format)
	} else {
		switch format {
		case formatStr8, formatBin8:
			n, err := d.readUint8()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr16, formatBin16:
			n, err := d.readUint16()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr32, formatBin32:
			n, err := d.readUint32()
			if err != nil {
				return nil, err
			}
			length = int(n)
		default:
			return nil, ErrTypeMismatch
		}
	}
	if err := d.validateBinaryLen(length); err != nil {
		return nil, err
	}
	bytes, err := d.readBytes(length)
	if err != nil {
		return nil, err
	}
	if sd.zeroCopy {
		// Zero-copy: slice points into input buffer
		return bytes, nil
	}
	// Copy for safety
	cp := make([]byte, length)
	copy(cp, bytes)
	return cp, nil
}

// decodeNestedStruct decodes a msgpack map into a nested struct using reflection.
func (sd *StructDecoder[T]) decodeNestedStruct(d *Decoder, ptr unsafe.Pointer, structType reflect.Type) error {
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

	// Get cached field map for the nested struct type
	fieldMap := getNestedFieldMap(structType)

	// Decode each key-value pair
	for i := 0; i < mapLen; i++ {
		keyBytes, err := d.readStringBytes()
		if err != nil {
			return err
		}
		key := unsafe.String(unsafe.SliceData(keyBytes), len(keyBytes))

		f, ok := fieldMap[key]
		if !ok {
			// Skip unknown field
			if _, err := d.Decode(); err != nil {
				return err
			}
			continue
		}

		// Decode into the nested field
		fieldPtr := unsafe.Add(ptr, f.Offset)
		nestedField := structField{
			name:   key,
			offset: f.Offset,
			kind:   f.Type.Kind(),
		}
		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Map {
			nestedField.elem = f.Type.Elem()
		} else if f.Type.Kind() == reflect.Struct {
			nestedField.structType = f.Type
		}

		format, err := d.readByte()
		if err != nil {
			return err
		}
		if format == formatNil {
			continue
		}

		switch nestedField.kind {
		case reflect.String:
			s, err := sd.decodeString(d, format)
			if err != nil {
				return err
			}
			*(*string)(fieldPtr) = s
		case reflect.Int:
			v, err := decodeInt(d, format)
			if err != nil {
				return err
			}
			*(*int)(fieldPtr) = int(v)
		case reflect.Int64:
			v, err := decodeInt(d, format)
			if err != nil {
				return err
			}
			*(*int64)(fieldPtr) = v
		case reflect.Int32:
			v, err := decodeInt(d, format)
			if err != nil {
				return err
			}
			*(*int32)(fieldPtr) = int32(v)
		case reflect.Uint:
			v, err := decodeUint(d, format)
			if err != nil {
				return err
			}
			*(*uint)(fieldPtr) = uint(v)
		case reflect.Uint64:
			v, err := decodeUint(d, format)
			if err != nil {
				return err
			}
			*(*uint64)(fieldPtr) = v
		case reflect.Float64:
			v, err := decodeFloat(d, format)
			if err != nil {
				return err
			}
			*(*float64)(fieldPtr) = v
		case reflect.Float32:
			v, err := decodeFloat(d, format)
			if err != nil {
				return err
			}
			*(*float32)(fieldPtr) = float32(v)
		case reflect.Bool:
			if format == formatTrue {
				*(*bool)(fieldPtr) = true
			} else if format == formatFalse {
				*(*bool)(fieldPtr) = false
			}
		case reflect.Struct:
			d.pos--
			if err := sd.decodeNestedStruct(d, fieldPtr, nestedField.structType); err != nil {
				return err
			}
		default:
			d.pos--
			if _, err := d.Decode(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Global decoder cache for common struct types
var structDecoderCache sync.Map
var structDecoderZeroCopyCache sync.Map

// Cache for nested struct field maps (keyed by reflect.Type)
var nestedFieldMapCache sync.Map

// getNestedFieldMap returns a cached field map for the given struct type.
func getNestedFieldMap(structType reflect.Type) map[string]reflect.StructField {
	if cached, ok := nestedFieldMapCache.Load(structType); ok {
		return cached.(map[string]reflect.StructField)
	}
	fieldMap := buildNestedFieldMap(structType)
	nestedFieldMapCache.Store(structType, fieldMap)
	return fieldMap
}

// GetStructDecoder returns a cached decoder for type T.
// Creates one if it doesn't exist.
// If zeroCopy is true, strings point directly into the input buffer and are
// only valid while the buffer exists.
func GetStructDecoder[T any](zeroCopy bool) *StructDecoder[T] {
	var zero T
	key := reflect.TypeOf(zero)
	if zeroCopy {
		if dec, ok := structDecoderZeroCopyCache.Load(key); ok {
			return dec.(*StructDecoder[T])
		}
		dec := newStructDecoder[T]().ZeroCopy()
		structDecoderZeroCopyCache.Store(key, dec)
		return dec
	}
	if dec, ok := structDecoderCache.Load(key); ok {
		return dec.(*StructDecoder[T])
	}
	dec := newStructDecoder[T]()
	structDecoderCache.Store(key, dec)
	return dec
}
