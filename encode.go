package msgpck

import (
	"reflect"
	"sync"
)

// Encode encodes any Go value to msgpack.
func (e *Encoder) Encode(v any) error {
	// Fast paths for common types to avoid reflection
	switch val := v.(type) {
	case nil:
		e.EncodeNil()
		return nil
	case bool:
		e.EncodeBool(val)
		return nil
	case int:
		e.EncodeInt(int64(val))
		return nil
	case int64:
		e.EncodeInt(val)
		return nil
	case int32:
		e.EncodeInt(int64(val))
		return nil
	case uint:
		e.EncodeUint(uint64(val))
		return nil
	case uint64:
		e.EncodeUint(val)
		return nil
	case uint32:
		e.EncodeUint(uint64(val))
		return nil
	case float64:
		e.EncodeFloat64(val)
		return nil
	case float32:
		e.EncodeFloat32(val)
		return nil
	case string:
		e.EncodeString(val)
		return nil
	case []byte:
		e.EncodeBinary(val)
		return nil
	case map[string]any:
		return e.encodeMapStringAny(val)
	case []any:
		return e.encodeSliceAny(val)
	}
	// Fall back to reflection for other types
	return e.encodeValue(reflect.ValueOf(v))
}

// encodeMapStringAny is a fast path for map[string]any
func (e *Encoder) encodeMapStringAny(m map[string]any) error {
	e.EncodeMapHeader(len(m))
	for k, v := range m {
		e.EncodeString(k)
		if err := e.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

// encodeSliceAny is a fast path for []any
func (e *Encoder) encodeSliceAny(s []any) error {
	e.EncodeArrayHeader(len(s))
	for _, v := range s {
		if err := e.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

// encodeValue encodes a reflect.Value
func (e *Encoder) encodeValue(rv reflect.Value) error {
	// Handle nil interface
	if !rv.IsValid() {
		e.EncodeNil()
		return nil
	}

	// Dereference pointers
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			e.EncodeNil()
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		e.EncodeBool(rv.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.EncodeInt(rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.EncodeUint(rv.Uint())

	case reflect.Float32:
		e.EncodeFloat32(float32(rv.Float()))

	case reflect.Float64:
		e.EncodeFloat64(rv.Float())

	case reflect.String:
		e.EncodeString(rv.String())

	case reflect.Slice:
		if rv.IsNil() {
			e.EncodeNil()
			return nil
		}
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			// []byte - encode as binary
			e.EncodeBinary(rv.Bytes())
		} else {
			return e.encodeSlice(rv)
		}

	case reflect.Array:
		return e.encodeArray(rv)

	case reflect.Map:
		if rv.IsNil() {
			e.EncodeNil()
			return nil
		}
		return e.encodeMap(rv)

	case reflect.Struct:
		return e.encodeStruct(rv)

	case reflect.Interface:
		if rv.IsNil() {
			e.EncodeNil()
			return nil
		}
		return e.encodeValue(rv.Elem())

	default:
		return ErrUnsupportedType
	}

	return nil
}

// encodeSlice encodes a slice (non-[]byte)
func (e *Encoder) encodeSlice(rv reflect.Value) error {
	length := rv.Len()
	e.EncodeArrayHeader(length)
	for i := 0; i < length; i++ {
		if err := e.encodeValue(rv.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

// encodeArray encodes a fixed-size array (same logic as slice)
func (e *Encoder) encodeArray(rv reflect.Value) error {
	return e.encodeSlice(rv)
}

// encodeMap encodes a map
func (e *Encoder) encodeMap(rv reflect.Value) error {
	e.EncodeMapHeader(rv.Len())

	// Use MapRange to avoid allocating a keys slice
	iter := rv.MapRange()
	for iter.Next() {
		if err := e.encodeValue(iter.Key()); err != nil {
			return err
		}
		if err := e.encodeValue(iter.Value()); err != nil {
			return err
		}
	}
	return nil
}

// encodeStruct encodes a struct as a map
func (e *Encoder) encodeStruct(rv reflect.Value) error {
	info := getStructInfo(rv.Type())

	// Count non-empty fields
	fields := make([]int, 0, len(info.fields))
	for i, fi := range info.fields {
		fv := rv.FieldByIndex(fi.index)
		if fi.omitempty && isEmptyValue(fv) {
			continue
		}
		fields = append(fields, i)
	}

	e.EncodeMapHeader(len(fields))

	for _, idx := range fields {
		fi := &info.fields[idx]
		e.EncodeStringBytes(fi.name)
		if err := e.encodeValue(rv.FieldByIndex(fi.index)); err != nil {
			return err
		}
	}

	return nil
}

// isEmptyValue returns true if the value is considered empty for omitempty
func isEmptyValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Map:
		return rv.IsNil() || rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Array:
		return rv.Len() == 0
	default:
		return false
	}
}

// Thread-local encoders using a pool with get/put pattern
var encoderPool = sync.Pool{
	New: func() any { return NewEncoder(256) },
}

// Marshal encodes a Go value to msgpack bytes.
// The returned bytes are a copy and safe to retain.
func Marshal(v any) ([]byte, error) {
	e := encoderPool.Get().(*Encoder)
	e.Reset()

	if err := e.Encode(v); err != nil {
		encoderPool.Put(e)
		return nil, err
	}

	// Copy result before returning encoder to pool to avoid race conditions
	result := make([]byte, len(e.buf))
	copy(result, e.buf)
	encoderPool.Put(e)
	return result, nil
}
