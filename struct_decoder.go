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
	name   string       // msgpack field name
	offset uintptr      // field offset in struct
	kind   reflect.Kind // field type
	elem   reflect.Type // for slices/maps: element type
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

		name := f.Tag.Get("msgpack")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		// Handle "name,omitempty" style tags
		for i := 0; i < len(name); i++ {
			if name[i] == ',' {
				name = name[:i]
				break
			}
		}

		sf := structField{
			name:   name,
			offset: f.Offset,
			kind:   f.Type.Kind(),
		}
		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Map {
			sf.elem = f.Type.Elem()
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

func (sd *StructDecoder[T]) decodeInto(d *Decoder, ptr unsafe.Pointer) error {
	format, err := d.readByte()
	if err != nil {
		return err
	}

	var mapLen int
	if IsFixmap(format) {
		mapLen = FixmapLen(format)
	} else if format == FormatMap16 {
		n, err := d.readUint16()
		if err != nil {
			return err
		}
		mapLen = int(n)
	} else if format == FormatMap32 {
		n, err := d.readUint32()
		if err != nil {
			return err
		}
		mapLen = int(n)
	} else if format == FormatNil {
		return nil
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

		// Find matching field (linear scan - fast for small n)
		var field *structField
		for j := range sd.fields {
			if sd.fields[j].name == key {
				field = &sd.fields[j]
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

	if format == FormatNil {
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
		if format == FormatTrue {
			*(*bool)(ptr) = true
		} else if format == FormatFalse {
			*(*bool)(ptr) = false
		} else {
			return ErrTypeMismatch
		}

	case reflect.Slice:
		if field.elem.Kind() == reflect.String {
			// []string
			arr, err := sd.decodeStringSlice(d, format)
			if err != nil {
				return err
			}
			*(*[]string)(ptr) = arr
		} else if field.elem.Kind() == reflect.Uint8 {
			// []byte
			b, err := sd.decodeBytes(d, format)
			if err != nil {
				return err
			}
			*(*[]byte)(ptr) = b
		} else {
			// Generic slice - skip for now
			d.pos-- // put format byte back
			if _, err := d.Decode(); err != nil {
				return err
			}
		}

	case reflect.Map:
		if field.elem.Kind() == reflect.String {
			// map[string]string
			m, err := sd.decodeStringMap(d, format)
			if err != nil {
				return err
			}
			*(*map[string]string)(ptr) = m
		} else {
			// Generic map - skip for now
			d.pos--
			if _, err := d.Decode(); err != nil {
				return err
			}
		}

	default:
		// Unknown type - skip
		d.pos--
		if _, err := d.Decode(); err != nil {
			return err
		}
	}

	return nil
}

func (sd *StructDecoder[T]) decodeString(d *Decoder, format byte) (string, error) {
	var length int
	if IsFixstr(format) {
		length = FixstrLen(format)
	} else {
		switch format {
		case FormatStr8:
			n, err := d.readUint8()
			if err != nil {
				return "", err
			}
			length = int(n)
		case FormatStr16:
			n, err := d.readUint16()
			if err != nil {
				return "", err
			}
			length = int(n)
		case FormatStr32:
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
	var arrLen int
	if IsFixarray(format) {
		arrLen = FixarrayLen(format)
	} else if format == FormatArray16 {
		n, err := d.readUint16()
		if err != nil {
			return nil, err
		}
		arrLen = int(n)
	} else if format == FormatArray32 {
		n, err := d.readUint32()
		if err != nil {
			return nil, err
		}
		arrLen = int(n)
	} else {
		return nil, ErrTypeMismatch
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

func (sd *StructDecoder[T]) decodeStringMap(d *Decoder, format byte) (map[string]string, error) {
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
	} else {
		return nil, ErrTypeMismatch
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

// Helper functions for decoding primitives
func decodeInt(d *Decoder, format byte) (int64, error) {
	if IsPositiveFixint(format) {
		return int64(format), nil
	}
	if IsNegativeFixint(format) {
		return int64(int8(format)), nil
	}
	switch format {
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
	default:
		return 0, ErrTypeMismatch
	}
}

func decodeUint(d *Decoder, format byte) (uint64, error) {
	if IsPositiveFixint(format) {
		return uint64(format), nil
	}
	switch format {
	case FormatUint8:
		v, err := d.readUint8()
		return uint64(v), err
	case FormatUint16:
		v, err := d.readUint16()
		return uint64(v), err
	case FormatUint32:
		v, err := d.readUint32()
		return uint64(v), err
	case FormatUint64:
		return d.readUint64()
	default:
		return 0, ErrTypeMismatch
	}
}

func decodeFloat(d *Decoder, format byte) (float64, error) {
	if IsPositiveFixint(format) {
		return float64(format), nil
	}
	if IsNegativeFixint(format) {
		return float64(int8(format)), nil
	}
	switch format {
	case FormatFloat32:
		v, err := d.readFloat32()
		return float64(v), err
	case FormatFloat64:
		return d.readFloat64()
	case FormatUint8:
		v, err := d.readUint8()
		return float64(v), err
	case FormatUint16:
		v, err := d.readUint16()
		return float64(v), err
	case FormatUint32:
		v, err := d.readUint32()
		return float64(v), err
	case FormatInt8:
		v, err := d.readInt8()
		return float64(v), err
	case FormatInt16:
		v, err := d.readInt16()
		return float64(v), err
	case FormatInt32:
		v, err := d.readInt32()
		return float64(v), err
	case FormatInt64:
		v, err := d.readInt64()
		return float64(v), err
	default:
		return 0, ErrTypeMismatch
	}
}

func (sd *StructDecoder[T]) decodeBytes(d *Decoder, format byte) ([]byte, error) {
	var length int
	if IsFixstr(format) {
		length = FixstrLen(format)
	} else {
		switch format {
		case FormatStr8, FormatBin8:
			n, err := d.readUint8()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case FormatStr16, FormatBin16:
			n, err := d.readUint16()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case FormatStr32, FormatBin32:
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

// Global decoder cache for common struct types
var structDecoderCache sync.Map
var structDecoderZeroCopyCache sync.Map

// GetStructDecoder returns a cached decoder for type T.
// Creates one if it doesn't exist.
func GetStructDecoder[T any]() *StructDecoder[T] {
	var zero T
	key := reflect.TypeOf(zero)
	if dec, ok := structDecoderCache.Load(key); ok {
		return dec.(*StructDecoder[T])
	}
	dec := newStructDecoder[T]()
	structDecoderCache.Store(key, dec)
	return dec
}

// GetStructDecoderZeroCopy returns a cached zero-copy decoder for type T.
// Zero-copy strings point into the input buffer - only valid while buffer exists.
func GetStructDecoderZeroCopy[T any]() *StructDecoder[T] {
	var zero T
	key := reflect.TypeOf(zero)
	if dec, ok := structDecoderZeroCopyCache.Load(key); ok {
		return dec.(*StructDecoder[T])
	}
	dec := newStructDecoder[T]().ZeroCopy()
	structDecoderZeroCopyCache.Store(key, dec)
	return dec
}

