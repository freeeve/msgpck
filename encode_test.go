package msgpck

import (
	"bytes"
	"math"
	"reflect"
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
		t.Errorf(errMsgNilFmt, b[0])
	}
}

func TestEncodeValueNilSlice(t *testing.T) {
	var s []int
	b, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf(errMsgNilFmt, b[0])
	}
}

func TestEncodeValueNilPointer(t *testing.T) {
	var p *int
	b, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf(errMsgNilFmt, b[0])
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

// TestMarshalVariants tests all Marshal variants
func TestMarshalVariants(t *testing.T) {
	data := map[string]any{"key": "value"}

	t.Run("Marshal", func(t *testing.T) {
		_, err := Marshal(data)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("MarshalCopy", func(t *testing.T) {
		b, err := MarshalCopy(data)
		if err != nil {
			t.Error(err)
		}
		// Verify we can use it after return
		_, err = Unmarshal(b)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("MarshalAppend", func(t *testing.T) {
		prefix := []byte{0x01, 0x02, 0x03}
		b, err := MarshalAppend(prefix, data)
		if err != nil {
			t.Error(err)
		}
		if !bytes.HasPrefix(b, prefix) {
			t.Error("prefix not preserved")
		}
	})
}

// TestEncodeCollectionTypes tests encoding of arrays, slices, and maps
func TestEncodeCollectionTypes(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		b, err := Marshal(arr)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("array encode failed")
		}
	})

	t.Run("slice of strings", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		b, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("slice encode failed")
		}
	})

	t.Run("map[string]int", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		b, err := Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 2 {
			t.Error("map encode failed")
		}
	})
}

// TestEncodePointerTypes tests encoding of pointer types
func TestEncodePointerTypes(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		val := 42
		ptr := &val
		b, err := Marshal(ptr)
		if err != nil {
			t.Fatal(err)
		}
		decoded, _ := Unmarshal(b)
		if decoded != int64(42) {
			t.Error("pointer encode failed")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *int
		b, err := Marshal(ptr)
		if err != nil {
			t.Fatal(err)
		}
		decoded, _ := Unmarshal(b)
		if decoded != nil {
			t.Error("nil pointer encode failed")
		}
	})
}

// TestEncodeNilTypes tests encoding of nil slices and maps
func TestEncodeNilTypes(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var s []int
		b, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		decoded, _ := Unmarshal(b)
		if decoded != nil {
			t.Error("nil slice encode failed")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var m map[string]int
		b, err := Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		decoded, _ := Unmarshal(b)
		if decoded != nil {
			t.Error("nil map encode failed")
		}
	})
}

// TestRoundTripPrimitives tests encoding and decoding of primitive types
func TestRoundTripPrimitives(t *testing.T) {
	// Note: DecodeAny normalizes all integers to int64 (unless > MaxInt64)
	// and float32 to float64 for consistent types
	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"true", true},
		{"false", false},
		{"zero", int64(0)},
		{"positive fixint", int64(42)},
		{"max fixint", int64(127)},
		{"negative fixint", int64(-1)},
		{"min fixint", int64(-32)},
		{"uint8", int64(200)},
		{"uint16", int64(1000)},
		{"uint32", int64(100000)},
		{"uint64", int64(1 << 40)},
		{"int8", int64(-100)},
		{"int16", int64(-1000)},
		{"int32", int64(-100000)},
		{"int64", int64(-1 << 40)},
		{"float32", float64(3.140000104904175)}, // float32 promoted to float64
		{"float64", float64(3.14159265359)},
		{"empty string", ""},
		{"short string", "hello"},
		{"fixstr max", string(make([]byte, 31))},
		{"str8", string(make([]byte, 100))},
		{"str16", string(make([]byte, 300))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf("got %v (%T), want %v (%T)", decoded, decoded, tt.value, tt.value)
			}
		})
	}
}

// TestRoundTripContainers tests arrays and maps
func TestRoundTripContainers(t *testing.T) {
	// Note: all integers decode to int64
	tests := []struct {
		name  string
		value any
	}{
		{"empty array", []any{}},
		{"int array", []any{int64(1), int64(2), int64(3)}},
		{"mixed array", []any{int64(1), "hello", true, nil}},
		{"empty map", map[string]any{}},
		{"string map", map[string]any{"a": int64(1), "b": int64(2)}},
		{"nested map", map[string]any{
			"inner": map[string]any{"x": int64(10)},
		}},
		{"nested array", []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf(errMsgGotWant, decoded, tt.value)
			}
		})
	}
}

// TestBinaryData tests encoding/decoding of binary data
func TestBinaryData(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	decoded, err := Unmarshal(encoded)
	if err != nil {
		t.Fatalf(errMsgUnmarshalFailed, err)
	}

	if !bytes.Equal(decoded.([]byte), original) {
		t.Errorf(errMsgGotWant, decoded, original)
	}
}

// TestFloatSpecialValues tests special float values
func TestFloatSpecialValues(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
		{"NaN", math.NaN()},
		{"max float64", math.MaxFloat64},
		{"smallest positive", math.SmallestNonzeroFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			decoded, err := Unmarshal(encoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			got := decoded.(float64)
			if math.IsNaN(tt.value) {
				if !math.IsNaN(got) {
					t.Errorf("expected NaN, got %v", got)
				}
			} else if got != tt.value {
				t.Errorf("got %v, want %v", got, tt.value)
			}
		})
	}
}
