package msgpck

import (
	"testing"
)

func TestEncodeExtFormats(t *testing.T) {
	tests := []struct {
		size int
	}{
		{1},   // fixext1
		{2},   // fixext2
		{4},   // fixext4
		{8},   // fixext8
		{16},  // fixext16
		{3},   // ext8
		{100}, // ext8
		{256}, // ext16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.size + 10)
		e.EncodeExt(1, make([]byte, tc.size))
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeExt size=%d failed", tc.size)
		}
	}
}

func TestEncodeVariousTypesExtra(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"bool true", true},
		{"bool false", false},
		{"int", 42},
		{"int8", int8(42)},
		{"int16", int16(42)},
		{"int32", int32(42)},
		{"int64", int64(42)},
		{"uint", uint(42)},
		{"uint8", uint8(42)},
		{"uint16", uint16(42)},
		{"uint32", uint32(42)},
		{"uint64", uint64(42)},
		{"float32", float32(3.14)},
		{"float64", float64(3.14)},
		{"string", "hello"},
		{"[]byte", []byte{1, 2, 3}},
		{"[]int", []int{1, 2, 3}},
		{"[]string", []string{"a", "b"}},
		{"[3]int", [3]int{1, 2, 3}},
		{"map[string]int", map[string]int{"a": 1}},
		{"map[string]any", map[string]any{"a": 1, "b": "x"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := MarshalCopy(tc.value)
			if err != nil {
				t.Errorf("%s encode failed: %v", tc.name, err)
			}
			if len(b) == 0 {
				t.Errorf("%s encoded to empty", tc.name)
			}
		})
	}
}

func TestEncodeValueAllTypes(t *testing.T) {
	e := NewEncoder(64)

	tests := []Value{
		{Type: TypeNil},
		{Type: TypeBool, Bool: true},
		{Type: TypeBool, Bool: false},
		{Type: TypeInt, Int: -42},
		{Type: TypeUint, Uint: 42},
		{Type: TypeFloat32, Float32: 3.14},
		{Type: TypeFloat64, Float64: 2.718},
		{Type: TypeString, Bytes: []byte("hello")},
		{Type: TypeBinary, Bytes: []byte{1, 2, 3}},
		{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 1}}},
		{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{Type: TypeInt, Int: 1}}}},
	}

	for _, v := range tests {
		e.Reset()
		e.EncodeValue(&v)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeValue produced no output for type %v", v.Type)
		}
	}
}

func TestEncodeMapWithNonStringKey(t *testing.T) {
	m := map[int]string{1: "one", 2: "two"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with non-string key returned error (expected): %v", err)
	}
}

func TestEncodeArray16Array32(t *testing.T) {
	t.Run("array16", func(t *testing.T) {
		// Array with 16 elements (needs array16 format)
		arr := make([]int, 16)
		for i := range arr {
			arr[i] = i
		}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("array16 encode failed")
		}
	})

	t.Run("large array", func(t *testing.T) {
		// Array with 256 elements
		arr := make([]int, 256)
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("large array encode failed")
		}
	})
}

func TestEncodeMap16(t *testing.T) {
	m := make(map[string]int)
	for i := 0; i < 20; i++ {
		m[string(rune('a'+i))] = i
	}
	b, err := Marshal(m)
	if err != nil || len(b) == 0 {
		t.Error("map16 encode failed")
	}
}

func TestMarshalAppend(t *testing.T) {
	prefix := []byte{0x01, 0x02}
	result, err := MarshalAppend(prefix, "hello")
	if err != nil {
		t.Error("MarshalAppend failed")
	}
	if len(result) < 3 || result[0] != 0x01 || result[1] != 0x02 {
		t.Error("MarshalAppend didn't preserve prefix")
	}
}

