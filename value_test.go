package msgpck

import (
	"testing"
)

func TestValueMethods2(t *testing.T) {
	t.Run("IsNil", func(t *testing.T) {
		v := Value{Type: TypeNil}
		if !v.IsNil() {
			t.Error("IsNil failed")
		}
	})

	t.Run("AsBool", func(t *testing.T) {
		v := Value{Type: TypeBool, Bool: true}
		if !v.AsBool() {
			t.Error("AsBool failed")
		}
	})

	t.Run("AsString", func(t *testing.T) {
		v := Value{Type: TypeString, Bytes: []byte("test")}
		if v.AsString() != "test" {
			t.Error("AsString failed")
		}
	})

	t.Run("AsBytes", func(t *testing.T) {
		v := Value{Type: TypeBinary, Bytes: []byte{1, 2, 3}}
		if len(v.AsBytes()) != 3 {
			t.Error("AsBytes failed")
		}
	})

	t.Run("Index", func(t *testing.T) {
		v := Value{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 42}}}
		if v.Index(0).Int != 42 {
			t.Error("Index failed")
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		v := Value{Type: TypeMap, Map: []KV{}}
		if v.Get([]byte("missing")) != nil {
			t.Error("Get should return nil for missing key")
		}
	})

	t.Run("Get on non-map", func(t *testing.T) {
		v := Value{Type: TypeInt}
		if v.Get([]byte("key")) != nil {
			t.Error("Get on non-map should return nil")
		}
	})
}

func TestValueToAny(t *testing.T) {
	// Create various Value types and decode them, then use DecodeAny which converts
	tests := []struct {
		name  string
		value Value
	}{
		{"nil", Value{Type: TypeNil}},
		{"bool true", Value{Type: TypeBool, Bool: true}},
		{"bool false", Value{Type: TypeBool, Bool: false}},
		{"int", Value{Type: TypeInt, Int: -42}},
		{"uint", Value{Type: TypeUint, Uint: 42}},
		{"float32", Value{Type: TypeFloat32, Float32: 3.14}},
		{"float64", Value{Type: TypeFloat64, Float64: 2.718}},
		{"string", Value{Type: TypeString, Bytes: []byte("hello")}},
		{"binary", Value{Type: TypeBinary, Bytes: []byte{1, 2, 3}}},
		{"empty array", Value{Type: TypeArray, Array: []Value{}}},
		{"array", Value{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 1}}}},
		{"empty map", Value{Type: TypeMap, Map: []KV{}}},
		{"map", Value{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{Type: TypeInt, Int: 1}}}}},
		{"ext", Value{Type: TypeExt, Ext: Ext{Type: 1, Data: []byte{1}}}},
		{"unknown type", Value{Type: Type(255)}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// This tests valueToAny internally
			_ = valueToAny(&tc.value)
		})
	}
}

func TestValueAsHelpers(t *testing.T) {
	t.Run("AsInt from uint", func(t *testing.T) {
		v := Value{Type: TypeUint, Uint: 42}
		if v.AsInt() != 42 {
			t.Error("AsInt from uint failed")
		}
	})

	t.Run("AsUint from int", func(t *testing.T) {
		v := Value{Type: TypeInt, Int: 42}
		if v.AsUint() != 42 {
			t.Error("AsUint from int failed")
		}
	})

	t.Run("AsFloat64 from int", func(t *testing.T) {
		v := Value{Type: TypeInt, Int: 42}
		if v.AsFloat64() != 42.0 {
			t.Error("AsFloat64 from int failed")
		}
	})

	t.Run("AsFloat64 from uint", func(t *testing.T) {
		v := Value{Type: TypeUint, Uint: 42}
		if v.AsFloat64() != 42.0 {
			t.Error("AsFloat64 from uint failed")
		}
	})

	t.Run("AsFloat64 from float32", func(t *testing.T) {
		v := Value{Type: TypeFloat32, Float32: 3.14}
		if v.AsFloat64() < 3.1 || v.AsFloat64() > 3.2 {
			t.Error("AsFloat64 from float32 failed")
		}
	})

	t.Run("Len of map", func(t *testing.T) {
		v := Value{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{}}}}
		if v.Len() != 1 {
			t.Error("Len of map failed")
		}
	})
}

func TestValueWithExt(t *testing.T) {
	v := Value{Type: TypeExt, Ext: Ext{Type: 1, Data: []byte{0xff}}}
	if v.Type != TypeExt {
		t.Error("Ext type not set")
	}
	if v.Ext.Type != 1 {
		t.Error("Ext type value wrong")
	}
}

func TestValueAccessorsFromInt(t *testing.T) {
	v := Value{Type: TypeInt, Int: 42}

	if v.AsInt() != 42 {
		t.Error("AsInt from Int failed")
	}
	if v.AsUint() != 42 {
		t.Error("AsUint from Int failed")
	}
	if v.AsFloat64() != 42.0 {
		t.Error("AsFloat64 from Int failed")
	}
}

