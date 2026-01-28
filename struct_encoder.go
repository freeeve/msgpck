package msgpck

import (
	"reflect"
	"sync"
	"unsafe"
)

// StructEncoder is a pre-registered encoder for a specific struct type.
// Create once, reuse many times for best performance.
type StructEncoder[T any] struct {
	fields []encodeField
}

type encodeField struct {
	name      []byte       // msgpack field name (pre-encoded)
	offset    uintptr      // field offset in struct
	kind      reflect.Kind // field type
	typ       reflect.Type // full type (for structs/pointers)
	elem      reflect.Type // for slices/maps: element type
	omitempty bool
}

// newStructEncoder creates a reusable encoder for type T.
// Caches field information for fast repeated encoding.
func newStructEncoder[T any]() *StructEncoder[T] {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	enc := &StructEncoder[T]{
		fields: make([]encodeField, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get("msgpack")
		if tag == "-" {
			continue
		}

		name, omitempty := parseFieldTag(tag, f.Name)

		ef := encodeField{
			name:      []byte(name),
			offset:    f.Offset,
			kind:      f.Type.Kind(),
			typ:       f.Type,
			omitempty: omitempty,
		}
		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Map {
			ef.elem = f.Type.Elem()
		}
		enc.fields = append(enc.fields, ef)
	}

	return enc
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// parseFieldTag parses a msgpack struct tag and returns the field name and omitempty flag.
func parseFieldTag(tag, fieldName string) (string, bool) {
	if tag == "" {
		return fieldName, false
	}
	name := tag
	omitempty := false
	for i := 0; i < len(tag); i++ {
		if tag[i] == ',' {
			name = tag[:i]
			omitempty = contains(tag[i+1:], "omitempty")
			break
		}
	}
	if name == "" {
		name = fieldName
	}
	return name, omitempty
}

// Encode encodes the struct to msgpack bytes.
// The returned bytes are a copy and safe to retain.
// For zero-allocation encoding, use EncodeWith with your own Encoder.
func (se *StructEncoder[T]) Encode(src *T) ([]byte, error) {
	e := encoderPool.Get().(*Encoder)
	e.Reset()

	err := se.encodeInto(e, unsafe.Pointer(src))
	if err != nil {
		encoderPool.Put(e)
		return nil, err
	}

	result := make([]byte, len(e.buf))
	copy(result, e.buf)
	encoderPool.Put(e)
	return result, nil
}

// EncodeWith encodes using a provided encoder.
// Use this to avoid pool overhead when you manage your own encoder.
// The encoder is NOT reset - call e.Reset() before if needed.
func (se *StructEncoder[T]) EncodeWith(e *Encoder, src *T) error {
	return se.encodeInto(e, unsafe.Pointer(src))
}

func (se *StructEncoder[T]) encodeInto(e *Encoder, ptr unsafe.Pointer) error {
	// Count non-empty fields if any have omitempty
	fieldCount := 0
	for i := range se.fields {
		f := &se.fields[i]
		if f.omitempty && se.isEmptyField(ptr, f) {
			continue
		}
		fieldCount++
	}

	// Write map header
	e.EncodeMapHeader(fieldCount)

	// Encode each field
	for i := range se.fields {
		f := &se.fields[i]
		if f.omitempty && se.isEmptyField(ptr, f) {
			continue
		}

		// Write field name
		e.EncodeStringBytes(f.name)

		// Write field value
		fieldPtr := unsafe.Add(ptr, f.offset)
		if err := se.encodeField(e, fieldPtr, f); err != nil {
			return err
		}
	}

	return nil
}

func (se *StructEncoder[T]) isEmptyField(ptr unsafe.Pointer, f *encodeField) bool {
	fieldPtr := unsafe.Add(ptr, f.offset)

	switch f.kind {
	case reflect.String:
		return *(*string)(fieldPtr) == ""
	case reflect.Int, reflect.Int64:
		return *(*int64)(fieldPtr) == 0
	case reflect.Int32:
		return *(*int32)(fieldPtr) == 0
	case reflect.Int16:
		return *(*int16)(fieldPtr) == 0
	case reflect.Int8:
		return *(*int8)(fieldPtr) == 0
	case reflect.Uint, reflect.Uint64:
		return *(*uint64)(fieldPtr) == 0
	case reflect.Uint32:
		return *(*uint32)(fieldPtr) == 0
	case reflect.Uint16:
		return *(*uint16)(fieldPtr) == 0
	case reflect.Uint8:
		return *(*uint8)(fieldPtr) == 0
	case reflect.Float64:
		return *(*float64)(fieldPtr) == 0
	case reflect.Float32:
		return *(*float32)(fieldPtr) == 0
	case reflect.Bool:
		return !*(*bool)(fieldPtr)
	case reflect.Slice:
		// Check slice length via header
		sh := (*sliceHeader)(fieldPtr)
		return sh.Len == 0
	case reflect.Map:
		// Maps are pointers - check if nil or empty
		// This is tricky with unsafe, just return false for now
		return false
	default:
		return false
	}
}

// sliceHeader mirrors reflect.SliceHeader
type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func (se *StructEncoder[T]) encodeField(e *Encoder, ptr unsafe.Pointer, f *encodeField) error {
	switch f.kind {
	case reflect.String:
		e.EncodeString(*(*string)(ptr))

	case reflect.Int:
		e.EncodeInt(int64(*(*int)(ptr)))

	case reflect.Int64:
		e.EncodeInt(*(*int64)(ptr))

	case reflect.Int32:
		e.EncodeInt(int64(*(*int32)(ptr)))

	case reflect.Int16:
		e.EncodeInt(int64(*(*int16)(ptr)))

	case reflect.Int8:
		e.EncodeInt(int64(*(*int8)(ptr)))

	case reflect.Uint:
		e.EncodeUint(uint64(*(*uint)(ptr)))

	case reflect.Uint64:
		e.EncodeUint(*(*uint64)(ptr))

	case reflect.Uint32:
		e.EncodeUint(uint64(*(*uint32)(ptr)))

	case reflect.Uint16:
		e.EncodeUint(uint64(*(*uint16)(ptr)))

	case reflect.Uint8:
		e.EncodeUint(uint64(*(*uint8)(ptr)))

	case reflect.Float64:
		e.EncodeFloat64(*(*float64)(ptr))

	case reflect.Float32:
		e.EncodeFloat32(*(*float32)(ptr))

	case reflect.Bool:
		e.EncodeBool(*(*bool)(ptr))

	case reflect.Slice:
		if f.elem == nil {
			return nil
		}
		switch f.elem.Kind() {
		case reflect.String:
			s := *(*[]string)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeString(v)
			}
		case reflect.Uint8:
			e.EncodeBinary(*(*[]byte)(ptr))
		case reflect.Int64:
			s := *(*[]int64)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeInt(v)
			}
		case reflect.Int:
			s := *(*[]int)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeInt(int64(v))
			}
		case reflect.Int32:
			s := *(*[]int32)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeInt(int64(v))
			}
		case reflect.Int16:
			s := *(*[]int16)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeInt(int64(v))
			}
		case reflect.Int8:
			s := *(*[]int8)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeInt(int64(v))
			}
		case reflect.Uint64:
			s := *(*[]uint64)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeUint(v)
			}
		case reflect.Uint:
			s := *(*[]uint)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeUint(uint64(v))
			}
		case reflect.Uint32:
			s := *(*[]uint32)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeUint(uint64(v))
			}
		case reflect.Uint16:
			s := *(*[]uint16)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeUint(uint64(v))
			}
		case reflect.Float64:
			s := *(*[]float64)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeFloat64(v)
			}
		case reflect.Float32:
			s := *(*[]float32)(ptr)
			e.EncodeArrayHeader(len(s))
			for _, v := range s {
				e.EncodeFloat32(v)
			}
		default:
			// Fallback to reflection for unsupported types
			rv := reflect.NewAt(reflect.SliceOf(f.elem), ptr).Elem()
			e.EncodeArrayHeader(rv.Len())
			for i := 0; i < rv.Len(); i++ {
				if err := e.Encode(rv.Index(i).Interface()); err != nil {
					return err
				}
			}
		}

	case reflect.Map:
		if f.elem == nil {
			return nil
		}
		switch f.elem.Kind() {
		case reflect.String:
			m := *(*map[string]string)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeString(v)
			}
		case reflect.Int:
			m := *(*map[string]int)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeInt(int64(v))
			}
		case reflect.Int64:
			m := *(*map[string]int64)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeInt(v)
			}
		case reflect.Int32:
			m := *(*map[string]int32)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeInt(int64(v))
			}
		case reflect.Uint64:
			m := *(*map[string]uint64)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeUint(v)
			}
		case reflect.Uint32:
			m := *(*map[string]uint32)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeUint(uint64(v))
			}
		case reflect.Float64:
			m := *(*map[string]float64)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeFloat64(v)
			}
		case reflect.Bool:
			m := *(*map[string]bool)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				e.EncodeBool(v)
			}
		case reflect.Interface:
			// map[string]any - encode values based on runtime type
			m := *(*map[string]any)(ptr)
			e.EncodeMapHeader(len(m))
			for k, v := range m {
				e.EncodeString(k)
				if err := e.Encode(v); err != nil {
					return err
				}
			}
		default:
			// Fallback to reflection
			rv := reflect.NewAt(reflect.MapOf(reflect.TypeOf(""), f.elem), ptr).Elem()
			e.EncodeMapHeader(rv.Len())
			iter := rv.MapRange()
			for iter.Next() {
				e.EncodeString(iter.Key().String())
				if err := e.Encode(iter.Value().Interface()); err != nil {
					return err
				}
			}
		}

	case reflect.Struct:
		rv := reflect.NewAt(f.typ, ptr).Elem()
		if err := e.Encode(rv.Interface()); err != nil {
			return err
		}

	case reflect.Ptr:
		ptrVal := *(*unsafe.Pointer)(ptr)
		if ptrVal == nil {
			e.EncodeNil()
		} else {
			rv := reflect.NewAt(f.typ, ptr).Elem()
			if err := e.Encode(rv.Interface()); err != nil {
				return err
			}
		}

	default:
		e.EncodeNil()
	}

	return nil
}

// Global encoder cache
var structEncoderCache sync.Map

// GetStructEncoder returns a cached encoder for type T.
func GetStructEncoder[T any]() *StructEncoder[T] {
	var zero T
	key := reflect.TypeOf(zero)
	if enc, ok := structEncoderCache.Load(key); ok {
		return enc.(*StructEncoder[T])
	}
	enc := newStructEncoder[T]()
	structEncoderCache.Store(key, enc)
	return enc
}
