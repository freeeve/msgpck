package msgpck

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

// TestAllFormats tests encoding and decoding of all msgpack formats
func TestAllFormats(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		data := []byte{formatNil}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeNil {
			t.Errorf("nil decode failed: %v, %v", v, err)
		}
	})

	t.Run("bool", func(t *testing.T) {
		// false
		d := NewDecoder([]byte{formatFalse})
		v, _ := d.Decode()
		if v.Type != TypeBool || v.Bool != false {
			t.Error("false decode failed")
		}
		// true
		d.Reset([]byte{formatTrue})
		v, _ = d.Decode()
		if v.Type != TypeBool || v.Bool != true {
			t.Error("true decode failed")
		}
	})

	t.Run("positive fixint", func(t *testing.T) {
		for i := 0; i <= 127; i++ {
			d := NewDecoder([]byte{byte(i)})
			v, _ := d.Decode()
			if v.Type != TypeUint || v.Uint != uint64(i) {
				t.Errorf("positive fixint %d failed", i)
			}
		}
	})

	t.Run("negative fixint", func(t *testing.T) {
		for i := -32; i <= -1; i++ {
			d := NewDecoder([]byte{byte(i)})
			v, _ := d.Decode()
			if v.Type != TypeInt || v.Int != int64(i) {
				t.Errorf("negative fixint %d failed: got %v", i, v.Int)
			}
		}
	})

	t.Run("uint8", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint8, 200})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 200 {
			t.Error("uint8 failed")
		}
	})

	t.Run("uint16", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint16, 0x12, 0x34})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x1234 {
			t.Error("uint16 failed")
		}
	})

	t.Run("uint32", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint32, 0x12, 0x34, 0x56, 0x78})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x12345678 {
			t.Error("uint32 failed")
		}
	})

	t.Run("uint64", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint64, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x123456789abcdef0 {
			t.Error("uint64 failed")
		}
	})

	t.Run("int8", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt8, 0x80}) // -128
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -128 {
			t.Error("int8 failed")
		}
	})

	t.Run("int16", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt16, 0x80, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -32768 {
			t.Error("int16 failed")
		}
	})

	t.Run("int32", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt32, 0x80, 0x00, 0x00, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -2147483648 {
			t.Error("int32 failed")
		}
	})

	t.Run("int64", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != math.MinInt64 {
			t.Error("int64 failed")
		}
	})

	t.Run("float32", func(t *testing.T) {
		d := NewDecoder([]byte{formatFloat32, 0x40, 0x48, 0xf5, 0xc3})
		v, _ := d.Decode()
		if v.Type != TypeFloat32 || math.Abs(float64(v.Float32)-3.14) > 0.001 {
			t.Errorf("float32 failed: %v", v.Float32)
		}
	})

	t.Run("float64", func(t *testing.T) {
		d := NewDecoder([]byte{formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18})
		v, _ := d.Decode()
		if v.Type != TypeFloat64 || math.Abs(v.Float64-3.14159265359) > 0.0000001 {
			t.Errorf("float64 failed: %v", v.Float64)
		}
	})

	t.Run("fixstr", func(t *testing.T) {
		data := append([]byte{0xa5}, []byte("hello")...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || string(v.Bytes) != "hello" {
			t.Error("fixstr failed")
		}
	})

	t.Run("str8", func(t *testing.T) {
		str := "this is a longer string for str8 format testing"
		data := append([]byte{formatStr8, byte(len(str))}, []byte(str)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || string(v.Bytes) != str {
			t.Error("str8 failed")
		}
	})

	t.Run("str16", func(t *testing.T) {
		str := string(make([]byte, 300))
		data := append([]byte{formatStr16, 0x01, 0x2c}, []byte(str)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || len(v.Bytes) != 300 {
			t.Error("str16 failed")
		}
	})

	t.Run("bin8", func(t *testing.T) {
		bin := []byte{1, 2, 3, 4, 5}
		data := append([]byte{formatBin8, 5}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || !bytes.Equal(v.Bytes, bin) {
			t.Error("bin8 failed")
		}
	})

	t.Run("bin16", func(t *testing.T) {
		bin := make([]byte, 300)
		data := append([]byte{formatBin16, 0x01, 0x2c}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || len(v.Bytes) != 300 {
			t.Error("bin16 failed")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		bin := make([]byte, 70000)
		data := append([]byte{formatBin32, 0x00, 0x01, 0x11, 0x70}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || len(v.Bytes) != 70000 {
			t.Error("bin32 failed")
		}
	})

	t.Run("fixarray", func(t *testing.T) {
		data := []byte{0x93, 0x01, 0x02, 0x03}
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("fixarray failed")
		}
	})

	t.Run("array16", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeArrayHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeInt(int64(i))
		}
		d := NewDecoder(e.Bytes())
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 20 {
			t.Error("array16 failed")
		}
	})

	t.Run("fixmap", func(t *testing.T) {
		data := []byte{0x82, 0xa1, 'a', 0x01, 0xa1, 'b', 0x02}
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 2 {
			t.Error("fixmap failed")
		}
	})

	t.Run("map16", func(t *testing.T) {
		e := NewEncoder(256)
		e.EncodeMapHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("key")
			e.EncodeInt(int64(i))
		}
		d := NewDecoder(e.Bytes())
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 20 {
			t.Error("map16 failed")
		}
	})

	t.Run("fixext", func(t *testing.T) {
		cases := []struct {
			format byte
			size   int
		}{
			{formatFixExt1, 1},
			{formatFixExt2, 2},
			{formatFixExt4, 4},
			{formatFixExt8, 8},
			{formatFixExt16, 16},
		}
		for _, c := range cases {
			data := append([]byte{c.format, 0x42}, make([]byte, c.size)...)
			d := NewDecoder(data)
			v, err := d.Decode()
			if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != c.size {
				t.Errorf("fixext%d failed", c.size)
			}
		}
	})

	t.Run("ext8", func(t *testing.T) {
		data := append([]byte{formatExt8, 5, 0x42}, make([]byte, 5)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 5 {
			t.Error("ext8 failed")
		}
	})

	t.Run("ext16", func(t *testing.T) {
		data := append([]byte{formatExt16, 0x01, 0x00, 0x42}, make([]byte, 256)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 256 {
			t.Error("ext16 failed")
		}
	})

	t.Run("ext32", func(t *testing.T) {
		data := append([]byte{formatExt32, 0x00, 0x01, 0x00, 0x00, 0x42}, make([]byte, 65536)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 65536 {
			t.Error("ext32 failed")
		}
	})
}

// TestEncoderAllFormats tests encoding produces correct formats
func TestEncoderAllFormats(t *testing.T) {
	t.Run("integers", func(t *testing.T) {
		tests := []struct {
			val    int64
			expect byte
		}{
			{0, 0x00},
			{127, 0x7f},
			{-1, 0xff},
			{-32, 0xe0},
			{-33, formatInt8},
			{128, formatUint8},
			{256, formatUint16},
			{65536, formatUint32},
			{-129, formatInt16},
			{-32769, formatInt32},
		}
		for _, tc := range tests {
			e := NewEncoder(16)
			e.EncodeInt(tc.val)
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeInt(%d): got format 0x%x, want 0x%x", tc.val, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("unsigned integers", func(t *testing.T) {
		tests := []struct {
			val    uint64
			expect byte
		}{
			{0, 0x00},
			{127, 0x7f},
			{128, formatUint8},
			{256, formatUint16},
			{65536, formatUint32},
			{1 << 32, formatUint64},
		}
		for _, tc := range tests {
			e := NewEncoder(16)
			e.EncodeUint(tc.val)
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeUint(%d): got format 0x%x, want 0x%x", tc.val, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("strings", func(t *testing.T) {
		tests := []struct {
			len    int
			expect byte
		}{
			{0, 0xa0},
			{31, 0xbf},
			{32, formatStr8},
			{255, formatStr8},
			{256, formatStr16},
			{65535, formatStr16},
		}
		for _, tc := range tests {
			e := NewEncoder(tc.len + 10)
			e.EncodeString(string(make([]byte, tc.len)))
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeString(len=%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("binary", func(t *testing.T) {
		tests := []struct {
			len    int
			expect byte
		}{
			{0, formatBin8},
			{255, formatBin8},
			{256, formatBin16},
			{65535, formatBin16},
		}
		for _, tc := range tests {
			e := NewEncoder(tc.len + 10)
			e.EncodeBinary(make([]byte, tc.len))
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeBinary(len=%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("arrays", func(t *testing.T) {
		tests := []struct {
			len    int
			expect byte
		}{
			{0, 0x90},
			{15, 0x9f},
			{16, formatArray16},
			{65535, formatArray16},
		}
		for _, tc := range tests {
			e := NewEncoder(16)
			e.EncodeArrayHeader(tc.len)
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeArrayHeader(%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("maps", func(t *testing.T) {
		tests := []struct {
			len    int
			expect byte
		}{
			{0, 0x80},
			{15, 0x8f},
			{16, formatMap16},
			{65535, formatMap16},
		}
		for _, tc := range tests {
			e := NewEncoder(16)
			e.EncodeMapHeader(tc.len)
			if e.Bytes()[0] != tc.expect {
				t.Errorf("EncodeMapHeader(%d): got format 0x%x, want 0x%x", tc.len, e.Bytes()[0], tc.expect)
			}
		}
	})

	t.Run("float32", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat32(3.14)
		if e.Bytes()[0] != formatFloat32 {
			t.Error("EncodeFloat32 format wrong")
		}
	})

	t.Run("float64", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat64(3.14)
		if e.Bytes()[0] != formatFloat64 {
			t.Error("EncodeFloat64 format wrong")
		}
	})
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

// TestTypedDecode tests typed decode functions
func TestTypedDecode(t *testing.T) {
	t.Run("UnmarshalMapStringAny zeroCopy=true", func(t *testing.T) {
		data, _ := Marshal(map[string]any{"a": int64(1)})
		m, err := UnmarshalMapStringAny(data, true)
		if err != nil || m["a"] != int64(1) {
			t.Error("zeroCopy decode failed")
		}
	})

	t.Run("UnmarshalMapStringAny zeroCopy=false", func(t *testing.T) {
		data, _ := Marshal(map[string]any{"a": int64(1)})
		m, err := UnmarshalMapStringAny(data, false)
		if err != nil || m["a"] != int64(1) {
			t.Error("copy decode failed")
		}
	})

	t.Run("UnmarshalMapStringString", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(2)
		e.EncodeString("key1")
		e.EncodeString("value1")
		e.EncodeString("key2")
		e.EncodeString("value2")

		m, err := UnmarshalMapStringString(e.Bytes(), true)
		if err != nil || m["key1"] != "value1" || m["key2"] != "value2" {
			t.Error("UnmarshalMapStringString failed")
		}
	})

	t.Run("DecodeMapFunc", func(t *testing.T) {
		data, _ := Marshal(map[string]any{"x": int64(42)})
		var got int64
		err := DecodeMapFunc(data, func(m map[string]any) error {
			got = m["x"].(int64)
			return nil
		})
		if err != nil || got != 42 {
			t.Error("DecodeMapFunc failed")
		}
	})

	t.Run("DecodeStringMapFunc", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("k")
		e.EncodeString("v")

		var got string
		err := DecodeStringMapFunc(e.Bytes(), func(m map[string]string) error {
			got = m["k"]
			return nil
		})
		if err != nil || got != "v" {
			t.Error("DecodeStringMapFunc failed")
		}
	})
}

// TestStructCodecs tests pre-registered struct codecs
func TestStructCodecs(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	t.Run("StructEncoder", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		p := Person{Name: "Alice", Age: 30}
		b, err := enc.Encode(&p)
		if err != nil {
			t.Fatal(err)
		}
		// Decode and verify
		m, _ := Unmarshal(b)
		if m.(map[string]any)["name"] != "Alice" {
			t.Error("encode failed")
		}
	})

	t.Run("StructDecoder", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		p := Person{Name: "Bob", Age: 25}
		b, _ := enc.Encode(&p)

		dec := GetStructDecoder[Person](false)
		var result Person
		err := dec.Decode(b, &result)
		if err != nil || result.Name != "Bob" || result.Age != 25 {
			t.Error("decode failed")
		}
	})

	t.Run("StructDecoder ZeroCopy", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		p := Person{Name: "Carol", Age: 35}
		b, _ := enc.Encode(&p)

		dec := GetStructDecoder[Person](true)
		var result Person
		err := dec.Decode(b, &result)
		if err != nil || result.Name != "Carol" {
			t.Error("zero copy decode failed")
		}
	})

	t.Run("GetStructDecoder", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		p := Person{Name: "Dave", Age: 40}
		b, _ := enc.Encode(&p)

		dec := GetStructDecoder[Person](false)
		var result Person
		err := dec.Decode(b, &result)
		if err != nil || result.Name != "Dave" {
			t.Error("GetStructDecoder failed")
		}
	})

	t.Run("DecodeStructFunc", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		p := Person{Name: "Eve", Age: 45}
		b, _ := enc.Encode(&p)

		var got string
		err := DecodeStructFunc(b, func(v *Person) error {
			got = v.Name
			return nil
		})
		if err != nil || got != "Eve" {
			t.Error("DecodeStructFunc failed")
		}
	})
}

// TestValueMethods tests Value type methods
func TestValueMethods(t *testing.T) {
	t.Run("AsInt from uint", func(t *testing.T) {
		v := Value{Type: TypeUint, Uint: 42}
		if v.AsInt() != 42 {
			t.Error("AsInt failed")
		}
	})

	t.Run("AsUint from int", func(t *testing.T) {
		v := Value{Type: TypeInt, Int: 42}
		if v.AsUint() != 42 {
			t.Error("AsUint failed")
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
		if math.Abs(v.AsFloat64()-3.14) > 0.001 {
			t.Error("AsFloat64 from float32 failed")
		}
	})

	t.Run("Len", func(t *testing.T) {
		arr := Value{Type: TypeArray, Array: make([]Value, 5)}
		if arr.Len() != 5 {
			t.Error("Len for array failed")
		}

		m := Value{Type: TypeMap, Map: make([]KV, 3)}
		if m.Len() != 3 {
			t.Error("Len for map failed")
		}

		s := Value{Type: TypeString, Bytes: []byte("hello")}
		if s.Len() != 5 {
			t.Error("Len for string failed")
		}
	})

	t.Run("GetString", func(t *testing.T) {
		m := Value{
			Type: TypeMap,
			Map: []KV{
				{Key: []byte("foo"), Value: Value{Type: TypeInt, Int: 42}},
			},
		}
		v := m.GetString("foo")
		if v == nil || v.Int != 42 {
			t.Error("GetString failed")
		}
		if m.GetString("notfound") != nil {
			t.Error("GetString should return nil for missing key")
		}
	})

	t.Run("Type.String", func(t *testing.T) {
		types := []Type{TypeNil, TypeBool, TypeInt, TypeUint, TypeFloat32, TypeFloat64, TypeString, TypeBinary, TypeArray, TypeMap, TypeExt}
		names := []string{"nil", "bool", "int", "uint", "float32", "float64", "string", "binary", "array", "map", "ext"}
		for i, tp := range types {
			if tp.String() != names[i] {
				t.Errorf("Type %d String() = %s, want %s", tp, tp.String(), names[i])
			}
		}
	})
}

// TestEncodeVariousTypes tests encoding of various Go types
func TestEncodeVariousTypes(t *testing.T) {
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

// TestDecodeStructComplex tests complex struct scenarios
func TestDecodeStructComplex(t *testing.T) {
	t.Run("nested struct", func(t *testing.T) {
		type Inner struct {
			X int `msgpack:"x"`
		}
		type Outer struct {
			Inner Inner `msgpack:"inner"`
		}

		o := Outer{Inner: Inner{X: 42}}
		b, _ := Marshal(o)

		var result Outer
		err := UnmarshalStruct(b, &result)
		if err != nil || result.Inner.X != 42 {
			t.Error("nested struct failed")
		}
	})

	t.Run("slice field", func(t *testing.T) {
		type Data struct {
			Items []int `msgpack:"items"`
		}

		d := Data{Items: []int{1, 2, 3}}
		b, _ := Marshal(d)

		var result Data
		err := UnmarshalStruct(b, &result)
		if err != nil || !reflect.DeepEqual(result.Items, d.Items) {
			t.Error("slice field failed")
		}
	})

	t.Run("map field", func(t *testing.T) {
		type Data struct {
			Meta map[string]string `msgpack:"meta"`
		}

		d := Data{Meta: map[string]string{"k": "v"}}
		b, _ := Marshal(d)

		var result Data
		err := UnmarshalStruct(b, &result)
		if err != nil || result.Meta["k"] != "v" {
			t.Error("map field failed")
		}
	})

	t.Run("omitempty", func(t *testing.T) {
		type Data struct {
			Name  string `msgpack:"name"`
			Value int    `msgpack:"value,omitempty"`
		}

		d := Data{Name: "test", Value: 0}
		b, _ := Marshal(d)

		m, _ := Unmarshal(b)
		if _, ok := m.(map[string]any)["value"]; ok {
			t.Error("omitempty field should be omitted")
		}
	})

	// Note: Pointer to primitive fields (*int, *string, etc.) are not yet
	// fully supported by the struct codecs. Use regular fields instead.
}

// TestDecoderHelpers tests Decoder helper methods
func TestDecoderHelpers(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	d := NewDecoder(data)

	if d.Remaining() != 3 {
		t.Error("Remaining failed")
	}

	if d.Position() != 0 {
		t.Error("Position failed")
	}

	d.readByte()

	if d.Position() != 1 {
		t.Error("Position after read failed")
	}
	if d.Remaining() != 2 {
		t.Error("Remaining after read failed")
	}
}

// TestConfigBuilders tests Config builder methods
func TestConfigBuilders(t *testing.T) {
	cfg := DefaultConfig().
		WithMaxStringLen(100).
		WithMaxBinaryLen(200).
		WithMaxArrayLen(50).
		WithMaxMapLen(60).
		WithMaxExtLen(300).
		WithMaxDepth(10)

	if cfg.MaxStringLen != 100 ||
		cfg.MaxBinaryLen != 200 ||
		cfg.MaxArrayLen != 50 ||
		cfg.MaxMapLen != 60 ||
		cfg.MaxExtLen != 300 ||
		cfg.MaxDepth != 10 {
		t.Error("Config builders failed")
	}
}

// TestErrorCases tests various error conditions
func TestErrorCases(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		d := NewDecoder([]byte{0xc1}) // reserved/never used
		_, err := d.Decode()
		if err != ErrInvalidFormat {
			t.Error("expected ErrInvalidFormat")
		}
	})

	t.Run("DecodeStruct not pointer", func(t *testing.T) {
		type S struct{}
		var s S
		d := NewDecoder([]byte{0x80})
		err := d.DecodeStruct(s)
		if err != ErrNotPointer {
			t.Error("expected ErrNotPointer")
		}
	})

	t.Run("DecodeStruct not struct", func(t *testing.T) {
		var i int
		d := NewDecoder([]byte{0x80})
		err := d.DecodeStruct(&i)
		if err != ErrNotStruct {
			t.Error("expected ErrNotStruct")
		}
	})

	t.Run("binary too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxBinaryLen(5)
		data := []byte{formatBin8, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrBinaryTooLong {
			t.Error("expected ErrBinaryTooLong")
		}
	})

	t.Run("ext too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxExtLen(5)
		data := []byte{formatExt8, 10, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrExtTooLong {
			t.Error("expected ErrExtTooLong")
		}
	})

	t.Run("UnmarshalWithConfig", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(1)
		data := []byte{0x82, 0xa1, 'a', 0x01, 0xa1, 'b', 0x02}
		_, err := UnmarshalWithConfig(data, cfg)
		if err != ErrMapTooLong {
			t.Error("expected ErrMapTooLong")
		}
	})

	t.Run("UnmarshalStructWithConfig", func(t *testing.T) {
		type S struct {
			A int `msgpack:"a"`
		}
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 'a', 0x01}
		var s S
		err := UnmarshalStructWithConfig(data, &s, cfg)
		if err != ErrMapTooLong {
			t.Error("expected ErrMapTooLong")
		}
	})
}

// TestEncodeValue tests encoding Value types
func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name string
		v    Value
	}{
		{"nil", Value{Type: TypeNil}},
		{"bool true", Value{Type: TypeBool, Bool: true}},
		{"bool false", Value{Type: TypeBool, Bool: false}},
		{"int", Value{Type: TypeInt, Int: -42}},
		{"uint", Value{Type: TypeUint, Uint: 42}},
		{"float32", Value{Type: TypeFloat32, Float32: 3.14}},
		{"float64", Value{Type: TypeFloat64, Float64: 3.14159}},
		{"string", Value{Type: TypeString, Bytes: []byte("hello")}},
		{"binary", Value{Type: TypeBinary, Bytes: []byte{1, 2, 3}}},
		{"array", Value{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 1}}}},
		{"map", Value{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{Type: TypeInt, Int: 1}}}}},
		{"ext", Value{Type: TypeExt, Ext: Ext{Type: 1, Data: []byte{1, 2, 3, 4}}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(64)
			e.EncodeValue(&tc.v)
			if len(e.Bytes()) == 0 {
				t.Error("EncodeValue produced no output")
			}
		})
	}
}

// TestNewEncoderBuffer tests NewEncoderBuffer
func TestNewEncoderBuffer(t *testing.T) {
	buf := make([]byte, 0, 100)
	e := NewEncoderBuffer(buf)
	e.EncodeString("test")
	if e.Len() == 0 {
		t.Error("NewEncoderBuffer failed")
	}
}