func TestValueAccessorsFromUint(t *testing.T) {
	v := Value{Type: TypeUint, Uint: 42}

	if v.AsInt() != 42 {
		t.Error("AsInt from Uint failed")
	}
	if v.AsUint() != 42 {
		t.Error("AsUint from Uint failed")
	}
	if v.AsFloat64() != 42.0 {
		t.Error("AsFloat64 from Uint failed")
	}
}

func TestValueAccessorsFromFloats(t *testing.T) {
	t.Run("from Float32", func(t *testing.T) {
		v := Value{Type: TypeFloat32, Float32: 3.14}
		f := v.AsFloat64()
		if f < 3.1 || f > 3.2 {
			t.Error("AsFloat64 from Float32 failed")
		}
	})

	t.Run("from Float64", func(t *testing.T) {
		v := Value{Type: TypeFloat64, Float64: 2.718}
		if v.AsFloat64() != 2.718 {
			t.Error("AsFloat64 from Float64 failed")
		}
	})
}

func TestValueLen(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v := Value{Type: TypeString, Bytes: []byte("hello")}
		if v.Len() != 5 {
			t.Error("Len of string failed")
		}
	})

	t.Run("binary", func(t *testing.T) {
		v := Value{Type: TypeBinary, Bytes: []byte{1, 2, 3}}
		if v.Len() != 3 {
			t.Error("Len of binary failed")
		}
	})

	t.Run("array", func(t *testing.T) {
		v := Value{Type: TypeArray, Array: []Value{{}, {}}}
		if v.Len() != 2 {
			t.Error("Len of array failed")
		}
	})

	t.Run("map", func(t *testing.T) {
		v := Value{Type: TypeMap, Map: []KV{{}, {}}}
		if v.Len() != 2 {
			t.Error("Len of map failed")
		}
	})

	t.Run("other", func(t *testing.T) {
		v := Value{Type: TypeInt, Int: 42}
		if v.Len() != 0 {
			t.Error("Len of int should be 0")
		}
	})
}

func TestIsEmptyValue(t *testing.T) {
	type Data struct {
		Slice []int          `msgpack:"slice,omitempty"`
		Map   map[string]int `msgpack:"map,omitempty"`
		Ptr   *int           `msgpack:"ptr,omitempty"`
		Array [2]int         `msgpack:"array,omitempty"`
		Iface any            `msgpack:"iface,omitempty"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{} // all empty
	b, err := enc.Encode(&d)
	if err != nil {
		t.Error("empty struct encode failed")
	}
	// Should produce empty map
	if len(b) == 0 {
		t.Error("empty result")
	}
}

func TestIsEmptyValueAllKinds(t *testing.T) {
	type TestStruct struct {
		B   bool              `msgpack:"b,omitempty"`
		I   int               `msgpack:"i,omitempty"`
		I8  int8              `msgpack:"i8,omitempty"`
		I16 int16             `msgpack:"i16,omitempty"`
		I32 int32             `msgpack:"i32,omitempty"`
		I64 int64             `msgpack:"i64,omitempty"`
		U   uint              `msgpack:"u,omitempty"`
		U8  uint8             `msgpack:"u8,omitempty"`
		U16 uint16            `msgpack:"u16,omitempty"`
		U32 uint32            `msgpack:"u32,omitempty"`
		U64 uint64            `msgpack:"u64,omitempty"`
		F32 float32           `msgpack:"f32,omitempty"`
		F64 float64           `msgpack:"f64,omitempty"`
		S   string            `msgpack:"s,omitempty"`
		Sl  []string          `msgpack:"sl,omitempty"`
		M   map[string]string `msgpack:"m,omitempty"`
		P   *int              `msgpack:"p,omitempty"`
		Ifc any               `msgpack:"ifc,omitempty"`
		Arr [2]int            `msgpack:"arr,omitempty"`
	}

	// Empty struct should have all fields omitted
	var empty TestStruct
	enc := NewEncoder(256)
	err := enc.Encode(empty)
	if err != nil {
		t.Fatal(err)
	}

	// Decode and verify it's minimal (just empty map)
	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeMap {
		t.Errorf("expected map, got %v", v.Type)
	}
}

func TestIsEmptyValueNonEmpty(t *testing.T) {
	type TestStruct struct {
		B   bool              `msgpack:"b,omitempty"`
		I   int               `msgpack:"i,omitempty"`
		S   string            `msgpack:"s,omitempty"`
		Sl  []string          `msgpack:"sl,omitempty"`
		M   map[string]string `msgpack:"m,omitempty"`
		P   *int              `msgpack:"p,omitempty"`
		Arr [2]int            `msgpack:"arr,omitempty"`
	}

	x := 42
	filled := TestStruct{
		B:   true,
		I:   1,
		S:   "test",
		Sl:  []string{"a"},
		M:   map[string]string{"k": "v"},
		P:   &x,
		Arr: [2]int{1, 2},
	}

	enc := NewEncoder(256)
	err := enc.Encode(filled)
	if err != nil {
		t.Fatal(err)
	}

	// Should have more data now
	if len(enc.Bytes()) < 10 {
		t.Error("expected more data for filled struct")
	}
}