func TestEncodeStructOmitempty(t *testing.T) {
	type Data struct {
		Name  string `msgpack:"name,omitempty"`
		Value int    `msgpack:"value,omitempty"`
		Ptr   *int   `msgpack:"ptr,omitempty"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{} // all zero values
	b, err := enc.Encode(&d)
	if err != nil {
		t.Error("omitempty encode failed")
	}
	// Should encode to empty map
	if len(b) == 0 {
		t.Error("empty result")
	}
}

func TestEncodeSliceArrayPaths(t *testing.T) {
	t.Run("[]any", func(t *testing.T) {
		arr := []any{1, "two", 3.0}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[]any encode failed")
		}
	})

	t.Run("[]float64", func(t *testing.T) {
		arr := []float64{1.1, 2.2, 3.3}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[]float64 encode failed")
		}
	})

	t.Run("[3]string", func(t *testing.T) {
		arr := [3]string{"a", "b", "c"}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[3]string encode failed")
		}
	})
}

func TestEncodeMapPaths(t *testing.T) {
	t.Run("map[string]string", func(t *testing.T) {
		m := map[string]string{"k": "v"}
		b, err := Marshal(m)
		if err != nil || len(b) == 0 {
			t.Error("map[string]string encode failed")
		}
	})

	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"k": 123}
		b, err := Marshal(m)
		if err != nil || len(b) == 0 {
			t.Error("map[string]any encode failed")
		}
	})
}

func TestEncodeStructFields(t *testing.T) {
	type AllTypes struct {
		Bool    bool    `msgpack:"bool"`
		Int     int     `msgpack:"int"`
		Int64   int64   `msgpack:"int64"`
		Uint    uint    `msgpack:"uint"`
		Uint64  uint64  `msgpack:"uint64"`
		Float32 float32 `msgpack:"float32"`
		Float64 float64 `msgpack:"float64"`
		String  string  `msgpack:"string"`
		Bytes   []byte  `msgpack:"bytes"`
	}

	s := AllTypes{
		Bool:    true,
		Int:     -42,
		Int64:   -100,
		Uint:    42,
		Uint64:  100,
		Float32: 3.14,
		Float64: 2.718,
		String:  "test",
		Bytes:   []byte{1, 2, 3},
	}

	enc := GetStructEncoder[AllTypes]()
	b, err := enc.Encode(&s)
	if err != nil || len(b) == 0 {
		t.Error("AllTypes encode failed")
	}
}

func TestMarshalError(t *testing.T) {
	// Unsupported type should error
	ch := make(chan int)
	_, err := Marshal(ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

func TestEncodeMapNonStringKey(t *testing.T) {
	m := map[int]string{1: "one"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with int key returned error (expected): %v", err)
	}
}

func TestEncodeSliceArrayError(t *testing.T) {
	// Slice with unsupported element type
	arr := []chan int{make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for channel slice")
	}
}

func TestEncodeMapError(t *testing.T) {
	// Map with unsupported value type
	m := map[string]chan int{"k": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel map value")
	}
}

func TestMarshalCopyError(t *testing.T) {
	// Use channel which can't be encoded
	ch := make(chan int)
	_, err := MarshalCopy(ch)
	if err == nil {
		t.Error("expected error for channel")
	}
}

func TestMarshalAppendError(t *testing.T) {
	// Use channel which can't be encoded
	ch := make(chan int)
	_, err := MarshalAppend(nil, ch)
	if err == nil {
		t.Error("expected error for channel")
	}
}

func TestMarshalAppendSuccess(t *testing.T) {
	dst := make([]byte, 0, 64)
	result, err := MarshalAppend(dst, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) == 0 {
		t.Error("empty result")
	}
}

func TestEncodeSliceAnyError(t *testing.T) {
	// Slice containing unencodable value
	s := []any{"ok", make(chan int)}
	_, err := Marshal(s)
	if err == nil {
		t.Error("expected error for slice with channel")
	}
}

func TestEncodeMapStringAnyError(t *testing.T) {
	// Map containing unencodable value
	m := map[string]any{"ok": "fine", "bad": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel")
	}
}

func TestEncodeArrayError(t *testing.T) {
	// Array with channel element
	arr := [2]any{"ok", make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for array with channel")
	}
}

func TestEncodeMapChannelKeyError(t *testing.T) {
	// Map with complex key that fails
	m := map[any]any{make(chan int): "value"}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel key")
	}
}

func TestEncodeMapKeyValueError(t *testing.T) {
	// Map with channel value
	m := map[string]any{"ch": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel value")
	}
}

func TestEncodeStringStr32(t *testing.T) {
	// Create a string longer than 65535 bytes to trigger str32 format
	longStr := string(make([]byte, 70000))
	enc := NewEncoder(80000)
	enc.EncodeString(longStr)

	b := enc.Bytes()
	if b[0] != formatStr32 {
		t.Errorf("expected str32 format (0xdb), got 0x%02x", b[0])
	}
}

func TestEncodeBinaryBin32(t *testing.T) {
	// Create binary longer than 65535 bytes to trigger bin32 format
	longBin := make([]byte, 70000)
	enc := NewEncoder(80000)
	enc.EncodeBinary(longBin)

	b := enc.Bytes()
	if b[0] != formatBin32 {
		t.Errorf("expected bin32 format (0xc6), got 0x%02x", b[0])
	}
}

func TestEncodeValueNilMap(t *testing.T) {
	var m map[string]int
	b, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf("expected nil format, got 0x%02x", b[0])
	}
}

func TestEncodeValueNilSlice(t *testing.T) {
	var s []int
	b, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf("expected nil format, got 0x%02x", b[0])
	}
}

func TestEncodeValueNilPointer(t *testing.T) {
	var p *int
	b, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf("expected nil format, got 0x%02x", b[0])
	}
}

func TestEncodeValueFloat32(t *testing.T) {
	var f float32 = 3.14
	b, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatFloat32 {
		t.Errorf("expected float32 format, got 0x%02x", b[0])
	}
}
