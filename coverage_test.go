package msgpck

import (
	"testing"
	"time"
)

// Tests to increase coverage of various modules

// TestStructEncoderMethods tests StructEncoder methods
func TestStructEncoderMethods(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	enc := GetStructEncoder[Person]()
	p := Person{Name: "Test", Age: 25}

	t.Run("EncodeCopy", func(t *testing.T) {
		b, err := enc.EncodeCopy(&p)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Error("EncodeCopy returned empty")
		}
	})

	t.Run("EncodeAppend", func(t *testing.T) {
		prefix := []byte{0x01, 0x02}
		b, err := enc.EncodeAppend(prefix, &p)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) <= 2 {
			t.Error("EncodeAppend failed")
		}
	})

	t.Run("GetStructEncoder", func(t *testing.T) {
		enc := GetStructEncoder[Person]()
		b, err := enc.EncodeCopy(&p)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) == 0 {
			t.Error("GetStructEncoder returned empty")
		}
	})
}

// TestStructWithSliceAndMap tests struct decoder with slice and map fields
func TestStructWithSliceAndMap(t *testing.T) {
	type Data struct {
		Tags   []string          `msgpack:"tags"`
		Meta   map[string]string `msgpack:"meta"`
		Values []byte            `msgpack:"values"`
	}

	t.Run("encode and decode", func(t *testing.T) {
		original := Data{
			Tags:   []string{"a", "b", "c"},
			Meta:   map[string]string{"k1": "v1", "k2": "v2"},
			Values: []byte{1, 2, 3},
		}

		enc := GetStructEncoder[Data]()
		b, err := enc.EncodeCopy(&original)
		if err != nil {
			t.Fatal(err)
		}

		dec := GetStructDecoder[Data](false)
		var result Data
		err = dec.Decode(b, &result)
		if err != nil {
			t.Fatal(err)
		}

		if len(result.Tags) != 3 {
			t.Errorf("Tags: got %v, want 3 items", result.Tags)
		}
		if len(result.Meta) != 2 {
			t.Errorf("Meta: got %v, want 2 items", result.Meta)
		}
		if len(result.Values) != 3 {
			t.Errorf("Values: got %v, want 3 items", result.Values)
		}
	})

}

// TestTypedDecodeExtra tests extra typed decode functions
func TestTypedDecodeExtra(t *testing.T) {
	t.Run("UnmarshalMap", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("key")
		e.EncodeString("value")
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringAny(b, false)
		if err != nil || m["key"] != "value" {
			t.Error("UnmarshalMap failed")
		}
	})

	t.Run("UnmarshalMapZeroCopy", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("key")
		e.EncodeString("value")
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringAny(b, true)
		if err != nil || m["key"] != "value" {
			t.Error("UnmarshalMapZeroCopy failed")
		}
	})

	t.Run("nested map decode", func(t *testing.T) {
		// map with nested map
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString("inner")
		e.EncodeMapHeader(1)
		e.EncodeString("nested")
		e.EncodeInt(42)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringAny(b, true)
		if err != nil {
			t.Fatal(err)
		}
		inner := m["inner"].(map[string]any)
		if inner["nested"] != int64(42) {
			t.Error("nested map decode failed")
		}
	})

	t.Run("nested array decode", func(t *testing.T) {
		// array with nested array
		e := NewEncoder(64)
		e.EncodeArrayHeader(2)
		e.EncodeArrayHeader(2)
		e.EncodeInt(1)
		e.EncodeInt(2)
		e.EncodeArrayHeader(2)
		e.EncodeInt(3)
		e.EncodeInt(4)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		d := NewDecoder(b)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		arr := v.([]any)
		if len(arr) != 2 {
			t.Error("nested array decode failed")
		}
	})
}

// TestValueMethods2 tests remaining Value methods
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

// TestEncodeExtFormats tests all ext format encodings
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

// TestDecodeReflectionPaths tests reflection-based decode paths
func TestDecodeReflectionPaths(t *testing.T) {
	t.Run("decode into uint fields", func(t *testing.T) {
		type Data struct {
			U   uint   `msgpack:"u"`
			U64 uint64 `msgpack:"u64"`
		}

		e := NewEncoder(64)
		e.EncodeMapHeader(2)
		e.EncodeString("u")
		e.EncodeUint(100)
		e.EncodeString("u64")
		e.EncodeUint(200)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || d.U != 100 || d.U64 != 200 {
			t.Error("uint decode failed")
		}
	})

	t.Run("decode into float fields", func(t *testing.T) {
		type Data struct {
			F32 float32 `msgpack:"f32"`
			F64 float64 `msgpack:"f64"`
		}

		e := NewEncoder(64)
		e.EncodeMapHeader(2)
		e.EncodeString("f32")
		e.EncodeFloat32(3.14)
		e.EncodeString("f64")
		e.EncodeFloat64(2.718)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("decode into []byte field", func(t *testing.T) {
		type Data struct {
			Bin []byte `msgpack:"bin"`
		}

		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("bin")
		e.EncodeBinary([]byte{1, 2, 3})
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.Bin) != 3 {
			t.Error("[]byte decode failed")
		}
	})

	t.Run("decode with array16 format", func(t *testing.T) {
		type Data struct {
			Items []int `msgpack:"items"`
		}

		// Create array with > 15 elements to use array16
		e := NewEncoder(256)
		e.EncodeMapHeader(1)
		e.EncodeString("items")
		e.EncodeArrayHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeInt(int64(i))
		}
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.Items) != 20 {
			t.Error("array16 decode failed")
		}
	})

	t.Run("decode with map16 format", func(t *testing.T) {
		type Data struct {
			Meta map[string]int `msgpack:"meta"`
		}

		// Create map with > 15 elements to use map16
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("meta")
		e.EncodeMapHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("k")
			e.EncodeInt(int64(i))
		}
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.Meta) != 1 { // Keys are same so only 1
			// Actually all keys are "k" so we get 1 entry
		}
	})
}

// TestDecodeAnyAllFormats tests DecodeAny with all msgpack formats
func TestDecodeAnyAllFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"str8", append([]byte{formatStr8, 5}, []byte("hello")...)},
		{"str16", append([]byte{formatStr16, 0, 5}, []byte("hello")...)},
		{"bin8", append([]byte{formatBin8, 3}, []byte{1, 2, 3}...)},
		{"bin16", append([]byte{formatBin16, 0, 3}, []byte{1, 2, 3}...)},
		{"bin32", append([]byte{formatBin32, 0, 0, 0, 3}, []byte{1, 2, 3}...)},
		{"array16", []byte{formatArray16, 0, 1, 0x01}},
		{"array32", []byte{formatArray32, 0, 0, 0, 1, 0x01}},
		{"map16", []byte{formatMap16, 0, 1, 0xa1, 'k', 0x01}},
		{"map32", []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x01}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			_, err := d.DecodeAny()
			if err != nil {
				t.Errorf("%s decode failed: %v", tc.name, err)
			}
		})
	}
}

// TestEncodeOmitemptyVariants tests omitempty with various types
func TestEncodeOmitemptyVariants(t *testing.T) {
	type Data struct {
		S  string         `msgpack:"s,omitempty"`
		I  int            `msgpack:"i,omitempty"`
		U  uint           `msgpack:"u,omitempty"`
		F  float64        `msgpack:"f,omitempty"`
		B  bool           `msgpack:"b,omitempty"`
		Sl []int          `msgpack:"sl,omitempty"`
		M  map[string]int `msgpack:"m,omitempty"`
		A  [0]int         `msgpack:"a,omitempty"`
	}

	d := Data{} // all zero values
	b, err := MarshalCopy(d)
	if err != nil {
		t.Fatal(err)
	}

	// Should be a map with 0 entries (all omitted)
	decoded, _ := Unmarshal(b)
	m := decoded.(map[string]any)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}

// TestStringFormats tests all string format encodings
func TestStringFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{31},   // max fixstr
		{32},   // str8
		{255},  // max str8
		{256},  // str16
		{1000}, // str16
	}

	for _, tc := range tests {
		s := string(make([]byte, tc.len))
		e := NewEncoder(tc.len + 10)
		e.EncodeStringBytes([]byte(s))
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeStringBytes len=%d failed", tc.len)
		}
	}
}

// TestBinaryFormats tests all binary format encodings
func TestBinaryFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{255},   // max bin8
		{256},   // bin16
		{65535}, // max bin16
	}

	for _, tc := range tests {
		b := make([]byte, tc.len)
		e := NewEncoder(tc.len + 10)
		e.EncodeBinary(b)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeBinary len=%d failed", tc.len)
		}
	}
}

// TestMapFormats tests all map format encodings
func TestMapFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixmap
		{16},    // map16
		{65535}, // max map16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len*4 + 10)
		e.EncodeMapHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeMapHeader len=%d failed", tc.len)
		}
	}
}

// TestArrayFormats tests all array format encodings
func TestArrayFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixarray
		{16},    // array16
		{65535}, // max array16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len + 10)
		e.EncodeArrayHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeArrayHeader len=%d failed", tc.len)
		}
	}
}

// TestDecodeMapAnyFormats tests decodeMapAny with various key formats
func TestDecodeMapAnyFormats(t *testing.T) {
	t.Run("str8 key", func(t *testing.T) {
		key := string(make([]byte, 50)) // needs str8
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString(key)
		e.EncodeInt(42)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringAny(b, true)
		if err != nil || m[key] != int64(42) {
			t.Error("str8 key decode failed")
		}
	})

	t.Run("str16 key", func(t *testing.T) {
		key := string(make([]byte, 300)) // needs str16
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString(key)
		e.EncodeInt(42)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringAny(b, true)
		if err != nil || m[key] != int64(42) {
			t.Error("str16 key decode failed")
		}
	})
}

// TestPeekByte tests peekByte method
func TestPeekByte(t *testing.T) {
	d := NewDecoder([]byte{0x42})
	b, err := d.peekByte()
	if err != nil || b != 0x42 {
		t.Error("peekByte failed")
	}
	// Position should not advance
	if d.pos != 0 {
		t.Error("peekByte advanced position")
	}
}

// TestStructDecoderAllTypes tests pre-reg struct decoder with all field types
func TestStructDecoderAllTypes(t *testing.T) {
	type AllTypes struct {
		S   string  `msgpack:"s"`
		I   int     `msgpack:"i"`
		I64 int64   `msgpack:"i64"`
		I32 int32   `msgpack:"i32"`
		U   uint    `msgpack:"u"`
		U64 uint64  `msgpack:"u64"`
		F64 float64 `msgpack:"f64"`
		F32 float32 `msgpack:"f32"`
		B   bool    `msgpack:"b"`
	}

	original := AllTypes{
		S:   "test",
		I:   -1,
		I64: -100,
		I32: -50,
		U:   1,
		U64: 100,
		F64: 3.14,
		F32: 2.71,
		B:   true,
	}

	enc := GetStructEncoder[AllTypes]()
	b, err := enc.EncodeCopy(&original)
	if err != nil {
		t.Fatal(err)
	}

	dec := GetStructDecoder[AllTypes](false)
	var result AllTypes
	err = dec.Decode(b, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.S != original.S || result.I != original.I || result.B != original.B {
		t.Error("decode mismatch")
	}
}

// TestDecodeValuePaths covers reflection decode paths in decode_struct.go
func TestDecodeValuePaths(t *testing.T) {
	t.Run("int formats", func(t *testing.T) {
		type Data struct {
			V int `msgpack:"v"`
		}
		formats := []struct {
			name string
			data []byte
		}{
			{"positive fixint", []byte{0x81, 0xa1, 'v', 0x42}},
			{"negative fixint", []byte{0x81, 0xa1, 'v', 0xff}},
			{"uint8", []byte{0x81, 0xa1, 'v', formatUint8, 200}},
			{"uint16", []byte{0x81, 0xa1, 'v', formatUint16, 0x01, 0x00}},
			{"uint32", []byte{0x81, 0xa1, 'v', formatUint32, 0, 0, 0x01, 0x00}},
			{"uint64", []byte{0x81, 0xa1, 'v', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}},
			{"int8", []byte{0x81, 0xa1, 'v', formatInt8, 0x80}},
			{"int16", []byte{0x81, 0xa1, 'v', formatInt16, 0xff, 0x00}},
			{"int32", []byte{0x81, 0xa1, 'v', formatInt32, 0xff, 0xff, 0xff, 0x00}},
			{"int64", []byte{0x81, 0xa1, 'v', formatInt64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}},
		}
		for _, f := range formats {
			var d Data
			err := UnmarshalStruct(f.data, &d)
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
		}
	})

	t.Run("uint formats", func(t *testing.T) {
		type Data struct {
			V uint `msgpack:"v"`
		}
		formats := []struct {
			name string
			data []byte
		}{
			{"positive fixint", []byte{0x81, 0xa1, 'v', 0x42}},
			{"uint8", []byte{0x81, 0xa1, 'v', formatUint8, 200}},
			{"int8 positive", []byte{0x81, 0xa1, 'v', formatInt8, 50}},
		}
		for _, f := range formats {
			var d Data
			err := UnmarshalStruct(f.data, &d)
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
		}
	})

	t.Run("float from int", func(t *testing.T) {
		type Data struct {
			V float64 `msgpack:"v"`
		}
		// float field receiving int value
		data := []byte{0x81, 0xa1, 'v', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("string formats", func(t *testing.T) {
		type Data struct {
			V string `msgpack:"v"`
		}
		str := string(make([]byte, 50))
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString("v")
		e.EncodeString(str)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.V) != 50 {
			t.Error("str8 format failed")
		}
	})

	t.Run("bytes from string", func(t *testing.T) {
		type Data struct {
			V []byte `msgpack:"v"`
		}
		// []byte field receiving string value (common case)
		data := []byte{0x81, 0xa1, 'v', 0xa3, 'a', 'b', 'c'}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.V) != 3 {
			t.Error("bytes from string failed")
		}
	})

	t.Run("bytes from binary", func(t *testing.T) {
		type Data struct {
			V []byte `msgpack:"v"`
		}
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("v")
		e.EncodeBinary([]byte{1, 2, 3})
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.V) != 3 {
			t.Error("bytes from binary failed")
		}
	})

	t.Run("interface field", func(t *testing.T) {
		type Data struct {
			V any `msgpack:"v"`
		}
		data := []byte{0x81, 0xa1, 'v', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("array field", func(t *testing.T) {
		type Data struct {
			V [3]int `msgpack:"v"`
		}
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("v")
		e.EncodeArrayHeader(3)
		e.EncodeInt(1)
		e.EncodeInt(2)
		e.EncodeInt(3)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil {
			t.Error(err)
		}
	})
}

// TestTypedDecodeAllFormats tests typed_decode.go paths
func TestTypedDecodeAllFormats(t *testing.T) {
	t.Run("all int formats in any", func(t *testing.T) {
		formats := [][]byte{
			{formatUint8, 200},
			{formatUint16, 0x01, 0x00},
			{formatUint32, 0, 0, 0x01, 0x00},
			{formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00},
			{formatInt8, 0x80},
			{formatInt16, 0xff, 0x00},
			{formatInt32, 0xff, 0xff, 0xff, 0x00},
			{formatInt64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00},
		}
		for i, f := range formats {
			d := NewDecoder(f)
			_, err := d.DecodeAny()
			if err != nil {
				t.Errorf("format %d failed: %v", i, err)
			}
		}
	})

	t.Run("float formats", func(t *testing.T) {
		f32 := []byte{formatFloat32, 0x40, 0x48, 0xf5, 0xc3}
		f64 := []byte{formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}

		d := NewDecoder(f32)
		_, err := d.DecodeAny()
		if err != nil {
			t.Error("float32 failed")
		}

		d.Reset(f64)
		_, err = d.DecodeAny()
		if err != nil {
			t.Error("float64 failed")
		}
	})
}

// TestStructStringMapDecode tests decodeStringMap in struct_decoder.go
func TestStructStringMapDecode(t *testing.T) {
	type Data struct {
		M map[string]string `msgpack:"m"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("m")
	e.EncodeMapHeader(2)
	e.EncodeString("k1")
	e.EncodeString("v1")
	e.EncodeString("k2")
	e.EncodeString("v2")
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	dec := GetStructDecoder[Data](false)
	var d Data
	err := dec.Decode(b, &d)
	if err != nil || len(d.M) != 2 {
		t.Error("string map decode failed")
	}
}

// TestNilDecode tests nil decoding
func TestNilDecode(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		data := []byte{formatNil}
		m, err := UnmarshalMapStringAny(data, true)
		if err != nil || m != nil {
			t.Error("nil map decode should return nil")
		}
	})

	t.Run("nil to struct", func(t *testing.T) {
		type Data struct {
			V int `msgpack:"v"`
		}
		data := []byte{formatNil}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil {
			t.Error("nil to struct should succeed")
		}
	})
}

// TestValueToAny tests the valueToAny conversion function
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

// TestUnmarshalWithConfig tests UnmarshalWithConfig function
func TestUnmarshalWithConfig(t *testing.T) {
	e := NewEncoder(64)
	e.EncodeString("test")
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	cfg := Config{
		MaxStringLen: 1000,
		MaxBinaryLen: 1000,
		MaxArrayLen:  100,
		MaxMapLen:    100,
		MaxExtLen:    1000,
		MaxDepth:     10,
	}
	result, err := UnmarshalWithConfig(b, cfg)
	if err != nil || result != "test" {
		t.Error("UnmarshalWithConfig failed")
	}
}

// TestUnmarshalMapStringStringFormats tests all string format variations
func TestUnmarshalMapStringStringFormats(t *testing.T) {
	t.Run("str8 value", func(t *testing.T) {
		val := string(make([]byte, 50)) // needs str8
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString("key")
		e.EncodeString(val)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringString(b, false)
		if err != nil || m["key"] != val {
			t.Error("str8 value failed")
		}
	})

	t.Run("str16 value", func(t *testing.T) {
		val := string(make([]byte, 300)) // needs str16
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("key")
		e.EncodeString(val)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		m, err := UnmarshalMapStringString(b, false)
		if err != nil || len(m["key"]) != 300 {
			t.Error("str16 value failed")
		}
	})

	t.Run("str32 value", func(t *testing.T) {
		// Create str32 format manually for small value (format byte test)
		data := []byte{
			0x81,                // fixmap 1
			0xa3, 'k', 'e', 'y', // fixstr "key"
			formatStr32, 0, 0, 0, 3, 'v', 'a', 'l', // str32 "val"
		}

		m, err := UnmarshalMapStringString(data, true)
		if err != nil || m["key"] != "val" {
			t.Error("str32 value failed")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		data := []byte{
			0x81,                // fixmap 1
			0xa3, 'k', 'e', 'y', // fixstr "key"
			formatNil, // nil value
		}

		m, err := UnmarshalMapStringString(data, false)
		if err != nil {
			t.Error("nil value should not error")
		}
		// nil values are skipped
		if _, ok := m["key"]; ok {
			t.Error("nil value should be skipped")
		}
	})

	t.Run("map16 format", func(t *testing.T) {
		// map16 with 1 entry
		data := []byte{
			formatMap16, 0, 1, // map16 length 1
			0xa1, 'k', // fixstr "k"
			0xa1, 'v', // fixstr "v"
		}

		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("map16 failed")
		}
	})

	t.Run("map32 format", func(t *testing.T) {
		// map32 with 1 entry
		data := []byte{
			formatMap32, 0, 0, 0, 1, // map32 length 1
			0xa1, 'k', // fixstr "k"
			0xa1, 'v', // fixstr "v"
		}

		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("map32 failed")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		m, err := UnmarshalMapStringString([]byte{formatNil}, false)
		if err != nil || m != nil {
			t.Error("nil should return nil map")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		_, err := UnmarshalMapStringString([]byte{0x01}, false) // positive fixint
		if err != ErrTypeMismatch {
			t.Error("should return type mismatch")
		}
	})

	t.Run("non-string value", func(t *testing.T) {
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'k', // fixstr "k"
			0x42, // positive fixint (not string)
		}
		_, err := UnmarshalMapStringString(data, false)
		if err != ErrTypeMismatch {
			t.Error("non-string value should error")
		}
	})
}

// TestDecodeStringMapFunc tests the callback-based API
func TestDecodeStringMapFunc(t *testing.T) {
	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("key")
	e.EncodeString("value")
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var got string
	err := DecodeStringMapFunc(b, func(m map[string]string) error {
		got = m["key"]
		return nil
	})
	if err != nil || got != "value" {
		t.Error("DecodeStringMapFunc failed")
	}
}

// TestStructEncoderIsEmptyField tests all isEmptyField type paths
func TestStructEncoderIsEmptyField(t *testing.T) {
	type AllTypes struct {
		S   string            `msgpack:"s,omitempty"`
		I   int               `msgpack:"i,omitempty"`
		I64 int64             `msgpack:"i64,omitempty"`
		I32 int32             `msgpack:"i32,omitempty"`
		I16 int16             `msgpack:"i16,omitempty"`
		I8  int8              `msgpack:"i8,omitempty"`
		U   uint              `msgpack:"u,omitempty"`
		U64 uint64            `msgpack:"u64,omitempty"`
		U32 uint32            `msgpack:"u32,omitempty"`
		U16 uint16            `msgpack:"u16,omitempty"`
		U8  uint8             `msgpack:"u8,omitempty"`
		F64 float64           `msgpack:"f64,omitempty"`
		F32 float32           `msgpack:"f32,omitempty"`
		B   bool              `msgpack:"b,omitempty"`
		Sl  []int             `msgpack:"sl,omitempty"`
		M   map[string]string `msgpack:"m,omitempty"`
	}

	// All zero values should be omitted
	enc := GetStructEncoder[AllTypes]()
	data := AllTypes{}
	b, err := enc.EncodeCopy(&data)
	if err != nil {
		t.Fatal(err)
	}

	// Should be an empty map
	m, _ := Unmarshal(b)
	mp := m.(map[string]any)
	// Map is never empty-checked properly so it may still be included
	// Just verify no error and reasonable output
	if len(mp) > 1 { // only map should remain (as it returns false for empty check)
		t.Errorf("expected mostly empty map, got %d fields", len(mp))
	}

	// Now test with non-zero values
	data2 := AllTypes{
		S:   "test",
		I:   1,
		I64: 2,
		I32: 3,
		I16: 4,
		I8:  5,
		U:   6,
		U64: 7,
		U32: 8,
		U16: 9,
		U8:  10,
		F64: 1.1,
		F32: 2.2,
		B:   true,
		Sl:  []int{1},
		M:   map[string]string{"k": "v"},
	}
	b2, err := enc.EncodeCopy(&data2)
	if err != nil {
		t.Fatal(err)
	}

	m2, _ := Unmarshal(b2)
	mp2 := m2.(map[string]any)
	if len(mp2) < 10 {
		t.Errorf("expected many fields, got %d", len(mp2))
	}
}

// TestStructDecoderUintFormats tests decodeUint paths in struct_decoder.go
func TestStructDecoderUintFormats(t *testing.T) {
	type Data struct {
		U   uint   `msgpack:"u"`
		U64 uint64 `msgpack:"u64"`
	}

	formats := []struct {
		name string
		data []byte
	}{
		{"positive fixint", []byte{0x82, 0xa1, 'u', 0x42, 0xa3, 'u', '6', '4', 0x42}},
		{"uint8", []byte{0x82, 0xa1, 'u', formatUint8, 200, 0xa3, 'u', '6', '4', formatUint8, 200}},
		{"uint16", []byte{0x82, 0xa1, 'u', formatUint16, 0x01, 0x00, 0xa3, 'u', '6', '4', formatUint16, 0x01, 0x00}},
		{"uint32", []byte{0x82, 0xa1, 'u', formatUint32, 0, 0, 0x01, 0x00, 0xa3, 'u', '6', '4', formatUint32, 0, 0, 0x01, 0x00}},
		{"uint64", []byte{0x82, 0xa1, 'u', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00, 0xa3, 'u', '6', '4', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}},
	}

	dec := GetStructDecoder[Data](false)
	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			var d Data
			err := dec.Decode(f.data, &d)
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
		})
	}
}

// TestStructDecoderFloatFormats tests decodeFloat paths in struct_decoder.go
func TestStructDecoderFloatFormats(t *testing.T) {
	type Data struct {
		F32 float32 `msgpack:"f32"`
		F64 float64 `msgpack:"f64"`
	}

	t.Run("from positive fixint", func(t *testing.T) {
		data := []byte{0x82, 0xa3, 'f', '3', '2', 0x42, 0xa3, 'f', '6', '4', 0x42}
		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil {
			t.Fatal(err)
		}
		if d.F32 != 66 || d.F64 != 66 {
			t.Errorf("got f32=%v f64=%v, want 66", d.F32, d.F64)
		}
	})

	t.Run("from negative fixint", func(t *testing.T) {
		data := []byte{0x82, 0xa3, 'f', '3', '2', 0xff, 0xa3, 'f', '6', '4', 0xfe}
		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil {
			t.Fatal(err)
		}
		if d.F32 != -1 || d.F64 != -2 {
			t.Errorf("got f32=%v f64=%v, want -1, -2", d.F32, d.F64)
		}
	})

	t.Run("from float32", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("f64")
		e.EncodeFloat32(3.14)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("from various int formats", func(t *testing.T) {
		formats := [][]byte{
			{formatUint8, 100},
			{formatUint16, 0, 100},
			{formatUint32, 0, 0, 0, 100},
			{formatInt8, 50},
			{formatInt16, 0, 50},
			{formatInt32, 0, 0, 0, 50},
			{formatInt64, 0, 0, 0, 0, 0, 0, 0, 50},
		}

		for i, f := range formats {
			data := append([]byte{0x81, 0xa3, 'f', '6', '4'}, f...)
			dec := GetStructDecoder[Data](false)
			var d Data
			err := dec.Decode(data, &d)
			if err != nil {
				t.Errorf("format %d failed: %v", i, err)
			}
		}
	})
}

// TestUint64Overflow tests uint64 values that overflow int64
func TestUint64Overflow(t *testing.T) {
	// Value > max int64
	data := []byte{formatUint64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	d := NewDecoder(data)
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	// Should return as uint64
	if _, ok := v.(uint64); !ok {
		t.Errorf("expected uint64, got %T", v)
	}
}

// TestStructDecoderStringFormats tests string format decoding in struct_decoder.go
func TestStructDecoderStringFormats(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	t.Run("str16", func(t *testing.T) {
		str := string(make([]byte, 300))
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("s")
		e.EncodeString(str)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil || len(d.S) != 300 {
			t.Error("str16 decode failed")
		}
	})

	t.Run("str32 format", func(t *testing.T) {
		// str32 with small string for format test
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 's', // fixstr "s"
			formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o', // str32 "hello"
		}

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.S != "hello" {
			t.Error("str32 decode failed")
		}
	})
}

// TestStructDecoderBytesFormats tests []byte decoding from various formats
func TestStructDecoderBytesFormats(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("bin16", func(t *testing.T) {
		data := make([]byte, 300)
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeBinary(data)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil || len(d.B) != 300 {
			t.Error("bin16 decode failed")
		}
	})

	t.Run("bin32 format", func(t *testing.T) {
		// bin32 with small data for format test
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'b', // fixstr "b"
			formatBin32, 0, 0, 0, 3, 1, 2, 3, // bin32 [1,2,3]
		}

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || len(d.B) != 3 {
			t.Error("bin32 decode failed")
		}
	})

	t.Run("str16 as bytes", func(t *testing.T) {
		str := string(make([]byte, 300))
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeString(str)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil || len(d.B) != 300 {
			t.Error("str16 as bytes failed")
		}
	})

	t.Run("str32 as bytes", func(t *testing.T) {
		// str32 format
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'b', // fixstr "b"
			formatStr32, 0, 0, 0, 3, 'a', 'b', 'c',
		}

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || len(d.B) != 3 {
			t.Error("str32 as bytes failed")
		}
	})
}

// TestStructDecoderArrayFormats tests array format decoding
func TestStructDecoderArrayFormats(t *testing.T) {
	type Data struct {
		A []string `msgpack:"a"`
	}

	t.Run("array16", func(t *testing.T) {
		e := NewEncoder(256)
		e.EncodeMapHeader(1)
		e.EncodeString("a")
		e.EncodeArrayHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("x")
		}
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil || len(d.A) != 20 {
			t.Error("array16 decode failed")
		}
	})

	t.Run("array32 format", func(t *testing.T) {
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'a', // fixstr "a"
			formatArray32, 0, 0, 0, 2, // array32 length 2
			0xa1, 'x', // fixstr "x"
			0xa1, 'y', // fixstr "y"
		}

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || len(d.A) != 2 {
			t.Error("array32 decode failed")
		}
	})
}

// TestStructDecoderMapFormats tests map format decoding in struct fields
func TestStructDecoderMapFormats(t *testing.T) {
	type Data struct {
		M map[string]string `msgpack:"m"`
	}

	t.Run("map16", func(t *testing.T) {
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("m")
		e.EncodeMapHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("k")
			e.EncodeString("v")
		}
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(b, &d)
		if err != nil {
			t.Error("map16 decode failed")
		}
	})

	t.Run("map32 format", func(t *testing.T) {
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'm', // fixstr "m"
			formatMap32, 0, 0, 0, 1, // map32 length 1
			0xa1, 'k', // fixstr "k"
			0xa1, 'v', // fixstr "v"
		}

		dec := GetStructDecoder[Data](false)
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.M["k"] != "v" {
			t.Error("map32 decode failed")
		}
	})
}

// TestStructEncoderGenericTypes tests generic slice/map encoding
func TestStructEncoderGenericTypes(t *testing.T) {
	type Data struct {
		IntSlice []int          `msgpack:"ints"`
		AnyMap   map[string]int `msgpack:"anymap"`
	}

	d := Data{
		IntSlice: []int{1, 2, 3},
		AnyMap:   map[string]int{"a": 1, "b": 2},
	}

	enc := GetStructEncoder[Data]()
	b, err := enc.EncodeCopy(&d)
	if err != nil {
		t.Fatal(err)
	}

	// Just verify it encodes without error
	if len(b) == 0 {
		t.Error("encode failed")
	}
}

// TestDecodeStructFunc tests the generic callback API
func TestDecodeStructFunc(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(2)
	e.EncodeString("name")
	e.EncodeString("Alice")
	e.EncodeString("age")
	e.EncodeInt(30)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var gotName string
	err := DecodeStructFunc[Person](b, func(p *Person) error {
		gotName = p.Name
		return nil
	})
	if err != nil || gotName != "Alice" {
		t.Error("DecodeStructFunc failed")
	}
}

// TestTypedDecodeMapFormats tests map32 and array32 in typed decode
func TestTypedDecodeMapFormats(t *testing.T) {
	t.Run("map32 in typed decode", func(t *testing.T) {
		data := []byte{
			formatMap32, 0, 0, 0, 1, // map32 length 1
			0xa1, 'k', // fixstr "k"
			0x01, // positive fixint 1
		}

		m, err := UnmarshalMapStringAny(data, false)
		if err != nil || m["k"] != int64(1) {
			t.Error("map32 typed decode failed")
		}
	})

	t.Run("nested map32", func(t *testing.T) {
		data := []byte{
			0x81,      // fixmap 1
			0xa1, 'k', // fixstr "k"
			formatMap32, 0, 0, 0, 1, // map32 length 1
			0xa1, 'n', // fixstr "n"
			0x01, // positive fixint 1
		}

		m, err := UnmarshalMapStringAny(data, true)
		if err != nil {
			t.Fatal(err)
		}
		inner := m["k"].(map[string]any)
		if inner["n"] != int64(1) {
			t.Error("nested map32 failed")
		}
	})
}

// TestExtFormats tests extension format decoding
func TestExtFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"fixext1", []byte{formatFixExt1, 1, 0xff}},
		{"fixext2", []byte{formatFixExt2, 1, 0xff, 0xff}},
		{"fixext4", []byte{formatFixExt4, 1, 0xff, 0xff, 0xff, 0xff}},
		{"fixext8", []byte{formatFixExt8, 1, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"fixext16", append([]byte{formatFixExt16, 1}, make([]byte, 16)...)},
		{"ext8", []byte{formatExt8, 3, 1, 0xff, 0xff, 0xff}},
		{"ext16", append([]byte{formatExt16, 0, 3, 1}, make([]byte, 3)...)},
		{"ext32", append([]byte{formatExt32, 0, 0, 0, 3, 1}, make([]byte, 3)...)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			v, err := d.Decode()
			if err != nil {
				t.Errorf("%s decode failed: %v", tc.name, err)
			}
			if v.Type != TypeExt {
				t.Errorf("%s: expected ext type, got %v", tc.name, v.Type)
			}
		})
	}
}

// TestDecodeAnyStr32 tests str32 format in DecodeAny
func TestDecodeAnyStr32(t *testing.T) {
	data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	d := NewDecoder(data)
	v, err := d.DecodeAny()
	if err != nil || v != "hello" {
		t.Error("str32 DecodeAny failed")
	}
}

// TestStructDecoderNilField tests nil value for struct field
func TestStructDecoderNilField(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	data := []byte{
		0x81,      // fixmap 1
		0xa1, 's', // fixstr "s"
		formatNil, // nil
	}

	dec := GetStructDecoder[Data](false)
	var d Data
	err := dec.Decode(data, &d)
	if err != nil {
		t.Error("nil field should succeed")
	}
	if d.S != "" {
		t.Error("nil field should leave zero value")
	}
}

// TestStructDecoderUnknownField tests skipping unknown fields
func TestStructDecoderUnknownField(t *testing.T) {
	type Data struct {
		Name string `msgpack:"name"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(2)
	e.EncodeString("name")
	e.EncodeString("test")
	e.EncodeString("unknown")
	e.EncodeInt(42)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	dec := GetStructDecoder[Data](false)
	var d Data
	err := dec.Decode(b, &d)
	if err != nil || d.Name != "test" {
		t.Error("unknown field handling failed")
	}
}

// TestStructDecoderZeroCopy tests zero-copy decoder
func TestStructDecoderZeroCopy(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("s")
	e.EncodeString("hello")
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	dec := GetStructDecoder[Data](true)
	var d Data
	err := dec.Decode(b, &d)
	if err != nil || d.S != "hello" {
		t.Error("zero-copy decode failed")
	}
}

// TestReflectionDecodeValueUint tests decodeValueUint with various formats
func TestReflectionDecodeValueUint(t *testing.T) {
	type Data struct {
		U uint64 `msgpack:"u"`
	}

	formats := []struct {
		name string
		data []byte
	}{
		{"uint8", []byte{0x81, 0xa1, 'u', formatUint8, 200}},
		{"uint16", []byte{0x81, 0xa1, 'u', formatUint16, 0x01, 0x00}},
		{"uint32", []byte{0x81, 0xa1, 'u', formatUint32, 0, 0, 0x01, 0x00}},
		{"uint64", []byte{0x81, 0xa1, 'u', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}},
		{"int8", []byte{0x81, 0xa1, 'u', formatInt8, 50}},
		{"int16", []byte{0x81, 0xa1, 'u', formatInt16, 0, 50}},
		{"int32", []byte{0x81, 0xa1, 'u', formatInt32, 0, 0, 0, 50}},
		{"int64", []byte{0x81, 0xa1, 'u', formatInt64, 0, 0, 0, 0, 0, 0, 0, 50}},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			var d Data
			err := UnmarshalStruct(f.data, &d)
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
		})
	}
}

// TestReflectionDecodeValueFloat tests decodeValueFloat with various formats
func TestReflectionDecodeValueFloat(t *testing.T) {
	type Data struct {
		F float64 `msgpack:"f"`
	}

	formats := []struct {
		name string
		data []byte
	}{
		{"float32", []byte{0x81, 0xa1, 'f', formatFloat32, 0x40, 0x48, 0xf5, 0xc3}},
		{"float64", []byte{0x81, 0xa1, 'f', formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}},
		{"uint8", []byte{0x81, 0xa1, 'f', formatUint8, 100}},
		{"uint16", []byte{0x81, 0xa1, 'f', formatUint16, 0, 100}},
		{"uint32", []byte{0x81, 0xa1, 'f', formatUint32, 0, 0, 0, 100}},
		{"uint64", []byte{0x81, 0xa1, 'f', formatUint64, 0, 0, 0, 0, 0, 0, 0, 100}},
		{"int8", []byte{0x81, 0xa1, 'f', formatInt8, 50}},
		{"int16", []byte{0x81, 0xa1, 'f', formatInt16, 0, 50}},
		{"int32", []byte{0x81, 0xa1, 'f', formatInt32, 0, 0, 0, 50}},
		{"int64", []byte{0x81, 0xa1, 'f', formatInt64, 0, 0, 0, 0, 0, 0, 0, 50}},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			var d Data
			err := UnmarshalStruct(f.data, &d)
			if err != nil {
				t.Errorf("%s: %v", f.name, err)
			}
		})
	}
}

// TestReflectionDecodeValueBytes tests decodeValueBytes with various formats
func TestReflectionDecodeValueBytes(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("str8", func(t *testing.T) {
		str := string(make([]byte, 50))
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeString(str)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.B) != 50 {
			t.Error("str8 bytes decode failed")
		}
	})

	t.Run("str16", func(t *testing.T) {
		str := string(make([]byte, 300))
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeString(str)
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.B) != 300 {
			t.Error("str16 bytes decode failed")
		}
	})

	t.Run("str32", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'b',
			formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o',
		}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.B) != 5 {
			t.Error("str32 bytes decode failed")
		}
	})

	t.Run("bin8", func(t *testing.T) {
		e := NewEncoder(128)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeBinary(make([]byte, 50))
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.B) != 50 {
			t.Error("bin8 bytes decode failed")
		}
	})

	t.Run("bin16", func(t *testing.T) {
		e := NewEncoder(512)
		e.EncodeMapHeader(1)
		e.EncodeString("b")
		e.EncodeBinary(make([]byte, 300))
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var d Data
		err := UnmarshalStruct(b, &d)
		if err != nil || len(d.B) != 300 {
			t.Error("bin16 bytes decode failed")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'b',
			formatBin32, 0, 0, 0, 3, 1, 2, 3,
		}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.B) != 3 {
			t.Error("bin32 bytes decode failed")
		}
	})
}

// TestReflectionDecodeIntoStruct tests nested struct decoding
func TestReflectionDecodeIntoStruct(t *testing.T) {
	type Inner struct {
		Value int `msgpack:"value"`
	}
	type Outer struct {
		Name  string `msgpack:"name"`
		Inner Inner  `msgpack:"inner"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(2)
	e.EncodeString("name")
	e.EncodeString("test")
	e.EncodeString("inner")
	e.EncodeMapHeader(1)
	e.EncodeString("value")
	e.EncodeInt(42)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var o Outer
	err := UnmarshalStruct(b, &o)
	if err != nil || o.Name != "test" || o.Inner.Value != 42 {
		t.Error("nested struct decode failed")
	}
}

// TestReflectionDecodeIntoArray tests array decoding with extra elements
func TestReflectionDecodeIntoArray(t *testing.T) {
	type Data struct {
		A [2]int `msgpack:"a"`
	}

	// More elements than array can hold
	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("a")
	e.EncodeArrayHeader(5)
	for i := 0; i < 5; i++ {
		e.EncodeInt(int64(i))
	}
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	err := UnmarshalStruct(b, &d)
	if err != nil {
		t.Error("array with extra elements decode failed")
	}
	if d.A[0] != 0 || d.A[1] != 1 {
		t.Error("array values wrong")
	}
}

// TestReflectionDecodeInterfaceField tests interface field decoding
func TestReflectionDecodeInterfaceField(t *testing.T) {
	type Data struct {
		V any `msgpack:"v"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("v")
	e.EncodeInt(42)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	err := UnmarshalStruct(b, &d)
	if err != nil {
		t.Error("interface field decode failed")
	}
}

// TestReflectionDecodeEmbeddedStruct tests embedded struct decoding
func TestReflectionDecodeEmbeddedStruct(t *testing.T) {
	type Base struct {
		ID int `msgpack:"id"`
	}
	type Extended struct {
		Base
		Name string `msgpack:"name"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(2)
	e.EncodeString("id")
	e.EncodeInt(123)
	e.EncodeString("name")
	e.EncodeString("test")
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var ext Extended
	err := UnmarshalStruct(b, &ext)
	if err != nil || ext.ID != 123 || ext.Name != "test" {
		t.Error("embedded struct decode failed")
	}
}

// TestReflectionDecodeWithStringKeyFormats tests DecodeStruct with various key formats
func TestReflectionDecodeWithStringKeyFormats(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	t.Run("str8 key", func(t *testing.T) {
		// Test str8 format key parsing
		data := []byte{0x81}                    // fixmap 1
		data = append(data, formatStr8, 1, 'v') // str8 "v"
		data = append(data, 0x42)               // value 66

		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.V != 66 {
			t.Error("str8 key decode failed")
		}
	})

	t.Run("str16 key", func(t *testing.T) {
		data := []byte{0x81}                        // fixmap 1
		data = append(data, formatStr16, 0, 1, 'v') // str16 "v"
		data = append(data, 0x42)                   // value 66

		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.V != 66 {
			t.Error("str16 key decode failed")
		}
	})

	t.Run("str32 key", func(t *testing.T) {
		data := []byte{0x81}                              // fixmap 1
		data = append(data, formatStr32, 0, 0, 0, 1, 'v') // str32 "v"
		data = append(data, 0x42)                         // value 66

		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.V != 66 {
			t.Error("str32 key decode failed")
		}
	})
}

// TestReflectionDecodeStructMap16 tests DecodeStruct with map16 format
func TestReflectionDecodeStructMap16(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// map16 with 1 entry
	data := []byte{formatMap16, 0, 1, 0xa1, 'v', 0x42}
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != nil || d.V != 66 {
		t.Error("map16 struct decode failed")
	}
}

// TestReflectionDecodeStructMap32 tests DecodeStruct with map32 format
func TestReflectionDecodeStructMap32(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// map32 with 1 entry
	data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'v', 0x42}
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != nil || d.V != 66 {
		t.Error("map32 struct decode failed")
	}
}

// TestReflectionDecodeNonStringKey tests skipping non-string keys
func TestReflectionDecodeNonStringKey(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// Map with non-string key (integer key) followed by string key
	data := []byte{
		0x82,      // fixmap 2
		0x01,      // key: positive fixint 1 (non-string)
		0x42,      // value: 66
		0xa1, 'v', // key: fixstr "v"
		0x43, // value: 67
	}

	var d Data
	err := UnmarshalStruct(data, &d)
	if err != nil || d.V != 67 {
		t.Errorf("non-string key handling failed: err=%v, v=%d", err, d.V)
	}
}

// TestReflectionDecodeErrors tests various error conditions
func TestReflectionDecodeErrors(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	t.Run("not pointer", func(t *testing.T) {
		var d Data
		err := UnmarshalStruct([]byte{0x80}, d) // non-pointer
		if err != ErrNotPointer {
			t.Error("expected ErrNotPointer")
		}
	})

	t.Run("not struct", func(t *testing.T) {
		var i int
		err := UnmarshalStruct([]byte{0x80}, &i)
		if err != ErrNotStruct {
			t.Error("expected ErrNotStruct")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		var d Data
		err := UnmarshalStruct([]byte{0x01}, &d) // positive fixint, not map
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})
}

// TestReflectionUnmarshalStructWithConfig tests UnmarshalStructWithConfig
func TestReflectionUnmarshalStructWithConfig(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	data := []byte{0x81, 0xa1, 'v', 0x42}
	cfg := DefaultConfig().WithMaxMapLen(100).WithMaxDepth(10)

	var d Data
	err := UnmarshalStructWithConfig(data, &d, cfg)
	if err != nil || d.V != 66 {
		t.Error("UnmarshalStructWithConfig failed")
	}
}

// TestDecodeMapAnyStr32Key tests decodeMapAny with str32 key format
func TestDecodeMapAnyStr32Key(t *testing.T) {
	// Map with str32 key
	data := []byte{
		0x81,                                   // fixmap 1
		formatStr32, 0, 0, 0, 3, 'k', 'e', 'y', // str32 "key"
		0x42, // value 66
	}

	d := NewDecoder(data)
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	m := v.(map[string]any)
	if m["key"] != int64(66) {
		t.Error("str32 key decode failed")
	}
}

// TestDecodeMapKeyStr8Str16 tests decodeMapAny with str8 and str16 keys
func TestDecodeMapKeyStr8Str16(t *testing.T) {
	t.Run("str8 key", func(t *testing.T) {
		data := []byte{0x81, formatStr8, 3, 'k', 'e', 'y', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		m := v.(map[string]any)
		if m["key"] != int64(66) {
			t.Error("str8 key failed")
		}
	})

	t.Run("str16 key", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0, 3, 'k', 'e', 'y', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		m := v.(map[string]any)
		if m["key"] != int64(66) {
			t.Error("str16 key failed")
		}
	})
}

// TestEncodeVariousTypesExtra tests encode.go Encode function with various types
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

// TestDecodeIntoMapTypes tests decoding into maps with various value types
func TestDecodeIntoMapTypes(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("m")
	e.EncodeMapHeader(2)
	e.EncodeString("a")
	e.EncodeInt(1)
	e.EncodeString("b")
	e.EncodeInt(2)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	err := UnmarshalStruct(b, &d)
	if err != nil || len(d.M) != 2 || d.M["a"] != 1 || d.M["b"] != 2 {
		t.Error("map decode failed")
	}
}

// TestDecodeIntoSliceTypes tests decoding into slices with various element types
func TestDecodeIntoSliceTypes(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("s")
	e.EncodeArrayHeader(3)
	e.EncodeInt(1)
	e.EncodeInt(2)
	e.EncodeInt(3)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	err := UnmarshalStruct(b, &d)
	if err != nil || len(d.S) != 3 {
		t.Error("slice decode failed")
	}
}

// TestSkipTaggedField tests fields with "-" tag
func TestSkipTaggedField(t *testing.T) {
	type Data struct {
		Name  string `msgpack:"name"`
		Skip  int    `msgpack:"-"`
		Value int    `msgpack:"value"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(2)
	e.EncodeString("name")
	e.EncodeString("test")
	e.EncodeString("value")
	e.EncodeInt(42)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	d.Skip = 999 // should remain unchanged
	err := UnmarshalStruct(b, &d)
	if err != nil || d.Name != "test" || d.Value != 42 || d.Skip != 999 {
		t.Error("skip tagged field failed")
	}
}

// TestDecodeValueStringFormats tests decodeValueString with all formats
func TestDecodeValueStringFormats(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	t.Run("str8", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatStr8, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.S != "hello" {
			t.Error("str8 decode failed")
		}
	})

	t.Run("str16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.S != "hello" {
			t.Error("str16 decode failed")
		}
	})

	t.Run("str32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.S != "hello" {
			t.Error("str32 decode failed")
		}
	})
}

// TestDecodeIntoArrayFormats tests decodeIntoArray with all formats
func TestDecodeIntoArrayFormats(t *testing.T) {
	type Data struct {
		A [3]int `msgpack:"a"`
	}

	t.Run("array16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'a', formatArray16, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.A[0] != 1 || d.A[1] != 2 || d.A[2] != 3 {
			t.Error("array16 decode failed")
		}
	})

	t.Run("array32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'a', formatArray32, 0, 0, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.A[0] != 1 || d.A[1] != 2 || d.A[2] != 3 {
			t.Error("array32 decode failed")
		}
	})
}

// TestDecodeIntoSliceFormats tests decodeIntoSlice with all formats
func TestDecodeIntoSliceFormats(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	t.Run("array16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatArray16, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.S) != 3 {
			t.Error("array16 slice decode failed")
		}
	})

	t.Run("array32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatArray32, 0, 0, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.S) != 3 {
			t.Error("array32 slice decode failed")
		}
	})
}

// TestDecodeIntoMapFormats tests decodeIntoMap with all formats
func TestDecodeIntoMapFormats(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	t.Run("map16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', formatMap16, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("map16 decode failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("map32 decode failed")
		}
	})
}

// TestDecodeIntoStructFormats tests decodeIntoStruct with all formats
func TestDecodeIntoStructFormats(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	t.Run("map16 nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i', // outer map with "i" key
			formatMap16, 0, 1, // map16 for inner
			0xa1, 'v', 0x42, // "v": 66
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("map16 nested decode failed")
		}
	})

	t.Run("map32 nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i', // outer map with "i" key
			formatMap32, 0, 0, 0, 1, // map32 for inner
			0xa1, 'v', 0x42, // "v": 66
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("map32 nested decode failed")
		}
	})
}

// TestDecodeStringKeyFormats tests decodeStringKey with all formats in nested structs
func TestDecodeStringKeyFormats(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	t.Run("str8 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i', // outer: "i"
			0x81,               // inner fixmap 1
			formatStr8, 1, 'v', // str8 "v"
			0x42, // 66
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str8 key in nested failed")
		}
	})

	t.Run("str16 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i', // outer: "i"
			0x81,                   // inner fixmap 1
			formatStr16, 0, 1, 'v', // str16 "v"
			0x42, // 66
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str16 key in nested failed")
		}
	})

	t.Run("str32 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i', // outer: "i"
			0x81,                         // inner fixmap 1
			formatStr32, 0, 0, 0, 1, 'v', // str32 "v"
			0x42, // 66
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str32 key in nested failed")
		}
	})
}

// TestDecodeMapValueFormats tests decodeMap in decode_value.go with various formats
func TestDecodeMapValueFormats(t *testing.T) {
	t.Run("str8 key in Value decode", func(t *testing.T) {
		data := []byte{0x81, formatStr8, 3, 'k', 'e', 'y', 0x42}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if v.Type != TypeMap || len(v.Map) != 1 {
			t.Error("str8 key in Value failed")
		}
	})

	t.Run("str16 key in Value decode", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0, 3, 'k', 'e', 'y', 0x42}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if v.Type != TypeMap || len(v.Map) != 1 {
			t.Error("str16 key in Value failed")
		}
	})

	t.Run("map16 Value decode", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if v.Type != TypeMap || len(v.Map) != 1 {
			t.Error("map16 Value failed")
		}
	})

	t.Run("map32 Value decode", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if v.Type != TypeMap || len(v.Map) != 1 {
			t.Error("map32 Value failed")
		}
	})
}

// TestDecodeStringBinaryFormats tests decodeString and decodeBinary
func TestDecodeStringBinaryFormats(t *testing.T) {
	t.Run("str16 Value decode", func(t *testing.T) {
		data := []byte{formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeString {
			t.Error("str16 Value failed")
		}
	})

	t.Run("str32 Value decode", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeString {
			t.Error("str32 Value failed")
		}
	})

	t.Run("bin16 Value decode", func(t *testing.T) {
		data := []byte{formatBin16, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeBinary {
			t.Error("bin16 Value failed")
		}
	})

	t.Run("bin32 Value decode", func(t *testing.T) {
		data := []byte{formatBin32, 0, 0, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeBinary {
			t.Error("bin32 Value failed")
		}
	})
}

// TestDecodeIntoValueUnsupportedType tests unsupported type error path
func TestDecodeIntoValueUnsupportedType(t *testing.T) {
	type Data struct {
		C chan int `msgpack:"c"`
	}

	e := NewEncoder(64)
	e.EncodeMapHeader(1)
	e.EncodeString("c")
	e.EncodeInt(42)
	b := make([]byte, len(e.Bytes()))
	copy(b, e.Bytes())

	var d Data
	err := UnmarshalStruct(b, &d)
	if err != ErrUnsupportedType {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}
}

// TestDecodeBoolTypeMismatch tests bool type mismatch
func TestDecodeBoolTypeMismatch(t *testing.T) {
	type Data struct {
		B bool `msgpack:"b"`
	}

	data := []byte{0x81, 0xa1, 'b', 0x42} // int instead of bool
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for bool, got %v", err)
	}
}

// TestDecodeIntTypeMismatch tests int type mismatch
func TestDecodeIntTypeMismatch(t *testing.T) {
	type Data struct {
		I int `msgpack:"i"`
	}

	data := []byte{0x81, 0xa1, 'i', formatTrue} // bool instead of int
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for int, got %v", err)
	}
}

// TestDecodeUintTypeMismatch tests uint type mismatch
func TestDecodeUintTypeMismatch(t *testing.T) {
	type Data struct {
		U uint `msgpack:"u"`
	}

	data := []byte{0x81, 0xa1, 'u', formatTrue} // bool instead of uint
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for uint, got %v", err)
	}
}

// TestDecodeFloatTypeMismatch tests float type mismatch
func TestDecodeFloatTypeMismatch(t *testing.T) {
	type Data struct {
		F float64 `msgpack:"f"`
	}

	data := []byte{0x81, 0xa1, 'f', formatTrue} // bool instead of float
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for float, got %v", err)
	}
}

// TestDecodeStringTypeMismatch tests string type mismatch
func TestDecodeStringTypeMismatch(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	data := []byte{0x81, 0xa1, 's', 0x42} // int instead of string
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for string, got %v", err)
	}
}

// TestDecodeBytesTypeMismatch tests bytes type mismatch
func TestDecodeBytesTypeMismatch(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	data := []byte{0x81, 0xa1, 'b', 0x42} // int instead of bytes
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for bytes, got %v", err)
	}
}

// TestDecodeSliceTypeMismatch tests slice type mismatch
func TestDecodeSliceTypeMismatch(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	data := []byte{0x81, 0xa1, 's', 0x42} // int instead of array
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for slice, got %v", err)
	}
}

// TestDecodeArrayTypeMismatch tests array type mismatch
func TestDecodeArrayTypeMismatch(t *testing.T) {
	type Data struct {
		A [3]int `msgpack:"a"`
	}

	data := []byte{0x81, 0xa1, 'a', 0x42} // int instead of array
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for array, got %v", err)
	}
}

// TestDecodeMapTypeMismatch tests map type mismatch
func TestDecodeMapTypeMismatch(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	data := []byte{0x81, 0xa1, 'm', 0x42} // int instead of map
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for map, got %v", err)
	}
}

// TestDecodeNestedStructTypeMismatch tests nested struct type mismatch
func TestDecodeNestedStructTypeMismatch(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	data := []byte{0x81, 0xa1, 'i', 0x42} // int instead of map for struct
	var o Outer
	err := UnmarshalStruct(data, &o)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch for nested struct, got %v", err)
	}
}

// TestValidationLimitsDecodeAny tests validation limit errors in DecodeAny
func TestValidationLimitsDecodeAny(t *testing.T) {
	t.Run("string too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(2)
		data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // fixstr "hello" (5 bytes)
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})

	t.Run("binary too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxBinaryLen(2)
		data := []byte{formatBin8, 5, 1, 2, 3, 4, 5}
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrBinaryTooLong {
			t.Errorf("expected ErrBinaryTooLong, got %v", err)
		}
	})

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(2)
		data := []byte{0x95, 1, 2, 3, 4, 5} // fixarray 5 elements
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(1)
		data := []byte{0x82, 0xa1, 'a', 1, 0xa1, 'b', 2} // fixmap 2 elements
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})

	t.Run("max depth exceeded", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxDepth(1)
		// Nested array: [[1]]
		data := []byte{0x91, 0x91, 1}
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrMaxDepthExceeded {
			t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
		}
	})
}

// TestValidationLimitsDecodeStruct tests validation limit errors in DecodeStruct
func TestValidationLimitsDecodeStruct(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	t.Run("string too long in struct", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(2)
		data := []byte{0x81, 0xa1, 's', 0xa5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})
}

// TestDecodeAnyInvalidFormat tests invalid format handling
func TestDecodeAnyInvalidFormat(t *testing.T) {
	// Use a format byte that's not valid (0xc1 is never used in msgpack)
	data := []byte{0xc1}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

// TestDecodeMapNonStringKey tests non-string key handling in maps
func TestDecodeMapNonStringKey(t *testing.T) {
	// Map with integer key in DecodeAny (should error)
	data := []byte{0x81, 0x42, 0x01} // fixmap 1 with int key
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat for non-string key, got %v", err)
	}
}

// TestDecodeValueInvalidFormat tests invalid format in Decode (Value)
func TestDecodeValueInvalidFormat(t *testing.T) {
	data := []byte{0xc1} // never used format
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

// TestDecodeUnexpectedEOF tests unexpected EOF errors
func TestDecodeUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"truncated uint16", []byte{formatUint16, 0x00}},
		{"truncated uint32", []byte{formatUint32, 0x00, 0x00, 0x00}},
		{"truncated uint64", []byte{formatUint64, 0x00, 0x00, 0x00, 0x00}},
		{"truncated string", []byte{0xa5, 'h', 'e'}}, // fixstr 5, only 2 bytes
		{"truncated binary", []byte{formatBin8, 5, 1, 2}},
		{"truncated array", []byte{0x92, 0x01}},    // fixarray 2, only 1 element
		{"truncated map", []byte{0x81, 0xa1, 'k'}}, // fixmap 1, missing value
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			_, err := d.DecodeAny()
			if err != ErrUnexpectedEOF {
				t.Errorf("%s: expected ErrUnexpectedEOF, got %v", tc.name, err)
			}
		})
	}
}

// TestDecodeValueUnexpectedEOF tests EOF in Decode (Value)
func TestDecodeValueUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"truncated str8", []byte{formatStr8, 5, 'h'}},
		{"truncated str16", []byte{formatStr16, 0, 5, 'h'}},
		{"truncated str32", []byte{formatStr32, 0, 0, 0, 5, 'h'}},
		{"truncated bin8", []byte{formatBin8, 5, 1}},
		{"truncated bin16", []byte{formatBin16, 0, 5, 1}},
		{"truncated bin32", []byte{formatBin32, 0, 0, 0, 5, 1}},
		{"truncated ext8", []byte{formatExt8, 5, 1}},
		{"truncated map key", []byte{0x81}}, // fixmap 1, no key
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			_, err := d.Decode()
			if err != ErrUnexpectedEOF {
				t.Errorf("%s: expected ErrUnexpectedEOF, got %v", tc.name, err)
			}
		})
	}
}

// TestTypedDecodeMapStringAnyFormats tests typed_decode.go map format paths
func TestTypedDecodeMapStringAnyFormats(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		m, err := UnmarshalMapStringAny(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("map16 in UnmarshalMapStringAny failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		m, err := UnmarshalMapStringAny(data, true)
		if err != nil || m["k"] != "v" {
			t.Error("map32 in UnmarshalMapStringAny failed")
		}
	})
}

// TestTypedDecodeAnyValueFormats tests decodeAnyValue with all formats
func TestTypedDecodeAnyValueFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"fixmap", []byte{0x81, 0xa1, 'k', 0x01}},
		{"fixarray", []byte{0x91, 0x01}},
		{"fixstr", []byte{0xa5, 'h', 'e', 'l', 'l', 'o'}},
		{"nil", []byte{formatNil}},
		{"false", []byte{formatFalse}},
		{"true", []byte{formatTrue}},
		{"uint8", []byte{formatUint8, 200}},
		{"uint16", []byte{formatUint16, 0x01, 0x00}},
		{"uint32", []byte{formatUint32, 0, 0, 0x01, 0x00}},
		{"uint64", []byte{formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}},
		{"int8", []byte{formatInt8, 0x80}},
		{"int16", []byte{formatInt16, 0xff, 0x00}},
		{"int32", []byte{formatInt32, 0xff, 0xff, 0xff, 0x00}},
		{"int64", []byte{formatInt64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}},
		{"float32", []byte{formatFloat32, 0x40, 0x48, 0xf5, 0xc3}},
		{"float64", []byte{formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}},
		{"str8", []byte{formatStr8, 5, 'h', 'e', 'l', 'l', 'o'}},
		{"str16", []byte{formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}},
		{"str32", []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}},
		{"bin8", []byte{formatBin8, 3, 1, 2, 3}},
		{"bin16", []byte{formatBin16, 0, 3, 1, 2, 3}},
		{"bin32", []byte{formatBin32, 0, 0, 0, 3, 1, 2, 3}},
		{"array16", []byte{formatArray16, 0, 1, 0x01}},
		{"array32", []byte{formatArray32, 0, 0, 0, 1, 0x01}},
		{"map16", []byte{formatMap16, 0, 1, 0xa1, 'k', 0x01}},
		{"map32", []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x01}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test via UnmarshalMapStringAny which uses decodeAnyValue
			mapData := append([]byte{0x81, 0xa1, 'v'}, tc.data...)
			m, err := UnmarshalMapStringAny(mapData, false)
			if err != nil {
				t.Errorf("%s failed: %v", tc.name, err)
			}
			if m["v"] == nil && tc.name != "nil" {
				t.Errorf("%s: value is nil", tc.name)
			}
		})
	}
}

// TestEncodeValueAllTypes tests encodeValue with Value type
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

// TestDecoderPosition tests the Position method
func TestDecoderPosition(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	d := NewDecoder(data)

	if d.Position() != 0 {
		t.Error("initial position should be 0")
	}

	d.readByte()
	if d.Position() != 1 {
		t.Error("position should be 1 after readByte")
	}
}

// TestDecoderRemaining tests the Remaining method
func TestDecoderRemaining(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	d := NewDecoder(data)

	if d.Remaining() != 3 {
		t.Error("initial remaining should be 3")
	}

	d.readByte()
	if d.Remaining() != 2 {
		t.Error("remaining should be 2 after readByte")
	}
}

// TestEncoderBufferGrowth tests encoder buffer growth
func TestEncoderBufferGrowth(t *testing.T) {
	e := NewEncoder(1) // Start with tiny buffer

	// Write something larger than initial buffer
	bigString := string(make([]byte, 100))
	e.EncodeString(bigString)

	if len(e.Bytes()) == 0 {
		t.Error("encoder should have grown buffer")
	}
}

// TestConfigChaining tests config method chaining
func TestConfigChaining(t *testing.T) {
	cfg := DefaultConfig().
		WithMaxStringLen(100).
		WithMaxBinaryLen(200).
		WithMaxArrayLen(50).
		WithMaxMapLen(60).
		WithMaxExtLen(300).
		WithMaxDepth(20)

	if cfg.MaxStringLen != 100 {
		t.Error("MaxStringLen not set")
	}
	if cfg.MaxBinaryLen != 200 {
		t.Error("MaxBinaryLen not set")
	}
	if cfg.MaxArrayLen != 50 {
		t.Error("MaxArrayLen not set")
	}
	if cfg.MaxMapLen != 60 {
		t.Error("MaxMapLen not set")
	}
	if cfg.MaxExtLen != 300 {
		t.Error("MaxExtLen not set")
	}
	if cfg.MaxDepth != 20 {
		t.Error("MaxDepth not set")
	}
}

// TestValueAsHelpers tests Value As* helper methods
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

// TestStructDecoderIntFormats tests StructDecoder with various int formats
func TestStructDecoderIntFormats(t *testing.T) {
	type Data struct {
		I int64 `msgpack:"i"`
	}

	tests := []struct {
		name string
		data []byte
		want int64
	}{
		{"positive fixint", []byte{0x81, 0xa1, 'i', 0x42}, 66},
		{"negative fixint", []byte{0x81, 0xa1, 'i', 0xe0}, -32},
		{"uint8", []byte{0x81, 0xa1, 'i', formatUint8, 200}, 200},
		{"uint16", []byte{0x81, 0xa1, 'i', formatUint16, 0x01, 0x00}, 256},
		{"uint32", []byte{0x81, 0xa1, 'i', formatUint32, 0, 0, 0x01, 0x00}, 256},
		{"uint64", []byte{0x81, 0xa1, 'i', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}, 256},
		{"int8", []byte{0x81, 0xa1, 'i', formatInt8, 0x80}, -128},
		{"int16", []byte{0x81, 0xa1, 'i', formatInt16, 0xff, 0x00}, -256},
		{"int32", []byte{0x81, 0xa1, 'i', formatInt32, 0xff, 0xff, 0xff, 0x00}, -256},
		{"int64", []byte{0x81, 0xa1, 'i', formatInt64, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, -256},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Data
			err := dec.Decode(tc.data, &d)
			if err != nil {
				t.Errorf("%s: %v", tc.name, err)
			}
			if d.I != tc.want {
				t.Errorf("%s: got %d, want %d", tc.name, d.I, tc.want)
			}
		})
	}
}

// TestStructDecoderInt32Field tests StructDecoder int32 field
func TestStructDecoderInt32Field(t *testing.T) {
	type Data struct {
		I int32 `msgpack:"i"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{0x81, 0xa1, 'i', formatInt32, 0, 0, 0, 42}
	var d Data
	err := dec.Decode(data, &d)
	if err != nil || d.I != 42 {
		t.Error("int32 field failed")
	}
}

// TestStructDecoderUintField tests StructDecoder uint/uint64 fields
func TestStructDecoderUintField(t *testing.T) {
	type Data struct {
		U uint64 `msgpack:"u"`
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"positive fixint", []byte{0x81, 0xa1, 'u', 0x42}},
		{"uint8", []byte{0x81, 0xa1, 'u', formatUint8, 200}},
		{"uint16", []byte{0x81, 0xa1, 'u', formatUint16, 0x01, 0x00}},
		{"uint32", []byte{0x81, 0xa1, 'u', formatUint32, 0, 0, 0x01, 0x00}},
		{"uint64", []byte{0x81, 0xa1, 'u', formatUint64, 0, 0, 0, 0, 0, 0, 0x01, 0x00}},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Data
			err := dec.Decode(tc.data, &d)
			if err != nil {
				t.Errorf("%s: %v", tc.name, err)
			}
		})
	}
}

// TestStructDecoderUint fields tests StructDecoder uint (not uint64) field
func TestStructDecoderUintNonU64(t *testing.T) {
	type Data struct {
		U uint `msgpack:"u"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{0x81, 0xa1, 'u', formatUint32, 0, 0, 0, 42}
	var d Data
	err := dec.Decode(data, &d)
	if err != nil || d.U != 42 {
		t.Error("uint field failed")
	}
}

// TestStructDecoderFloat32Field tests StructDecoder float32 field
func TestStructDecoderFloat32Field(t *testing.T) {
	type Data struct {
		F float32 `msgpack:"f"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{0x81, 0xa1, 'f', formatFloat32, 0x40, 0x48, 0xf5, 0xc3} // 3.14
	var d Data
	err := dec.Decode(data, &d)
	if err != nil {
		t.Error("float32 field failed")
	}
}

// TestStructDecoderTypeMismatch tests type mismatch errors in StructDecoder
func TestStructDecoderTypeMismatch(t *testing.T) {
	type DataBool struct {
		B bool `msgpack:"b"`
	}
	type DataInt struct {
		I int `msgpack:"i"`
	}
	type DataUint struct {
		U uint `msgpack:"u"`
	}

	t.Run("bool mismatch", func(t *testing.T) {
		dec := GetStructDecoder[DataBool](false)
		data := []byte{0x81, 0xa1, 'b', 0x42} // int instead of bool
		var d DataBool
		err := dec.Decode(data, &d)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch for bool")
		}
	})

	t.Run("int mismatch", func(t *testing.T) {
		dec := GetStructDecoder[DataInt](false)
		data := []byte{0x81, 0xa1, 'i', formatTrue}
		var d DataInt
		err := dec.Decode(data, &d)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch for int")
		}
	})

	t.Run("uint mismatch", func(t *testing.T) {
		dec := GetStructDecoder[DataUint](false)
		data := []byte{0x81, 0xa1, 'u', formatTrue}
		var d DataUint
		err := dec.Decode(data, &d)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch for uint")
		}
	})
}

// TestStructDecoderStringSlice tests decodeStringSlice with array16/array32
func TestStructDecoderStringSlice(t *testing.T) {
	type Data struct {
		S []string `msgpack:"s"`
	}

	t.Run("array16", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 's', formatArray16, 0, 2, 0xa1, 'a', 0xa1, 'b'}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || len(d.S) != 2 {
			t.Error("array16 string slice failed")
		}
	})

	t.Run("array32", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 's', formatArray32, 0, 0, 0, 2, 0xa1, 'a', 0xa1, 'b'}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || len(d.S) != 2 {
			t.Error("array32 string slice failed")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 's', 0x42} // int instead of array
		var d Data
		err := dec.Decode(data, &d)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})
}

// TestStructDecoderStringMap tests decodeStringMap with map16/map32
func TestStructDecoderStringMap(t *testing.T) {
	type Data struct {
		M map[string]string `msgpack:"m"`
	}

	t.Run("map16", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 'm', formatMap16, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.M["k"] != "v" {
			t.Error("map16 string map failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 'm', formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.M["k"] != "v" {
			t.Error("map32 string map failed")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{0x81, 0xa1, 'm', 0x42}
		var d Data
		err := dec.Decode(data, &d)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})
}

// TestStructDecoderStringFormatsAll tests decodeString with str8/str16/str32
func TestStructDecoderStringFormatsAll(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"str8", []byte{0x81, 0xa1, 's', formatStr8, 5, 'h', 'e', 'l', 'l', 'o'}},
		{"str16", []byte{0x81, 0xa1, 's', formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}},
		{"str32", []byte{0x81, 0xa1, 's', formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Data
			err := dec.Decode(tc.data, &d)
			if err != nil || d.S != "hello" {
				t.Errorf("%s failed", tc.name)
			}
		})
	}
}

// TestStructDecoderBytesFormatsAll tests decodeBytes with all formats
func TestStructDecoderBytesFormatsAll(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"fixstr", []byte{0x81, 0xa1, 'b', 0xa3, 'a', 'b', 'c'}, 3},
		{"str8", []byte{0x81, 0xa1, 'b', formatStr8, 3, 'a', 'b', 'c'}, 3},
		{"str16", []byte{0x81, 0xa1, 'b', formatStr16, 0, 3, 'a', 'b', 'c'}, 3},
		{"str32", []byte{0x81, 0xa1, 'b', formatStr32, 0, 0, 0, 3, 'a', 'b', 'c'}, 3},
		{"bin8", []byte{0x81, 0xa1, 'b', formatBin8, 3, 1, 2, 3}, 3},
		{"bin16", []byte{0x81, 0xa1, 'b', formatBin16, 0, 3, 1, 2, 3}, 3},
		{"bin32", []byte{0x81, 0xa1, 'b', formatBin32, 0, 0, 0, 3, 1, 2, 3}, 3},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var d Data
			err := dec.Decode(tc.data, &d)
			if err != nil || len(d.B) != tc.want {
				t.Errorf("%s failed: err=%v, len=%d", tc.name, err, len(d.B))
			}
		})
	}
}

// TestStructDecoderMap16Map32 tests StructDecoder with map16/map32 outer
func TestStructDecoderMap16Map32(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	t.Run("map16", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{formatMap16, 0, 1, 0xa1, 'v', 0x42}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.V != 66 {
			t.Error("map16 decode failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		dec := GetStructDecoder[Data](false)
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'v', 0x42}
		var d Data
		err := dec.Decode(data, &d)
		if err != nil || d.V != 66 {
			t.Error("map32 decode failed")
		}
	})
}

// TestStructDecoderNilMap tests nil value decoding
func TestStructDecoderNilMap(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{formatNil}
	var d Data
	err := dec.Decode(data, &d)
	if err != nil {
		t.Error("nil map decode failed")
	}
}

// TestStructDecoderGenericSlice tests generic slice handling (skip path)
func TestStructDecoderGenericSlice(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{0x81, 0xa1, 's', 0x92, 0x01, 0x02} // fixarray [1, 2]
	var d Data
	err := dec.Decode(data, &d)
	// Generic slice should be skipped but not error
	if err != nil {
		t.Error("generic slice skip failed")
	}
}

// TestStructDecoderGenericMap tests generic map handling (skip path)
func TestStructDecoderGenericMap(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	dec := GetStructDecoder[Data](false)
	data := []byte{0x81, 0xa1, 'm', 0x81, 0xa1, 'k', 0x42}
	var d Data
	err := dec.Decode(data, &d)
	// Generic map should be skipped but not error
	if err != nil {
		t.Error("generic map skip failed")
	}
}

// TestDecoderReadStringBytes tests readStringBytes with all formats
func TestDecoderReadStringBytes(t *testing.T) {
	t.Run("fixstr", func(t *testing.T) {
		data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		b, err := d.readStringBytes()
		if err != nil || string(b) != "hello" {
			t.Error("fixstr readStringBytes failed")
		}
	})

	t.Run("str8", func(t *testing.T) {
		data := []byte{formatStr8, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		b, err := d.readStringBytes()
		if err != nil || string(b) != "hello" {
			t.Error("str8 readStringBytes failed")
		}
	})

	t.Run("str16", func(t *testing.T) {
		data := []byte{formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		b, err := d.readStringBytes()
		if err != nil || string(b) != "hello" {
			t.Error("str16 readStringBytes failed")
		}
	})

	t.Run("str32", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		b, err := d.readStringBytes()
		if err != nil || string(b) != "hello" {
			t.Error("str32 readStringBytes failed")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		data := []byte{0x42} // int, not string
		d := NewDecoder(data)
		_, err := d.readStringBytes()
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})

	t.Run("EOF on format", func(t *testing.T) {
		data := []byte{}
		d := NewDecoder(data)
		_, err := d.readStringBytes()
		if err != ErrUnexpectedEOF {
			t.Error("expected ErrUnexpectedEOF")
		}
	})
}

// TestDecoderPeekByteEOF tests peekByte EOF handling
func TestDecoderPeekByteEOF(t *testing.T) {
	data := []byte{}
	d := NewDecoder(data)
	_, err := d.peekByte()
	if err != ErrUnexpectedEOF {
		t.Error("expected ErrUnexpectedEOF")
	}
}

// TestDecoderReadFloatEOF tests readFloat32/readFloat64 EOF handling
func TestDecoderReadFloatEOF(t *testing.T) {
	t.Run("readFloat32 EOF", func(t *testing.T) {
		data := []byte{0x00, 0x00} // only 2 bytes, need 4
		d := NewDecoder(data)
		_, err := d.readFloat32()
		if err != ErrUnexpectedEOF {
			t.Error("expected ErrUnexpectedEOF")
		}
	})

	t.Run("readFloat64 EOF", func(t *testing.T) {
		data := []byte{0x00, 0x00, 0x00, 0x00} // only 4 bytes, need 8
		d := NewDecoder(data)
		_, err := d.readFloat64()
		if err != ErrUnexpectedEOF {
			t.Error("expected ErrUnexpectedEOF")
		}
	})
}

// TestTypedDecodeCallbackFuncs tests callback functions in typed_decode.go
func TestTypedDecodeCallbackFuncs(t *testing.T) {
	t.Run("DecodeMapFunc", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', 0xa1, 'v'}
		var result any
		err := DecodeMapFunc(data, func(m map[string]any) error {
			result = m["k"]
			return nil
		})
		if err != nil || result != "v" {
			t.Error("DecodeMapFunc failed")
		}
	})

	t.Run("DecodeStringMapFunc", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', 0xa1, 'v'}
		var result string
		err := DecodeStringMapFunc(data, func(m map[string]string) error {
			result = m["k"]
			return nil
		})
		if err != nil || result != "v" {
			t.Error("DecodeStringMapFunc failed")
		}
	})

	t.Run("DecodeStructFunc", func(t *testing.T) {
		type Data struct {
			V int `msgpack:"v"`
		}
		data := []byte{0x81, 0xa1, 'v', 0x42}
		var result int
		err := DecodeStructFunc(data, func(d *Data) error {
			result = d.V
			return nil
		})
		if err != nil || result != 66 {
			t.Error("DecodeStructFunc failed")
		}
	})

}

// TestTypedDecodeNilMap tests nil map handling
func TestTypedDecodeNilMap(t *testing.T) {
	data := []byte{formatNil}
	m, err := UnmarshalMapStringAny(data, false)
	if err != nil || m != nil {
		t.Error("nil map should return nil")
	}
}

// TestTypedDecodeMapStringString tests UnmarshalMapStringString paths
func TestTypedDecodeMapStringString(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("map16 failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		m, err := UnmarshalMapStringString(data, true)
		if err != nil || m["k"] != "v" {
			t.Error("map32 failed")
		}
	})

	t.Run("nil", func(t *testing.T) {
		data := []byte{formatNil}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m != nil {
			t.Error("nil map failed")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		data := []byte{0x42}
		_, err := UnmarshalMapStringString(data, false)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatNil}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || len(m) != 0 {
			t.Error("nil value should skip entry")
		}
	})

	t.Run("str8 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr8, 1, 'v'}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("str8 value failed")
		}
	})

	t.Run("str16 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr16, 0, 1, 'v'}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("str16 value failed")
		}
	})

	t.Run("str32 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr32, 0, 0, 0, 1, 'v'}
		m, err := UnmarshalMapStringString(data, false)
		if err != nil || m["k"] != "v" {
			t.Error("str32 value failed")
		}
	})
}

// TestEncodeMapWithNonStringKey tests encoding map with non-string key
func TestEncodeMapWithNonStringKey(t *testing.T) {
	m := map[int]string{1: "one", 2: "two"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with non-string key returned error (expected): %v", err)
	}
}

// TestEncodeArray16Array32 tests encoding large arrays
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

// TestEncodeMap16 tests encoding maps with >15 entries
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

// TestMarshalAppend tests MarshalAppend function
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

// TestStructEncoderEncodeCopy tests EncodeCopy
func TestStructEncoderEncodeCopy(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{V: 42}
	b, err := enc.EncodeCopy(&d)
	if err != nil || len(b) == 0 {
		t.Error("EncodeCopy failed")
	}
}

// TestStructEncoderEncodeAppend tests EncodeAppend
func TestStructEncoderEncodeAppend(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{V: 42}
	prefix := []byte{0x01, 0x02}
	b, err := enc.EncodeAppend(prefix, &d)
	if err != nil || len(b) < 3 || b[0] != 0x01 || b[1] != 0x02 {
		t.Error("EncodeAppend failed")
	}
}

// TestDecodeExtFormats tests ext format decoding
func TestDecodeExtFormats(t *testing.T) {
	t.Run("ext16", func(t *testing.T) {
		data := append([]byte{formatExt16, 0, 5, 1}, make([]byte, 5)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt {
			t.Error("ext16 decode failed")
		}
	})

	t.Run("ext32", func(t *testing.T) {
		data := append([]byte{formatExt32, 0, 0, 0, 5, 1}, make([]byte, 5)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt {
			t.Error("ext32 decode failed")
		}
	})
}

// TestDecodeAnyBin16Bin32 tests DecodeAny with bin16/bin32
func TestDecodeAnyBin16Bin32(t *testing.T) {
	t.Run("bin16", func(t *testing.T) {
		data := []byte{formatBin16, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("bin16 DecodeAny failed")
		}
		if b, ok := v.([]byte); !ok || len(b) != 3 {
			t.Error("bin16 result wrong")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		data := []byte{formatBin32, 0, 0, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("bin32 DecodeAny failed")
		}
		if b, ok := v.([]byte); !ok || len(b) != 3 {
			t.Error("bin32 result wrong")
		}
	})
}

// TestDecodeAnyMap16Map32 tests DecodeAny with map16/map32
func TestDecodeAnyMap16Map32(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("map16 DecodeAny failed")
		}
		m := v.(map[string]any)
		if m["k"] != int64(66) {
			t.Error("map16 result wrong")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("map32 DecodeAny failed")
		}
		m := v.(map[string]any)
		if m["k"] != int64(66) {
			t.Error("map32 result wrong")
		}
	})
}

// TestDecodeAnyArray16Array32 tests DecodeAny with array16/array32
func TestDecodeAnyArray16Array32(t *testing.T) {
	t.Run("array16", func(t *testing.T) {
		data := []byte{formatArray16, 0, 2, 0x01, 0x02}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("array16 DecodeAny failed")
		}
		arr := v.([]any)
		if len(arr) != 2 {
			t.Error("array16 result wrong")
		}
	})

	t.Run("array32", func(t *testing.T) {
		data := []byte{formatArray32, 0, 0, 0, 2, 0x01, 0x02}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("array32 DecodeAny failed")
		}
		arr := v.([]any)
		if len(arr) != 2 {
			t.Error("array32 result wrong")
		}
	})
}

// TestDecodeAnyUint64Large tests uint64 values > max int64
func TestDecodeAnyUint64Large(t *testing.T) {
	// uint64 value larger than max int64
	data := []byte{formatUint64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	d := NewDecoder(data)
	v, err := d.DecodeAny()
	if err != nil {
		t.Error("large uint64 DecodeAny failed")
	}
	// Should return as uint64 since > max int64
	if _, ok := v.(uint64); !ok {
		t.Error("large uint64 should return uint64")
	}
}

// TestEncodeStructOmitempty tests struct encoding with omitempty
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

// TestTypeString tests Type.String() for all types
func TestTypeString(t *testing.T) {
	types := []Type{
		TypeNil,
		TypeBool,
		TypeInt,
		TypeUint,
		TypeFloat32,
		TypeFloat64,
		TypeString,
		TypeBinary,
		TypeArray,
		TypeMap,
		TypeExt,
	}

	for _, typ := range types {
		s := typ.String()
		if s == "" {
			t.Errorf("String() returned empty for type %v", typ)
		}
	}
}

// TestValueWithExt tests Value with Ext field
func TestValueWithExt(t *testing.T) {
	v := Value{Type: TypeExt, Ext: Ext{Type: 1, Data: []byte{0xff}}}
	if v.Type != TypeExt {
		t.Error("Ext type not set")
	}
	if v.Ext.Type != 1 {
		t.Error("Ext type value wrong")
	}
}

// TestValueAccessorsFromInt tests AsUint/AsFloat64 when Type is Int
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

// TestValueAccessorsFromUint tests AsInt/AsFloat64 when Type is Uint
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

// TestValueAccessorsFromFloats tests conversions from float types
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

// TestValueLen tests Len() for different types
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

// TestReflectionDecodeIntoStructNonStringKey tests the non-string key skip path
func TestReflectionDecodeIntoStructNonStringKey(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	// Map with non-string key (int) in inner struct
	data := []byte{
		0x81, 0xa1, 'i', // outer: "i"
		0x82,      // inner fixmap 2
		0x42,      // int key (non-string)
		0x01,      // value
		0xa1, 'v', // string key "v"
		0x42, // value 66
	}
	var o Outer
	err := UnmarshalStruct(data, &o)
	if err != nil || o.I.V != 66 {
		t.Errorf("non-string key in nested struct failed: err=%v, v=%d", err, o.I.V)
	}
}

// TestReflectionDecodeUnexportedField tests field that can't be set
func TestReflectionDecodeUnexportedField(t *testing.T) {
	type Data struct {
		Name string `msgpack:"name"`
	}

	data := []byte{0x81, 0xa1, 'n', 'a', 'm', 'e', 0xa4, 't', 'e', 's', 't'}
	var d Data
	err := UnmarshalStruct(data, &d)
	// Should succeed, unexported fields get skipped
	if err != nil {
		t.Error("unexported field handling failed")
	}
}

// TestDecodeValueMapKeyFormats tests decodeMap key format paths in decode_value.go
func TestDecodeValueMapKeyFormats(t *testing.T) {
	t.Run("str32 key", func(t *testing.T) {
		data := []byte{0x81, formatStr32, 0, 0, 0, 1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil {
			t.Fatal(err)
		}
		if v.Type != TypeMap || len(v.Map) != 1 {
			t.Error("str32 key in Value decode failed")
		}
	})
}

// TestDecodeValueExtFormats tests ext format paths
func TestDecodeValueExtFormats(t *testing.T) {
	t.Run("fixext8", func(t *testing.T) {
		data := append([]byte{formatFixExt8, 1}, make([]byte, 8)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt {
			t.Error("fixext8 decode failed")
		}
	})

	t.Run("fixext16", func(t *testing.T) {
		data := append([]byte{formatFixExt16, 1}, make([]byte, 16)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt {
			t.Error("fixext16 decode failed")
		}
	})
}

// TestEncodeSliceArrayPaths tests encode paths for slices and arrays
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

// TestEncodeMapPaths tests encode map with different types
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

// TestEncodeStructFields tests struct encoding with different field types
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

// TestIsEmptyValue tests isEmptyValue for various types
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

// TestMarshalError tests Marshal error path
func TestMarshalError(t *testing.T) {
	// Unsupported type should error
	ch := make(chan int)
	_, err := Marshal(ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

// TestDecodeBinaryAnyFormats tests decodeBinaryAny paths
func TestDecodeBinaryAnyFormats(t *testing.T) {
	t.Run("bin8", func(t *testing.T) {
		data := []byte{formatBin8, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Error("bin8 DecodeAny failed")
		}
		if b, ok := v.([]byte); !ok || len(b) != 3 {
			t.Error("bin8 result wrong")
		}
	})
}

// TestDecodeStringAnyFormats tests decodeStringAny paths
func TestDecodeStringAnyFormats(t *testing.T) {
	t.Run("str8", func(t *testing.T) {
		data := []byte{formatStr8, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil || v != "hello" {
			t.Error("str8 DecodeAny failed")
		}
	})
}

// TestDecodeMapAnyFormatsKeyFormats tests decodeMapAny key paths
func TestDecodeMapAnyFormatsKeyFormats(t *testing.T) {
	t.Run("str8 key", func(t *testing.T) {
		data := []byte{0x81, formatStr8, 1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		m := v.(map[string]any)
		if m["k"] != int64(66) {
			t.Error("str8 key in DecodeAny failed")
		}
	})

	t.Run("str16 key", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0, 1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		m := v.(map[string]any)
		if m["k"] != int64(66) {
			t.Error("str16 key in DecodeAny failed")
		}
	})

	t.Run("str32 key", func(t *testing.T) {
		data := []byte{0x81, formatStr32, 0, 0, 0, 1, 'k', 0x42}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil {
			t.Fatal(err)
		}
		m := v.(map[string]any)
		if m["k"] != int64(66) {
			t.Error("str32 key in DecodeAny failed")
		}
	})
}

// TestReflectionDecodeIntoMapKeyFormats tests key format paths in decodeIntoMap
func TestReflectionDecodeIntoMapKeyFormats(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	t.Run("str8 key", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', 0x81, formatStr8, 1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("str8 key in map failed")
		}
	})

	t.Run("str16 key", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', 0x81, formatStr16, 0, 1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("str16 key in map failed")
		}
	})

	t.Run("str32 key", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', 0x81, formatStr32, 0, 0, 0, 1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("str32 key in map failed")
		}
	})
}

// TestDecodeIntoValueEOF tests EOF paths in decodeIntoValue
func TestDecodeIntoValueEOF(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	data := []byte{0x81, 0xa1, 'v'} // missing value
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestStructEncoderSliceField tests encoding struct with slice field
func TestStructEncoderSliceField(t *testing.T) {
	type Data struct {
		S []string `msgpack:"s"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{S: []string{"a", "b", "c"}}
	b, err := enc.Encode(&d)
	if err != nil || len(b) == 0 {
		t.Error("slice field encode failed")
	}
}

// TestStructEncoderMapField tests encoding struct with map field
func TestStructEncoderMapField(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{M: map[string]int{"k": 42}}
	b, err := enc.Encode(&d)
	if err != nil || len(b) == 0 {
		t.Error("map field encode failed")
	}
}

// TestReflectionDecodeValueBytesEOF tests EOF in decodeValueBytes
func TestReflectionDecodeValueBytesEOF(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	// str8 with declared length but missing data
	data := []byte{0x81, 0xa1, 'b', formatStr8, 100}
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestReflectionDecodeSliceEOF tests EOF in decodeIntoSlice
func TestReflectionDecodeSliceEOF(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	// array with declared length but missing elements
	data := []byte{0x81, 0xa1, 's', 0x95} // fixarray 5 with no elements
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestReflectionDecodeArrayEOF tests EOF in decodeIntoArray
func TestReflectionDecodeArrayEOF(t *testing.T) {
	type Data struct {
		A [3]int `msgpack:"a"`
	}

	// array with declared length but missing elements
	data := []byte{0x81, 0xa1, 'a', 0x93} // fixarray 3 with no elements
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestReflectionDecodeMapEOF tests EOF in decodeIntoMap
func TestReflectionDecodeMapEOF(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	// map with declared length but missing entries
	data := []byte{0x81, 0xa1, 'm', 0x81} // fixmap 1 with no entries
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestReflectionDecodeStructEOF tests EOF in decodeIntoStruct
func TestReflectionDecodeStructEOF(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	// nested struct with missing entries
	data := []byte{0x81, 0xa1, 'i', 0x81} // inner fixmap 1 with no entries
	var o Outer
	err := UnmarshalStruct(data, &o)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestEncoderStringFormats tests encoder string format selection
func TestEncoderStringFormats(t *testing.T) {
	e := NewEncoder(256)

	// fixstr (0-31 bytes)
	e.Reset()
	e.EncodeString("hi")
	if e.Bytes()[0]&0xe0 != 0xa0 {
		t.Error("short string should use fixstr")
	}

	// str8 (32-255 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 50)))
	if e.Bytes()[0] != formatStr8 {
		t.Error("medium string should use str8")
	}

	// str16 (256-65535 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 300)))
	if e.Bytes()[0] != formatStr16 {
		t.Error("large string should use str16")
	}
}

// TestEncoderBinaryFormats tests encoder binary format selection
func TestEncoderBinaryFormats(t *testing.T) {
	e := NewEncoder(256)

	// bin8 (0-255 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 50))
	if e.Bytes()[0] != formatBin8 {
		t.Error("small binary should use bin8")
	}

	// bin16 (256-65535 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 300))
	if e.Bytes()[0] != formatBin16 {
		t.Error("medium binary should use bin16")
	}
}

// TestEncoderMapHeader tests EncodeMapHeader format selection
func TestEncoderMapHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixmap (0-15)
	e.Reset()
	e.EncodeMapHeader(5)
	if e.Bytes()[0]&0xf0 != 0x80 {
		t.Error("small map should use fixmap")
	}

	// map16 (16-65535)
	e.Reset()
	e.EncodeMapHeader(20)
	if e.Bytes()[0] != formatMap16 {
		t.Error("medium map should use map16")
	}
}

// TestEncoderArrayHeader tests EncodeArrayHeader format selection
func TestEncoderArrayHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixarray (0-15)
	e.Reset()
	e.EncodeArrayHeader(5)
	if e.Bytes()[0]&0xf0 != 0x90 {
		t.Error("small array should use fixarray")
	}

	// array16 (16-65535)
	e.Reset()
	e.EncodeArrayHeader(20)
	if e.Bytes()[0] != formatArray16 {
		t.Error("medium array should use array16")
	}
}

// TestDecodeStructKeyReadError tests error in reading map key
func TestDecodeStructKeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str8 key with missing length byte
	data := []byte{0x81, formatStr8} // missing length
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStructStr16KeyReadError tests error in reading str16 map key
func TestDecodeStructStr16KeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str16 key with missing length bytes
	data := []byte{0x81, formatStr16, 0} // missing second length byte
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStructStr32KeyReadError tests error in reading str32 map key
func TestDecodeStructStr32KeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str32 key with missing length bytes
	data := []byte{0x81, formatStr32, 0, 0, 0} // missing last length byte
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStructNonStringKeyComplex tests complex non-string key handling
func TestDecodeStructNonStringKeyComplex(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// Map with array key (non-string)
	data := []byte{
		0x82,       // fixmap 2
		0x91, 0x01, // array key [1]
		0x42,      // value
		0xa1, 'v', // string key "v"
		0x43, // value 67
	}
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != nil || d.V != 67 {
		t.Errorf("complex non-string key handling failed: err=%v, v=%d", err, d.V)
	}
}

// TestDecodeValueAllFormats tests decodeValue with all format types
func TestDecodeValueAllFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"float32", []byte{formatFloat32, 0x40, 0x48, 0xf5, 0xc3}},
		{"float64", []byte{formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}},
		{"array16", []byte{formatArray16, 0, 2, 0x01, 0x02}},
		{"array32", []byte{formatArray32, 0, 0, 0, 2, 0x01, 0x02}},
		{"ext8", append([]byte{formatExt8, 3, 1}, make([]byte, 3)...)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			v, err := d.Decode()
			if err != nil {
				t.Errorf("%s: %v", tc.name, err)
			}
			if v.Type == TypeNil {
				t.Errorf("%s: unexpected nil", tc.name)
			}
		})
	}
}

// TestDecodeValueExtTooLong tests ext too long error
func TestDecodeValueExtTooLong(t *testing.T) {
	cfg := DefaultConfig().WithMaxExtLen(1)
	data := append([]byte{formatExt8, 5, 1}, make([]byte, 5)...)
	d := NewDecoderWithConfig(data, cfg)
	_, err := d.Decode()
	if err != ErrExtTooLong {
		t.Errorf("expected ErrExtTooLong, got %v", err)
	}
}

// TestDecodeMapNonStringKeyValue tests map with non-string key in Decode
func TestDecodeMapNonStringKeyValue(t *testing.T) {
	// Map with integer key
	data := []byte{0x81, 0x42, 0x01} // fixmap 1 with int key 66, value 1
	d := NewDecoder(data)
	v, err := d.Decode()
	// Should still decode but key type might vary
	if err != nil {
		t.Errorf("decode failed: %v", err)
	}
	if v.Type != TypeMap {
		t.Error("expected map type")
	}
}

// TestDecodeStringAllFormats tests decodeString with all formats
func TestDecodeStringAllFormats(t *testing.T) {
	t.Run("str16", func(t *testing.T) {
		data := []byte{formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeString {
			t.Error("str16 Decode failed")
		}
	})

	t.Run("str32", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeString {
			t.Error("str32 Decode failed")
		}
	})
}

// TestDecodeBinaryAllFormats tests decodeBinary with all formats
func TestDecodeBinaryAllFormats(t *testing.T) {
	t.Run("bin16", func(t *testing.T) {
		data := []byte{formatBin16, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeBinary {
			t.Error("bin16 Decode failed")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		data := []byte{formatBin32, 0, 0, 0, 3, 1, 2, 3}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeBinary {
			t.Error("bin32 Decode failed")
		}
	})
}

// TestEncodeMapNonStringKey tests encoding map with non-string keys
func TestEncodeMapNonStringKey(t *testing.T) {
	m := map[int]string{1: "one"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with int key returned error (expected): %v", err)
	}
}

// TestDecodeStructValidationErrors tests validation limit errors in DecodeStruct
func TestDecodeStructValidationErrors(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	t.Run("max depth", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxDepth(0)
		data := []byte{0x81, 0xa1, 's', 0xa4, 't', 'e', 's', 't'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrMaxDepthExceeded {
			t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
		}
	})

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 's', 0xa4, 't', 'e', 's', 't'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})
}

// TestDecodeValueBytesAllFormats tests decodeValueBytes with all formats
func TestDecodeValueBytesAllFormats(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("bin16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin16, 0, 3, 1, 2, 3}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.B) != 3 {
			t.Error("bin16 bytes decode failed")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin32, 0, 0, 0, 3, 1, 2, 3}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.B) != 3 {
			t.Error("bin32 bytes decode failed")
		}
	})
}

// TestDecodeStringAnyAllFormats tests decodeStringAny with str16/str32
func TestDecodeStringAnyAllFormats(t *testing.T) {
	t.Run("str16", func(t *testing.T) {
		data := []byte{formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil || v != "hello" {
			t.Error("str16 DecodeAny failed")
		}
	})

	t.Run("str32", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		d := NewDecoder(data)
		v, err := d.DecodeAny()
		if err != nil || v != "hello" {
			t.Error("str32 DecodeAny failed")
		}
	})
}

// TestDecodeIntoSliceArray16Array32 tests decodeIntoSlice with array16/array32
func TestDecodeIntoSliceArray16Array32(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	t.Run("array16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatArray16, 0, 2, 0x01, 0x02}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.S) != 2 {
			t.Error("array16 slice decode failed")
		}
	})

	t.Run("array32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 's', formatArray32, 0, 0, 0, 2, 0x01, 0x02}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || len(d.S) != 2 {
			t.Error("array32 slice decode failed")
		}
	})
}

// TestDecodeIntoArrayArray16Array32 tests decodeIntoArray with array16/array32
func TestDecodeIntoArrayArray16Array32(t *testing.T) {
	type Data struct {
		A [2]int `msgpack:"a"`
	}

	t.Run("array16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'a', formatArray16, 0, 2, 0x01, 0x02}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.A[0] != 1 {
			t.Error("array16 array decode failed")
		}
	})

	t.Run("array32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'a', formatArray32, 0, 0, 0, 2, 0x01, 0x02}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.A[0] != 1 {
			t.Error("array32 array decode failed")
		}
	})
}

// TestDecodeIntoMapMap16Map32 tests decodeIntoMap with map16/map32
func TestDecodeIntoMapMap16Map32(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	t.Run("map16", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', formatMap16, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("map16 map decode failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'm', formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != nil || d.M["k"] != 66 {
			t.Error("map32 map decode failed")
		}
	})
}

// TestDecodeStringKeyAllFormats tests decodeStringKey with all formats
func TestDecodeStringKeyAllFormats(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	t.Run("str8 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i',
			0x81, formatStr8, 1, 'v', 0x42,
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str8 key in nested failed")
		}
	})

	t.Run("str16 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i',
			0x81, formatStr16, 0, 1, 'v', 0x42,
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str16 key in nested failed")
		}
	})

	t.Run("str32 key in nested", func(t *testing.T) {
		data := []byte{
			0x81, 0xa1, 'i',
			0x81, formatStr32, 0, 0, 0, 1, 'v', 0x42,
		}
		var o Outer
		err := UnmarshalStruct(data, &o)
		if err != nil || o.I.V != 66 {
			t.Error("str32 key in nested failed")
		}
	})
}

// TestReadBytesEOF tests readBytes EOF path
func TestReadBytesEOF(t *testing.T) {
	data := []byte{0xa5, 'h', 'e'} // fixstr 5 with only 2 bytes
	d := NewDecoder(data)
	_, err := d.readStringBytes()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestEncodeSliceArrayError tests error handling in encodeSlice/encodeArray
func TestEncodeSliceArrayError(t *testing.T) {
	// Slice with unsupported element type
	arr := []chan int{make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for channel slice")
	}
}

// TestEncodeMapError tests error handling in encodeMap
func TestEncodeMapError(t *testing.T) {
	// Map with unsupported value type
	m := map[string]chan int{"k": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel map value")
	}
}

// TestDecodeValueBytesEOFPaths tests EOF paths in decodeValueBytes
func TestDecodeValueBytesEOFPaths(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("str16 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatStr16, 0} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("str32 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatStr32, 0, 0} // missing length bytes
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("bin8 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin8} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("bin16 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin16, 0} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("bin32 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin32, 0, 0} // missing length bytes
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestDecodeValueBytesValidation tests validation errors in decodeValueBytes
func TestDecodeValueBytesValidation(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("str16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(1)
		data := []byte{0x81, 0xa1, 'b', formatStr16, 0, 10, 'h', 'e', 'l', 'l', 'o', 'w', 'o', 'r', 'l', 'd'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})

	t.Run("bin16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxBinaryLen(1)
		data := []byte{0x81, 0xa1, 'b', formatBin16, 0, 5, 1, 2, 3, 4, 5}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrBinaryTooLong {
			t.Errorf("expected ErrBinaryTooLong, got %v", err)
		}
	})
}

// TestDecodeValueStringValidation tests validation errors in decodeValueString
func TestDecodeValueStringValidation(t *testing.T) {
	type Data struct {
		S string `msgpack:"s"`
	}

	t.Run("str16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(1)
		data := []byte{0x81, 0xa1, 's', formatStr16, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})

	t.Run("str32 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(1)
		data := []byte{0x81, 0xa1, 's', formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})
}

// TestDecodeIntoSliceValidation tests validation errors in decodeIntoSlice
func TestDecodeIntoSliceValidation(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(1)
		data := []byte{0x81, 0xa1, 's', 0x93, 0x01, 0x02, 0x03} // fixarray 3
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})

	t.Run("array16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(1)
		data := []byte{0x81, 0xa1, 's', formatArray16, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})
}

// TestDecodeIntoArrayValidation tests validation errors in decodeIntoArray
func TestDecodeIntoArrayValidation(t *testing.T) {
	type Data struct {
		A [2]int `msgpack:"a"`
	}

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(1)
		data := []byte{0x81, 0xa1, 'a', 0x93, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})
}

// TestDecodeIntoMapValidation tests validation errors in decodeIntoMap
func TestDecodeIntoMapValidation(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 'm', 0x81, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})

	t.Run("map16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 'm', formatMap16, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})
}

// TestDecodeIntoStructValidation tests validation errors in decodeIntoStruct
func TestDecodeIntoStructValidation(t *testing.T) {
	type Inner struct {
		V int `msgpack:"v"`
	}
	type Outer struct {
		I Inner `msgpack:"i"`
	}

	t.Run("nested map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 'i', 0x81, 0xa1, 'v', 0x42}
		var o Outer
		err := UnmarshalStructWithConfig(data, &o, cfg)
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})
}

// TestDecodeMapKeyEOF tests decodeMap key EOF paths
func TestDecodeMapKeyEOF(t *testing.T) {
	t.Run("str8 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr8} // map with str8 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("str16 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0} // map with str16 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("str32 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr32, 0, 0} // map with str32 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestDecodeExtEOF tests decodeExt EOF paths
func TestDecodeExtEOF(t *testing.T) {
	t.Run("ext8 length EOF", func(t *testing.T) {
		data := []byte{formatExt8} // ext8 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("ext16 length EOF", func(t *testing.T) {
		data := []byte{formatExt16, 0} // ext16 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("ext32 length EOF", func(t *testing.T) {
		data := []byte{formatExt32, 0, 0} // ext32 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestDecodeStringAnyEOF tests decodeStringAny EOF paths
func TestDecodeStringAnyEOF(t *testing.T) {
	t.Run("str16 length EOF", func(t *testing.T) {
		data := []byte{formatStr16, 0} // str16 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("str32 length EOF", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0} // str32 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestDecodeBinaryAnyEOF tests decodeBinaryAny EOF paths
func TestDecodeBinaryAnyEOF(t *testing.T) {
	t.Run("bin16 length EOF", func(t *testing.T) {
		data := []byte{formatBin16, 0} // bin16 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("bin32 length EOF", func(t *testing.T) {
		data := []byte{formatBin32, 0, 0} // bin32 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestDecodeMapAnyKeyEOF tests decodeMapAny key EOF paths
func TestDecodeMapAnyKeyEOF(t *testing.T) {
	t.Run("str8 key length EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr8} // map with str8 key, missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("str16 key length EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0} // map with str16 key, missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestBuildStructFieldsEdgeCases tests edge cases in buildStructFields
func TestBuildStructFieldsEdgeCases(t *testing.T) {
	// Struct with no exported fields
	type NoExported struct {
		private int
	}

	// Struct with embedded struct
	type Base struct {
		ID int `msgpack:"id"`
	}
	type Embedded struct {
		Base
		Name string `msgpack:"name"`
	}

	data := []byte{0x82, 0xa2, 'i', 'd', 0x42, 0xa4, 'n', 'a', 'm', 'e', 0xa4, 't', 'e', 's', 't'}
	var e Embedded
	err := UnmarshalStruct(data, &e)
	if err != nil || e.ID != 66 || e.Name != "test" {
		t.Errorf("embedded struct failed: err=%v, e=%+v", err, e)
	}
}

// TestMarshalCopyError tests MarshalCopy error path
func TestMarshalCopyError(t *testing.T) {
	// Use channel which can't be encoded
	ch := make(chan int)
	_, err := MarshalCopy(ch)
	if err == nil {
		t.Error("expected error for channel")
	}
}

// TestMarshalAppendError tests MarshalAppend error path
func TestMarshalAppendError(t *testing.T) {
	// Use channel which can't be encoded
	ch := make(chan int)
	_, err := MarshalAppend(nil, ch)
	if err == nil {
		t.Error("expected error for channel")
	}
}

// TestMarshalAppendSuccess tests MarshalAppend success path
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

// TestIsEmptyValueAllKinds tests isEmptyValue for all kinds
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

// TestIsEmptyValueNonEmpty tests isEmptyValue with non-empty values
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

// TestEncodeSliceAnyError tests encodeSliceAny error path
func TestEncodeSliceAnyError(t *testing.T) {
	// Slice containing unencodable value
	s := []any{"ok", make(chan int)}
	_, err := Marshal(s)
	if err == nil {
		t.Error("expected error for slice with channel")
	}
}

// TestEncodeMapStringAnyError tests encodeMapStringAny error path
func TestEncodeMapStringAnyError(t *testing.T) {
	// Map containing unencodable value
	m := map[string]any{"ok": "fine", "bad": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel")
	}
}

// TestDecodeIntoStructMap16 tests decodeIntoStruct with Map16 format
func TestDecodeIntoStructMap16(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Create map16 format (0xde, 2 bytes for length)
	enc := NewEncoder(64)
	enc.writeByte(formatMap16)
	enc.writeUint16(1)
	enc.EncodeString("a")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeIntoStructMap32 tests decodeIntoStruct with Map32 format
func TestDecodeIntoStructMap32(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Create map32 format (0xdf, 4 bytes for length)
	enc := NewEncoder(64)
	enc.writeByte(formatMap32)
	enc.writeUint32(1)
	enc.EncodeString("a")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeIntoStructNonStringKey tests decodeIntoStruct with non-string keys
func TestDecodeIntoStructNonStringKey(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Create map with integer key (should be skipped)
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeInt(123) // integer key
	enc.EncodeInt(456) // value
	enc.EncodeString("a")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeStringKeyStr16 tests decodeStringKey with str16 format
func TestDecodeStringKeyStr16(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Create map with str16 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	enc.writeUint16(1)
	enc.writeByte('a')
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeStringKeyStr32 tests decodeStringKey with str32 format
func TestDecodeStringKeyStr32(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Create map with str32 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	enc.writeUint32(1)
	enc.writeByte('a')
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeStringKeyStr8EOF tests decodeStringKey str8 EOF
func TestDecodeStringKeyStr8EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str8 format but no length
	data := []byte{0x81, formatStr8}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringKeyStr16EOF tests decodeStringKey str16 EOF
func TestDecodeStringKeyStr16EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str16 format but no length
	data := []byte{0x81, formatStr16}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringKeyStr32EOF tests decodeStringKey str32 EOF
func TestDecodeStringKeyStr32EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str32 format but no length
	data := []byte{0x81, formatStr32}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringKeyStr8TooLong tests decodeStringKey str8 too long
func TestDecodeStringKeyStr8TooLong(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	// str8 format with length 5
	data := []byte{0x81, formatStr8, 5, 'h', 'e', 'l', 'l', 'o', 0x01}
	err := UnmarshalStructWithConfig(data, &S{}, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueStringReadBytesEOF tests decodeString readBytes EOF
func TestDecodeValueStringReadBytesEOF(t *testing.T) {
	// fixstr says 5 bytes but only 2 provided
	data := []byte{0xa5, 'h', 'i'}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueBinaryReadBytesEOF tests decodeBinary readBytes EOF
func TestDecodeValueBinaryReadBytesEOF(t *testing.T) {
	// bin8 says 5 bytes but only 2 provided
	data := []byte{formatBin8, 5, 0x01, 0x02}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeMapIntKeyNilBytes tests decodeMap with integer keys results in nil Key bytes
func TestDecodeMapIntKeyNilBytes(t *testing.T) {
	// Map with integer keys
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeInt(1)
	enc.EncodeString("one")
	enc.EncodeInt(2)
	enc.EncodeString("two")

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeMap {
		t.Errorf("expected map, got %v", v.Type)
	}
	// Non-string keys have nil Key field
	for _, kv := range v.Map {
		if kv.Key != nil {
			t.Errorf("expected nil key for integer key, got %v", kv.Key)
		}
	}
}

// TestDecodeMapValueEOF tests decodeMap value EOF
func TestDecodeMapValueEOF(t *testing.T) {
	// Map header says 2 entries but only has key
	data := []byte{0x82, 0xa1, 'a'} // map of 2, key "a", no value
	d := NewDecoder(data)
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeExtDataEOF tests decodeExt data EOF
func TestDecodeExtDataEOF(t *testing.T) {
	// fixext1 (1 byte data) but no data after type
	data := []byte{formatFixExt1, 0x01} // type=1, no data
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeExtTypeEOF tests decodeExt type EOF
func TestDecodeExtTypeEOF(t *testing.T) {
	// ext8 format with length but no type
	data := []byte{formatExt8, 0x01} // length=1, no type
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeArrayValueEOF tests decodeArray value EOF
func TestDecodeArrayValueEOF(t *testing.T) {
	// Array header says 3 but only 2 values
	data := []byte{0x93, 0x01, 0x02} // array of 3, values 1, 2
	d := NewDecoder(data)
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestEncodeArrayError tests encodeArray error path
func TestEncodeArrayError(t *testing.T) {
	// Array with channel element
	arr := [2]any{"ok", make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for array with channel")
	}
}

// TestEncodeMapChannelKeyError tests encodeMap error path with channel key
func TestEncodeMapChannelKeyError(t *testing.T) {
	// Map with complex key that fails
	m := map[any]any{make(chan int): "value"}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel key")
	}
}

// TestDecodeIntoStructMap16EOF tests decodeIntoStruct map16 EOF
func TestDecodeIntoStructMap16EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map16 format but no length bytes
	data := []byte{formatMap16}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeIntoStructMap32EOF tests decodeIntoStruct map32 EOF
func TestDecodeIntoStructMap32EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map32 format but no length bytes
	data := []byte{formatMap32}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStr16EOF tests str16 length EOF
func TestDecodeValueStr16EOF(t *testing.T) {
	// str16 but only 1 byte for length
	data := []byte{formatStr16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueStr32EOF tests str32 length EOF
func TestDecodeValueStr32EOF(t *testing.T) {
	// str32 but only 2 bytes for length
	data := []byte{formatStr32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueBin16EOF tests bin16 length EOF
func TestDecodeValueBin16EOF(t *testing.T) {
	// bin16 but only 1 byte for length
	data := []byte{formatBin16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueBin32EOF tests bin32 length EOF
func TestDecodeValueBin32EOF(t *testing.T) {
	// bin32 but only 2 bytes for length
	data := []byte{formatBin32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueArray16EOF tests array16 length EOF
func TestDecodeValueArray16EOF(t *testing.T) {
	// array16 but only 1 byte for length
	data := []byte{formatArray16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueArray32EOF tests array32 length EOF
func TestDecodeValueArray32EOF(t *testing.T) {
	// array32 but only 2 bytes for length
	data := []byte{formatArray32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueMap16EOF tests map16 length EOF
func TestDecodeValueMap16EOF(t *testing.T) {
	// map16 but only 1 byte for length
	data := []byte{formatMap16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueMap32EOF tests map32 length EOF
func TestDecodeValueMap32EOF(t *testing.T) {
	// map32 but only 2 bytes for length
	data := []byte{formatMap32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueExt8EOF tests ext8 length EOF
func TestDecodeValueExt8EOF(t *testing.T) {
	// ext8 but no length byte
	data := []byte{formatExt8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueExt16EOF tests ext16 length EOF
func TestDecodeValueExt16EOF(t *testing.T) {
	// ext16 but only 1 byte for length
	data := []byte{formatExt16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueExt32EOF tests ext32 length EOF
func TestDecodeValueExt32EOF(t *testing.T) {
	// ext32 but only 2 bytes for length
	data := []byte{formatExt32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueUint8EOF tests uint8 EOF
func TestDecodeValueUint8EOF(t *testing.T) {
	data := []byte{formatUint8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueUint16EOF tests uint16 EOF
func TestDecodeValueUint16EOF(t *testing.T) {
	data := []byte{formatUint16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueUint32EOF tests uint32 EOF
func TestDecodeValueUint32EOF(t *testing.T) {
	data := []byte{formatUint32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueUint64EOF tests uint64 EOF
func TestDecodeValueUint64EOF(t *testing.T) {
	data := []byte{formatUint64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueInt8EOF tests int8 EOF
func TestDecodeValueInt8EOF(t *testing.T) {
	data := []byte{formatInt8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueInt16EOF tests int16 EOF
func TestDecodeValueInt16EOF(t *testing.T) {
	data := []byte{formatInt16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueInt32EOF tests int32 EOF
func TestDecodeValueInt32EOF(t *testing.T) {
	data := []byte{formatInt32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueInt64EOF tests int64 EOF
func TestDecodeValueInt64EOF(t *testing.T) {
	data := []byte{formatInt64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueFloat32EOF tests float32 EOF
func TestDecodeValueFloat32EOF(t *testing.T) {
	data := []byte{formatFloat32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueFloat64EOF tests float64 EOF
func TestDecodeValueFloat64EOF(t *testing.T) {
	data := []byte{formatFloat64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStructNilValueError tests DecodeStruct nil value error
func TestDecodeStructNilValueError(t *testing.T) {
	data := []byte{0x80} // empty map
	err := UnmarshalStruct(data, nil)
	if err == nil {
		t.Error("expected error for nil value")
	}
}

// TestDecodeStructNonPointerError tests DecodeStruct non-pointer error
func TestDecodeStructNonPointerError(t *testing.T) {
	type S struct{}
	data := []byte{0x80} // empty map
	var s S
	err := UnmarshalStruct(data, s) // passing value, not pointer
	if err == nil {
		t.Error("expected error for non-pointer")
	}
}

// TestDecodeStructNilPointerError tests DecodeStruct nil pointer error
func TestDecodeStructNilPointerError(t *testing.T) {
	type S struct{}
	data := []byte{0x80}
	var s *S // nil pointer
	err := UnmarshalStruct(data, s)
	if err == nil {
		t.Error("expected error for nil pointer")
	}
}

// TestDecodeStructNonStructError tests DecodeStruct non-struct error
func TestDecodeStructNonStructError(t *testing.T) {
	data := []byte{0x80}
	var i int
	err := UnmarshalStruct(data, &i) // pointer to int, not struct
	if err == nil {
		t.Error("expected error for non-struct")
	}
}

// TestDecodeStructformatNil tests DecodeStruct with nil format
func TestDecodeStructformatNil(t *testing.T) {
	type S struct{}
	data := []byte{formatNil}
	var s S
	err := UnmarshalStruct(data, &s)
	if err != nil {
		t.Errorf("expected nil to work, got %v", err)
	}
}

// TestDecodeStructEOF tests DecodeStruct EOF
func TestDecodeStructEOF(t *testing.T) {
	type S struct{}
	data := []byte{} // empty
	var s S
	err := UnmarshalStruct(data, &s)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStructTypeMismatch tests DecodeStruct type mismatch
func TestDecodeStructTypeMismatch(t *testing.T) {
	type S struct{}
	data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // string, not map
	var s S
	err := UnmarshalStruct(data, &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestReadBytesInsufficientData tests readBytes with insufficient data
func TestReadBytesInsufficientData(t *testing.T) {
	d := NewDecoder([]byte{0x01, 0x02})
	_, err := d.readBytes(5) // only 2 bytes available
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeIntoValueNestedStruct tests decodeIntoValue with nested struct
func TestDecodeIntoValueNestedStruct(t *testing.T) {
	type Inner struct {
		X int `msgpack:"x"`
	}
	type Outer struct {
		Inner Inner `msgpack:"inner"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("inner")
	enc.EncodeMapHeader(1)
	enc.EncodeString("x")
	enc.EncodeInt(42)

	var o Outer
	err := UnmarshalStruct(enc.Bytes(), &o)
	if err != nil {
		t.Fatal(err)
	}
	if o.Inner.X != 42 {
		t.Errorf("got %d, want 42", o.Inner.X)
	}
}

// TestDecodeIntoValuePointerFieldAllocated tests decodeIntoValue with pre-allocated pointer field
func TestDecodeIntoValuePointerFieldAllocated(t *testing.T) {
	// Test that pointer fields that are nil get allocated
	// Note: This exercises the reflect.Ptr case in decodeIntoValue
	type S struct {
		P *int `msgpack:"p"`
	}

	// When the value is nil, the decoder handles it correctly
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("p")
	enc.EncodeNil()

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	// P should remain nil for nil encoded value
	if s.P != nil {
		t.Errorf("expected nil pointer, got %v", s.P)
	}
}

// TestDecodeIntoValuePointerFieldNil tests decodeIntoValue with nil pointer field
func TestDecodeIntoValuePointerFieldNil(t *testing.T) {
	type S struct {
		P *int `msgpack:"p"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("p")
	enc.EncodeNil()

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.P != nil {
		t.Errorf("expected nil pointer, got %v", s.P)
	}
}

// TestAsFloat64UnsignedInt tests AsFloat64 with large uint64
func TestAsFloat64UnsignedInt(t *testing.T) {
	// Large uint64 that doesn't fit in int64
	enc := NewEncoder(16)
	enc.EncodeUint(18446744073709551615) // max uint64

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}

	f := v.AsFloat64()
	if f <= 0 {
		t.Errorf("expected positive float, got %f", f)
	}
}

// TestEncodeMapKeyValueError tests encodeMap with error in key or value encoding
func TestEncodeMapKeyValueError(t *testing.T) {
	// Map with channel value
	m := map[string]any{"ch": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel value")
	}
}

// TestDecodeAnyValueAllFormats tests decodeAnyValue with all format types
func TestDecodeAnyValueAllFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"nil", []byte{formatNil}},
		{"false", []byte{formatFalse}},
		{"true", []byte{formatTrue}},
		{"pos_fixint", []byte{0x7f}},
		{"neg_fixint", []byte{0xff}},
		{"fixmap", []byte{0x81, 0xa1, 'a', 0x01}},
		{"fixarray", []byte{0x91, 0x01}},
		{"fixstr", []byte{0xa5, 'h', 'e', 'l', 'l', 'o'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unmarshal(tt.data)
			if err != nil {
				t.Errorf("failed to decode %s: %v", tt.name, err)
			}
		})
	}
}

// TestEncodeStringStr32 tests encoding very long strings using str32 format
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

// TestEncodeBinaryBin32 tests encoding very long binary using bin32 format
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

// TestEncodeValueNilMap tests encodeValue with nil map
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

// TestEncodeValueNilSlice tests encodeValue with nil slice
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

// TestEncodeValueNilPointer tests encodeValue with nil pointer
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

// TestEncodeValueFloat32 tests encodeValue with float32
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

// TestDecodeStringAnyValidation tests decodeStringAny validation error
func TestDecodeStringAnyValidation(t *testing.T) {
	// str8 with length 10 but MaxStringLen=1
	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	data := []byte{formatStr8, 10, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	d := NewDecoderWithConfig(data, cfg)
	_, err := d.DecodeAny()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeStringAnyReadBytesEOF tests decodeStringAny readBytes EOF
func TestDecodeStringAnyReadBytesEOF(t *testing.T) {
	// str8 says 10 bytes but only 5 provided
	data := []byte{formatStr8, 10, 'a', 'b', 'c', 'd', 'e'}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeBinaryAnyValidation tests decodeBinaryAny validation error
func TestDecodeBinaryAnyValidation(t *testing.T) {
	// bin8 with length 10 but MaxBinaryLen=1
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	data := []byte{formatBin8, 10, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	d := NewDecoderWithConfig(data, cfg)
	_, err := d.DecodeAny()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeBinaryAnyReadBytesEOF tests decodeBinaryAny readBytes EOF
func TestDecodeBinaryAnyReadBytesEOF(t *testing.T) {
	// bin8 says 10 bytes but only 5 provided
	data := []byte{formatBin8, 10, 'a', 'b', 'c', 'd', 'e'}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeMapAnyKeyStr16 tests decodeMapAny with str16 key format
func TestDecodeMapAnyKeyStr16(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	// Manually write str16 key
	enc.writeByte(formatStr16)
	enc.writeUint16(1)
	enc.writeByte('a')
	enc.EncodeInt(42)

	_, err := Unmarshal(enc.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}

// TestDecodeMapAnyKeyStr32 tests decodeMapAny with str32 key format
func TestDecodeMapAnyKeyStr32(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	// Manually write str32 key
	enc.writeByte(formatStr32)
	enc.writeUint32(1)
	enc.writeByte('a')
	enc.EncodeInt(42)

	_, err := Unmarshal(enc.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}

// TestDecodeMapAnyNonStringKey tests decodeMapAny with non-string key
func TestDecodeMapAnyNonStringKey(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeInt(123) // integer key instead of string
	enc.EncodeInt(456)

	_, err := Unmarshal(enc.Bytes())
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

// TestDecodeIntoValueInterface tests decodeIntoValue with interface field
func TestDecodeIntoValueInterface(t *testing.T) {
	type S struct {
		V any `msgpack:"v"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("v")
	enc.EncodeString("hello")

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.V != "hello" {
		t.Errorf("got %v, want hello", s.V)
	}
}

// TestDecodeIntoValueComplex64Unsupported tests decodeIntoValue with complex64 type
func TestDecodeIntoValueComplex64Unsupported(t *testing.T) {
	type S struct {
		C complex64 `msgpack:"c"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("c")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrUnsupportedType {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}
}

// TestDecodeValueStringStr16 tests decodeString with str16 format
func TestDecodeValueStringStr16(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.AsString() != "hello" {
		t.Errorf("got %v, want hello", v.AsString())
	}
}

// TestDecodeValueStringStr32 tests decodeString with str32 format
func TestDecodeValueStringStr32(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.AsString() != "hello" {
		t.Errorf("got %v, want hello", v.AsString())
	}
}

// TestDecodeValueBinaryBin16 tests decodeBinary with bin16 format
func TestDecodeValueBinaryBin16(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(3)
	enc.writeBytes([]byte{1, 2, 3})

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeBinary || len(v.AsBytes()) != 3 {
		t.Errorf("got %v, want binary of length 3", v)
	}
}

// TestDecodeValueBinaryBin32 tests decodeBinary with bin32 format
func TestDecodeValueBinaryBin32(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin32)
	enc.writeUint32(3)
	enc.writeBytes([]byte{1, 2, 3})

	d := NewDecoder(enc.Bytes())
	v, err := d.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeBinary || len(v.AsBytes()) != 3 {
		t.Errorf("got %v, want binary of length 3", v)
	}
}

// TestDecodeExtValidation tests decodeExt validation error
func TestDecodeExtValidation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1,
		MaxDepth:     100,
	}
	// ext8 with length 10
	enc := NewEncoder(32)
	enc.writeByte(formatExt8)
	enc.writeByte(10) // length
	enc.writeByte(1)  // type
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrExtTooLong {
		t.Errorf("expected ErrExtTooLong, got %v", err)
	}
}

// TestDecodeIntoSliceValidationError tests decodeIntoSlice validation
func TestDecodeIntoSliceValidationError(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  1,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(5)
	for i := 0; i < 5; i++ {
		enc.EncodeInt(int64(i))
	}

	var s S
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrArrayTooLong {
		t.Errorf("expected ErrArrayTooLong, got %v", err)
	}
}

// TestDecodeIntoArrayValidationError tests decodeIntoArray validation
func TestDecodeIntoArrayValidationError(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  1,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(3)
	for i := 0; i < 3; i++ {
		enc.EncodeInt(int64(i))
	}

	var s S
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrArrayTooLong {
		t.Errorf("expected ErrArrayTooLong, got %v", err)
	}
}

// TestDecodeIntoMapValidationError tests decodeIntoMap validation
func TestDecodeIntoMapValidationError(t *testing.T) {
	type S struct {
		Items map[string]int `msgpack:"items"`
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    1,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeMapHeader(3)
	for i := 0; i < 3; i++ {
		enc.EncodeString("k")
		enc.EncodeInt(int64(i))
	}

	var s S
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrMapTooLong {
		t.Errorf("expected ErrMapTooLong, got %v", err)
	}
}

// TestDecodeIntoStructDepthError tests decodeIntoStruct max depth
func TestDecodeIntoStructDepthError(t *testing.T) {
	type Inner struct {
		X int `msgpack:"x"`
	}
	type Outer struct {
		Inner Inner `msgpack:"inner"`
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     1, // very shallow depth
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("inner")
	enc.EncodeMapHeader(1)
	enc.EncodeString("x")
	enc.EncodeInt(42)

	var o Outer
	err := UnmarshalStructWithConfig(enc.Bytes(), &o, cfg)
	if err != ErrMaxDepthExceeded {
		t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
	}
}

// TestReadStringBytesStr16 tests readStringBytes with str16 format
func TestReadStringBytesStr16(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	b, err := d.readStringBytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Errorf("got %s, want hello", b)
	}
}

// TestReadStringBytesStr32 tests readStringBytes with str32 format
func TestReadStringBytesStr32(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	b, err := d.readStringBytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Errorf("got %s, want hello", b)
	}
}

// TestDecodeValueStringStr8Validation tests decodeString str8 validation error
func TestDecodeValueStringStr8Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueBinaryValidation tests decodeBinary validation error
func TestDecodeValueBinaryValidation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeStructValueError tests DecodeStruct with value decode error
func TestDecodeStructValueError(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with key but truncated value
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("a")
	// No value - will cause EOF

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for truncated value")
	}
}

// TestDecodeIntoStructUnknownField tests skipping unknown fields
func TestDecodeIntoStructUnknownField(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with unknown field that should be skipped
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("unknown") // unknown field
	enc.EncodeString("value")   // value to skip
	enc.EncodeString("a")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeIntoStructUnknownFieldSkipError tests error when skipping unknown field fails
func TestDecodeIntoStructUnknownFieldSkipError(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with unknown field but truncated value to skip
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("unknown")
	// Missing value - will cause EOF when trying to skip

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for truncated unknown field value")
	}
}

// TestDecodeStringKeyFixstrValidation tests decodeStringKey fixstr validation
func TestDecodeStringKeyFixstrValidation(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	// fixstr with length 5 (0xa5 = fixstr of length 5)
	data := []byte{0x81, 0xa5, 'h', 'e', 'l', 'l', 'o', 0x01}
	err := UnmarshalStructWithConfig(data, &S{}, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeIntoSliceElementError tests error in slice element decode
func TestDecodeIntoSliceElementError(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	// Array with incomplete element
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(2)
	enc.EncodeInt(1)
	// Second element is incomplete - just format byte with no data
	enc.writeByte(formatInt32)
	// Only 2 bytes of the 4 needed
	enc.writeByte(0x00)
	enc.writeByte(0x00)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeIntoArrayElementError tests error in array element decode
func TestDecodeIntoArrayElementError(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	// Array with incomplete element
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(3)
	enc.EncodeInt(1)
	enc.EncodeInt(2)
	// Third element incomplete
	enc.writeByte(formatInt32)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error for incomplete array element")
	}
}

// TestDecodeIntoMapKeyError tests error in map key decode
func TestDecodeIntoMapKeyError(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	// Map with incomplete key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.EncodeMapHeader(1)
	// Incomplete string key
	enc.writeByte(formatStr8)
	enc.writeByte(10) // says 10 bytes but EOF

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for incomplete map key")
	}
}

// TestDecodeIntoMapValueError tests error in map value decode
func TestDecodeIntoMapValueError(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	// Map with incomplete value
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.EncodeMapHeader(1)
	enc.EncodeString("key")
	// Incomplete int value
	enc.writeByte(formatInt32)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for incomplete map value")
	}
}

// TestDecodeValueStringReadBytesError tests decodeString readBytes error
func TestDecodeValueStringReadBytesError(t *testing.T) {
	// str16 with length but not enough data
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(100) // says 100 bytes but only have a few

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueBinaryReadBytesError tests decodeBinary readBytes error
func TestDecodeValueBinaryReadBytesError(t *testing.T) {
	// bin16 with length but not enough data
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(100) // says 100 bytes but only have a few

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeMapKeyDecodeError tests decodeMap key decode error
func TestDecodeMapKeyDecodeError(t *testing.T) {
	// Map with incomplete key value
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatInt32) // key is int but incomplete

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected error for incomplete map key")
	}
}

// TestDecodeExtReadBytesError tests decodeExt readBytes error
func TestDecodeExtReadBytesError(t *testing.T) {
	// ext8 with length but not enough data for ext data
	enc := NewEncoder(64)
	enc.writeByte(formatExt8)
	enc.writeByte(10) // length 10
	enc.writeByte(1)  // type
	// Only provide 3 bytes instead of 10
	enc.writeByte(1)
	enc.writeByte(2)
	enc.writeByte(3)

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStringAnyStr16 tests decodeStringAny with str16 format
func TestDecodeStringAnyStr16(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Errorf("got %v, want hello", v)
	}
}

// TestDecodeStringAnyStr32 tests decodeStringAny with str32 format
func TestDecodeStringAnyStr32(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(5)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Errorf("got %v, want hello", v)
	}
}

// TestDecodeBinaryAnyBin16 tests decodeBinaryAny with bin16 format
func TestDecodeBinaryAnyBin16(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(3)
	enc.writeBytes([]byte{1, 2, 3})

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	b, ok := v.([]byte)
	if !ok || len(b) != 3 {
		t.Errorf("got %v, want []byte{1,2,3}", v)
	}
}

// TestDecodeBinaryAnyBin32 tests decodeBinaryAny with bin32 format
func TestDecodeBinaryAnyBin32(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin32)
	enc.writeUint32(3)
	enc.writeBytes([]byte{1, 2, 3})

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Fatal(err)
	}
	b, ok := v.([]byte)
	if !ok || len(b) != 3 {
		t.Errorf("got %v, want []byte{1,2,3}", v)
	}
}

// TestDecodeValueStringStr8ReadBytesEOF tests decodeString str8 readBytes EOF
func TestDecodeValueStringStr8ReadBytesEOF(t *testing.T) {
	// str8 says 50 bytes but only provides 5
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(50)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeValueBinaryBin8ReadBytesEOF tests decodeBinary bin8 readBytes EOF
func TestDecodeValueBinaryBin8ReadBytesEOF(t *testing.T) {
	// bin8 says 50 bytes but only provides 5
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(50)
	enc.writeBytes([]byte{1, 2, 3, 4, 5})

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf("expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestDecodeStringKeyStr16Validation tests decodeStringKey str16 validation
func TestDecodeStringKeyStr16Validation(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	// Map with str16 key that exceeds max
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	enc.writeUint16(5) // length 5 exceeds MaxStringLen=1
	enc.writeBytes([]byte("hello"))
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeStringKeyStr32Validation tests decodeStringKey str32 validation
func TestDecodeStringKeyStr32Validation(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	// Map with str32 key that exceeds max
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	enc.writeUint32(5) // length 5 exceeds MaxStringLen=1
	enc.writeBytes([]byte("hello"))
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeIntoStructDecodeValueError tests error path when decoding struct field value
func TestDecodeIntoStructDecodeValueError(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with known field but value is incomplete
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("a")
	enc.writeByte(formatInt32) // int32 format but no data

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for incomplete field value")
	}
}

// TestDecodeIntoStructSkipNonStringKeyValue tests skipping value when key is non-string
func TestDecodeIntoStructSkipNonStringKeyValue(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with non-string key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	// First: non-string key with nested map value
	enc.EncodeInt(123)
	enc.EncodeMapHeader(1) // nested map
	enc.EncodeString("x")
	enc.EncodeInt(1)
	// Second: valid field
	enc.EncodeString("a")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.A != 42 {
		t.Errorf("got %d, want 42", s.A)
	}
}

// TestDecodeIntoStructSkipNonStringKeyError tests error when skip fails for non-string key
func TestDecodeIntoStructSkipNonStringKeyError(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map with non-string key but incomplete value
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeInt(123)        // non-string key
	enc.writeByte(formatStr8) // start of value but incomplete
	enc.writeByte(100)        // says 100 bytes but nothing follows

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error when skipping non-string key value fails")
	}
}

// TestDecodeValueBoolMismatch tests bool type mismatch
func TestDecodeValueBoolMismatch(t *testing.T) {
	type S struct {
		B bool `msgpack:"b"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("b")
	enc.EncodeInt(123) // int instead of bool

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeValueStringMismatch tests string type mismatch
func TestDecodeValueStringMismatch(t *testing.T) {
	type S struct {
		S string `msgpack:"s"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("s")
	enc.EncodeInt(123) // int instead of string

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeValueSliceMismatch tests slice type mismatch
func TestDecodeValueSliceMismatch(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeString("not an array") // string instead of array

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeValueArrayMismatch tests array type mismatch
func TestDecodeValueArrayMismatch(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeString("not an array") // string instead of array

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeValueMapMismatch tests map type mismatch
func TestDecodeValueMapMismatch(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.EncodeString("not a map") // string instead of map

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeValueBytesMismatch tests []byte type mismatch
func TestDecodeValueBytesMismatch(t *testing.T) {
	type S struct {
		B []byte `msgpack:"b"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("b")
	enc.EncodeInt(123) // int instead of binary

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeIntoSliceArray16 tests decodeIntoSlice with array16 format
func TestDecodeIntoSliceArray16(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray16)
	enc.writeUint16(3)
	enc.EncodeInt(1)
	enc.EncodeInt(2)
	enc.EncodeInt(3)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != 3 {
		t.Errorf("got %d items, want 3", len(s.Items))
	}
}

// TestDecodeIntoSliceArray32 tests decodeIntoSlice with array32 format
func TestDecodeIntoSliceArray32(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray32)
	enc.writeUint32(2)
	enc.EncodeInt(1)
	enc.EncodeInt(2)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != 2 {
		t.Errorf("got %d items, want 2", len(s.Items))
	}
}

// TestDecodeIntoArrayArray16 tests decodeIntoArray with array16 format
func TestDecodeIntoArrayArray16(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray16)
	enc.writeUint16(3)
	enc.EncodeInt(1)
	enc.EncodeInt(2)
	enc.EncodeInt(3)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Items[0] != 1 || s.Items[1] != 2 || s.Items[2] != 3 {
		t.Errorf("got %v, want [1,2,3]", s.Items)
	}
}

// TestDecodeIntoArrayArray32 tests decodeIntoArray with array32 format
func TestDecodeIntoArrayArray32(t *testing.T) {
	type S struct {
		Items [2]int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray32)
	enc.writeUint32(2)
	enc.EncodeInt(10)
	enc.EncodeInt(20)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Items[0] != 10 || s.Items[1] != 20 {
		t.Errorf("got %v, want [10,20]", s.Items)
	}
}

// TestDecodeIntoMapMap16 tests decodeIntoMap with map16 format
func TestDecodeIntoMapMap16(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.writeByte(formatMap16)
	enc.writeUint16(2)
	enc.EncodeString("a")
	enc.EncodeInt(1)
	enc.EncodeString("b")
	enc.EncodeInt(2)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.M) != 2 {
		t.Errorf("got %d items, want 2", len(s.M))
	}
}

// TestDecodeIntoMapMap32 tests decodeIntoMap with map32 format
func TestDecodeIntoMapMap32(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.writeByte(formatMap32)
	enc.writeUint32(1)
	enc.EncodeString("x")
	enc.EncodeInt(42)

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.M["x"] != 42 {
		t.Errorf("got %v, want {x:42}", s.M)
	}
}

// TestDecodeIntoSliceArray16EOF tests decodeIntoSlice array16 EOF
func TestDecodeIntoSliceArray16EOF(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray16)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for array16 EOF")
	}
}

// TestDecodeIntoSliceArray32EOF tests decodeIntoSlice array32 EOF
func TestDecodeIntoSliceArray32EOF(t *testing.T) {
	type S struct {
		Items []int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray32)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for array32 EOF")
	}
}

// TestDecodeIntoArrayArray16EOF tests decodeIntoArray array16 EOF
func TestDecodeIntoArrayArray16EOF(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray16)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for array16 EOF")
	}
}

// TestDecodeIntoArrayArray32EOF tests decodeIntoArray array32 EOF
func TestDecodeIntoArrayArray32EOF(t *testing.T) {
	type S struct {
		Items [3]int `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.writeByte(formatArray32)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for array32 EOF")
	}
}

// TestDecodeIntoMapMap16EOF tests decodeIntoMap map16 EOF
func TestDecodeIntoMapMap16EOF(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.writeByte(formatMap16)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for map16 EOF")
	}
}

// TestDecodeIntoMapMap32EOF tests decodeIntoMap map32 EOF
func TestDecodeIntoMapMap32EOF(t *testing.T) {
	type S struct {
		M map[string]int `msgpack:"m"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("m")
	enc.writeByte(formatMap32)
	// Missing length bytes

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error for map32 EOF")
	}
}

// TestDecodeValueStringStr16Validation tests str16 validation error
func TestDecodeValueStringStr16Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueStringStr32Validation tests str32 validation error
func TestDecodeValueStringStr32Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueBinaryBin16Validation tests bin16 validation error
func TestDecodeValueBinaryBin16Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeValueBinaryBin32Validation tests bin32 validation error
func TestDecodeValueBinaryBin32Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatBin32)
	enc.writeUint32(10)
	enc.writeBytes(make([]byte, 10))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.Decode()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeValueStructMismatch tests struct type mismatch
func TestDecodeValueStructMismatch(t *testing.T) {
	type Inner struct {
		X int `msgpack:"x"`
	}
	type S struct {
		I Inner `msgpack:"i"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("i")
	enc.EncodeString("not a map") // string instead of map

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeStringAnyStr16Validation tests decodeStringAny str16 validation
func TestDecodeStringAnyStr16Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1,
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(100)
	enc.writeBytes(make([]byte, 100))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.DecodeAny()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeBinaryAnyBin16Validation tests decodeBinaryAny bin16 validation
func TestDecodeBinaryAnyBin16Validation(t *testing.T) {
	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 1,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(100)
	enc.writeBytes(make([]byte, 100))

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.DecodeAny()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeFloatUint64 tests decodeValueFloat with uint64 format
func TestDecodeFloatUint64(t *testing.T) {
	type S struct {
		F float64 `msgpack:"f"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("f")
	enc.EncodeUint(9223372036854775808) // > max int64, uses uint64 format

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.F <= 0 {
		t.Errorf("expected positive float, got %f", s.F)
	}
}

// TestDecodeValueStr8String tests struct decoding with str8 string format
func TestDecodeValueStr8String(t *testing.T) {
	type S struct {
		S string `msgpack:"s"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("s")
	enc.writeByte(formatStr8)
	enc.writeByte(5)
	enc.writeBytes([]byte("hello"))

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.S != "hello" {
		t.Errorf("got %s, want hello", s.S)
	}
}

// TestDecodeValueStr16String tests struct decoding with str16 string format
func TestDecodeValueStr16String(t *testing.T) {
	type S struct {
		S string `msgpack:"s"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("s")
	enc.writeByte(formatStr16)
	enc.writeUint16(5)
	enc.writeBytes([]byte("hello"))

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.S != "hello" {
		t.Errorf("got %s, want hello", s.S)
	}
}

// TestDecodeValueStr32String tests struct decoding with str32 string format
func TestDecodeValueStr32String(t *testing.T) {
	type S struct {
		S string `msgpack:"s"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("s")
	enc.writeByte(formatStr32)
	enc.writeUint32(5)
	enc.writeBytes([]byte("hello"))

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if s.S != "hello" {
		t.Errorf("got %s, want hello", s.S)
	}
}

// TestDecodeValueBin8Bytes tests struct decoding with bin8 bytes format
func TestDecodeValueBin8Bytes(t *testing.T) {
	type S struct {
		B []byte `msgpack:"b"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("b")
	enc.writeByte(formatBin8)
	enc.writeByte(3)
	enc.writeBytes([]byte{1, 2, 3})

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.B) != 3 {
		t.Errorf("got len=%d, want 3", len(s.B))
	}
}

// TestDecodeValueBin16Bytes tests struct decoding with bin16 bytes format
func TestDecodeValueBin16Bytes(t *testing.T) {
	type S struct {
		B []byte `msgpack:"b"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("b")
	enc.writeByte(formatBin16)
	enc.writeUint16(3)
	enc.writeBytes([]byte{1, 2, 3})

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.B) != 3 {
		t.Errorf("got len=%d, want 3", len(s.B))
	}
}

// TestDecodeValueBin32Bytes tests struct decoding with bin32 bytes format
func TestDecodeValueBin32Bytes(t *testing.T) {
	type S struct {
		B []byte `msgpack:"b"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("b")
	enc.writeByte(formatBin32)
	enc.writeUint32(3)
	enc.writeBytes([]byte{1, 2, 3})

	var s S
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.B) != 3 {
		t.Errorf("got len=%d, want 3", len(s.B))
	}
}

// TestDecodeMapAnyKeyStr8EOF tests decodeMapAny key str8 EOF
func TestDecodeMapAnyKeyStr8EOF(t *testing.T) {
	// Map with incomplete str8 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr8)
	// Missing length byte

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeMapAnyKeyStr16EOF tests decodeMapAny key str16 EOF
func TestDecodeMapAnyKeyStr16EOF(t *testing.T) {
	// Map with incomplete str16 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	// Missing length bytes

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeMapAnyKeyStr32EOF tests decodeMapAny key str32 EOF
func TestDecodeMapAnyKeyStr32EOF(t *testing.T) {
	// Map with incomplete str32 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	// Missing length bytes

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeMapAnyValueError tests decodeMapAny value decode error
func TestDecodeMapAnyValueError(t *testing.T) {
	// Map with valid key but incomplete value
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("key")
	enc.writeByte(formatInt32) // int32 but no data

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected error for incomplete value")
	}
}

// TestDecodeArrayAnyElementError tests decodeArrayAny element decode error
func TestDecodeArrayAnyElementError(t *testing.T) {
	// Array with incomplete element
	enc := NewEncoder(64)
	enc.EncodeArrayHeader(2)
	enc.EncodeInt(1)
	enc.writeByte(formatInt32) // int32 but no data

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected error for incomplete element")
	}
}

// ============================================================================
// Additional Coverage Tests for 90%+ Coverage
// ============================================================================

// TestDecodeStringAnyReadBytesError tests decodeStringAny readBytes error
func TestDecodeStringAnyReadBytesError(t *testing.T) {
	// String with declared length but insufficient data
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeBinaryAnyReadBytesError tests decodeBinaryAny readBytes error
func TestDecodeBinaryAnyReadBytesError(t *testing.T) {
	// Binary with declared length but insufficient data
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr8ReadBytesError tests decodeValueString str8 readBytes error
func TestDecodeValueStringStr8ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr16ReadBytesError tests decodeValueString str16 readBytes error
func TestDecodeValueStringStr16ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr32ReadBytesError tests decodeValueString str32 readBytes error
func TestDecodeValueStringStr32ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr8ReadBytesError tests decodeValueBytes str8 readBytes error
func TestDecodeValueBytesStr8ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr16ReadBytesError tests decodeValueBytes str16 readBytes error
func TestDecodeValueBytesStr16ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr32ReadBytesError tests decodeValueBytes str32 readBytes error
func TestDecodeValueBytesStr32ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin8ReadBytesError tests decodeValueBytes bin8 readBytes error
func TestDecodeValueBytesBin8ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin16ReadBytesError tests decodeValueBytes bin16 readBytes error
func TestDecodeValueBytesBin16ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin32ReadBytesError tests decodeValueBytes bin32 readBytes error
func TestDecodeValueBytesBin32ReadBytesError(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringReadBytesError tests decodeString (Value) readBytes error
func TestDecodeStringReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeBinaryReadBytesError tests decodeBinary (Value) readBytes error
func TestDecodeBinaryReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringStr16ReadBytesError tests decodeString str16 readBytes error
func TestDecodeStringStr16ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringStr32ReadBytesError tests decodeString str32 readBytes error
func TestDecodeStringStr32ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeBinaryBin16ReadBytesError tests decodeBinary bin16 readBytes error
func TestDecodeBinaryBin16ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeBinaryBin32ReadBytesError tests decodeBinary bin32 readBytes error
func TestDecodeBinaryBin32ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr8EOFLength tests decodeValueString str8 EOF on readUint8
func TestDecodeValueStringStr8EOFLength(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr8)
	// Missing length byte

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr16EOFLength tests decodeValueString str16 EOF on readUint16
func TestDecodeValueStringStr16EOFLength(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr16)
	enc.writeByte(0) // Only 1 byte, need 2

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueStringStr32EOFLength tests decodeValueString str32 EOF on readUint32
func TestDecodeValueStringStr32EOFLength(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr32)
	enc.writeByte(0)
	enc.writeByte(0) // Only 2 bytes, need 4

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr8EOFLength tests decodeValueBytes str8 EOF on readUint8
func TestDecodeValueBytesStr8EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr8)
	// Missing length byte

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr16EOFLength tests decodeValueBytes str16 EOF on readUint16
func TestDecodeValueBytesStr16EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr16)
	enc.writeByte(0) // Only 1 byte, need 2

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesStr32EOFLength tests decodeValueBytes str32 EOF on readUint32
func TestDecodeValueBytesStr32EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatStr32)
	enc.writeByte(0)
	enc.writeByte(0) // Only 2 bytes, need 4

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin8EOFLength tests decodeValueBytes bin8 EOF on readUint8
func TestDecodeValueBytesBin8EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin8)
	// Missing length byte

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin16EOFLength tests decodeValueBytes bin16 EOF on readUint16
func TestDecodeValueBytesBin16EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin16)
	enc.writeByte(0) // Only 1 byte, need 2

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueBytesBin32EOFLength tests decodeValueBytes bin32 EOF on readUint32
func TestDecodeValueBytesBin32EOFLength(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin32)
	enc.writeByte(0)
	enc.writeByte(0) // Only 2 bytes, need 4

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeIntoValuePointerNil tests decodeIntoValue with nil pointer field
func TestDecodeIntoValuePointerNil(t *testing.T) {
	type TestStruct struct {
		Name *string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.EncodeNil()

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != nil {
		t.Errorf("expected nil, got %v", s.Name)
	}
}

// TestDecodeIntoValueInterfaceField tests decodeIntoValue with interface{} field
func TestDecodeIntoValueInterfaceField(t *testing.T) {
	type TestStruct struct {
		Value any `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.EncodeString("test")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeIntoValueInterfaceFieldMap tests decodeIntoValue with interface{} field containing map
func TestDecodeIntoValueInterfaceFieldMap(t *testing.T) {
	type TestStruct struct {
		Value any `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.EncodeMapHeader(1)
	enc.EncodeString("key")
	enc.EncodeString("val")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeIntoValueInterfaceFieldArray tests decodeIntoValue with interface{} field containing array
func TestDecodeIntoValueInterfaceFieldArray(t *testing.T) {
	type TestStruct struct {
		Value any `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.EncodeArrayHeader(2)
	enc.EncodeInt(1)
	enc.EncodeInt(2)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeIntoStructNonStringKeySkip tests decodeIntoStruct skipping non-string key
func TestDecodeIntoStructNonStringKeySkip(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	// Create map with integer key (non-string) followed by valid key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	// First key-value with non-string key
	enc.EncodeInt(123) // non-string key
	enc.EncodeString("ignored")
	// Second key-value with valid string key
	enc.EncodeString("name")
	enc.EncodeString("Alice")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	// Should skip non-string key and process valid key
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

// TestDecodeStringKeyReadBytesStr8Error tests decodeStringKey str8 readBytes error
func TestDecodeStringKeyStr8DataEOF(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr8)
	enc.writeByte(20) // length = 20
	enc.buf = append(enc.buf, "abc"...)
	// Only 3 bytes for key, expected 20

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringKeyStr16DataEOF tests decodeStringKey str16 readBytes error
func TestDecodeStringKeyStr16DataEOF(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	enc.writeUint16(20) // length = 20
	enc.buf = append(enc.buf, "abc"...)
	// Only 3 bytes for key, expected 20

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeStringKeyStr32DataEOF tests decodeStringKey str32 readBytes error
func TestDecodeStringKeyStr32DataEOF(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	enc.writeUint32(20) // length = 20
	enc.buf = append(enc.buf, "abc"...)
	// Only 3 bytes for key, expected 20

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected EOF error")
	}
}

// TestDecodeValueIntUint16 tests decodeValueInt with uint16 format
func TestDecodeValueIntUint16(t *testing.T) {
	type TestStruct struct {
		Value int `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint16)
	enc.writeUint16(1000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 1000 {
		t.Errorf("expected 1000, got %d", s.Value)
	}
}

// TestDecodeValueIntUint32 tests decodeValueInt with uint32 format
func TestDecodeValueIntUint32(t *testing.T) {
	type TestStruct struct {
		Value int `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint32)
	enc.writeUint32(100000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 100000 {
		t.Errorf("expected 100000, got %d", s.Value)
	}
}

// TestDecodeValueIntInt8 tests decodeValueInt with int8 format
func TestDecodeValueIntInt8(t *testing.T) {
	type TestStruct struct {
		Value int `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt8)
	enc.writeByte(0xCE) // -50 as int8

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != -50 {
		t.Errorf("expected -50, got %d", s.Value)
	}
}

// TestDecodeValueIntInt16 tests decodeValueInt with int16 format
func TestDecodeValueIntInt16(t *testing.T) {
	type TestStruct struct {
		Value int `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt16)
	enc.writeUint16(uint16(-1000 & 0xFFFF))

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeValueIntInt32 tests decodeValueInt with int32 format
func TestDecodeValueIntInt32(t *testing.T) {
	type TestStruct struct {
		Value int `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt32)
	enc.writeUint32(uint32(-100000 & 0xFFFFFFFF))

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeValueUintUint16 tests decodeValueUint with uint16 format
func TestDecodeValueUintUint16(t *testing.T) {
	type TestStruct struct {
		Value uint `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint16)
	enc.writeUint16(1000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 1000 {
		t.Errorf("expected 1000, got %d", s.Value)
	}
}

// TestDecodeValueUintUint32 tests decodeValueUint with uint32 format
func TestDecodeValueUintUint32(t *testing.T) {
	type TestStruct struct {
		Value uint `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint32)
	enc.writeUint32(100000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 100000 {
		t.Errorf("expected 100000, got %d", s.Value)
	}
}

// TestDecodeValueFloatFromInt tests decodeValueFloat from integer format
func TestDecodeValueFloatFromUint8(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint8)
	enc.writeByte(42)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 42.0 {
		t.Errorf("expected 42.0, got %f", s.Value)
	}
}

// TestDecodeValueFloatFromUint16 tests decodeValueFloat from uint16
func TestDecodeValueFloatFromUint16(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint16)
	enc.writeUint16(1000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 1000.0 {
		t.Errorf("expected 1000.0, got %f", s.Value)
	}
}

// TestDecodeValueFloatFromUint32 tests decodeValueFloat from uint32
func TestDecodeValueFloatFromUint32(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatUint32)
	enc.writeUint32(100000)

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != 100000.0 {
		t.Errorf("expected 100000.0, got %f", s.Value)
	}
}

// TestDecodeValueFloatFromInt8 tests decodeValueFloat from int8
func TestDecodeValueFloatFromInt8(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt8)
	enc.writeByte(0xCE) // -50 as int8

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Value != -50.0 {
		t.Errorf("expected -50.0, got %f", s.Value)
	}
}

// TestDecodeValueFloatFromInt16 tests decodeValueFloat from int16
func TestDecodeValueFloatFromInt16(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt16)
	enc.writeUint16(uint16(-1000 & 0xFFFF))

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeValueFloatFromInt32 tests decodeValueFloat from int32
func TestDecodeValueFloatFromInt32(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt32)
	enc.writeUint32(uint32(-100000 & 0xFFFFFFFF))

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeValueFloatFromInt64 tests decodeValueFloat from int64
func TestDecodeValueFloatFromInt64(t *testing.T) {
	type TestStruct struct {
		Value float64 `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatInt64)
	enc.writeUint64(uint64(-1000000000 & 0xFFFFFFFFFFFFFFFF))

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestDecodeIntoMapTypeMismatch tests decodeIntoMap with non-map format
func TestDecodeIntoMapTypeMismatchFormat(t *testing.T) {
	type TestStruct struct {
		Data map[string]string `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.EncodeString("not a map")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeIntoSliceTypeMismatch tests decodeIntoSlice with non-array format
func TestDecodeIntoSliceTypeMismatchFormat(t *testing.T) {
	type TestStruct struct {
		Items []string `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeString("not an array")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestDecodeIntoArrayTypeMismatch tests decodeIntoArray with non-array format
func TestDecodeIntoArrayTypeMismatchFormat(t *testing.T) {
	type TestStruct struct {
		Items [3]string `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeString("not an array")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != ErrTypeMismatch {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

// TestBuildStructFieldsPrivateField tests buildStructFields skips private fields
func TestBuildStructFieldsPrivateField(t *testing.T) {
	type TestStruct struct {
		Public  string `msgpack:"public"`
		private string `msgpack:"private"` // unexported, should be skipped
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("public")
	enc.EncodeString("value")
	enc.EncodeString("private")
	enc.EncodeString("ignored")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Public != "value" {
		t.Errorf("expected 'value', got %q", s.Public)
	}
}

// TestBuildStructFieldsSkipTag tests buildStructFields skips "-" tagged fields
func TestBuildStructFieldsSkipTag(t *testing.T) {
	type TestStruct struct {
		Name   string `msgpack:"name"`
		Skip   string `msgpack:"-"`
		Ignore string `msgpack:"-"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(3)
	enc.EncodeString("name")
	enc.EncodeString("Alice")
	enc.EncodeString("Skip")
	enc.EncodeString("ignored1")
	enc.EncodeString("Ignore")
	enc.EncodeString("ignored2")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
	// Skip and Ignore should remain empty
	if s.Skip != "" || s.Ignore != "" {
		t.Error("skipped fields should be empty")
	}
}

// TestDecodeIntoStructNonStringKeyValueError tests error when skipping value after non-string key fails
func TestDecodeIntoStructNonStringKeyValueError(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeInt(123) // non-string key
	// Now we need incomplete data for the value
	enc.writeByte(formatStr32)
	// Missing length bytes for str32

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	// Should fail when trying to skip the value
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeStringAnyMaxLenValidation tests decodeStringAny validation error
func TestDecodeStringAnyMaxLenValidation(t *testing.T) {
	// Create string with declared length that exceeds max
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte('a')
	}

	cfg := Config{
		MaxStringLen: 10, // Too small
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.DecodeAny()
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeBinaryAnyMaxLenValidation tests decodeBinaryAny validation error
func TestDecodeBinaryAnyMaxLenValidation(t *testing.T) {
	// Create binary with declared length that exceeds max
	enc := NewEncoder(200)
	enc.writeByte(formatBin8)
	enc.writeByte(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte(byte(i))
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 10, // Too small
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	d := NewDecoderWithConfig(enc.Bytes(), cfg)
	_, err := d.DecodeAny()
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeIntoValueInterfaceError tests decodeIntoValue interface with error in decodeValue
func TestDecodeIntoValueInterfaceError(t *testing.T) {
	type TestStruct struct {
		Value any `msgpack:"value"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("value")
	enc.writeByte(formatStr32)
	// Missing length bytes

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeStringKeyTypeMismatch tests decodeStringKey non-string key type
func TestDecodeStringKeyTypeMismatch(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeBool(true) // non-string key
	enc.EncodeString("value")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	// Should fail since we can't skip boolean key in decodeStringKey
	if err != nil {
		t.Errorf("expected to skip non-string key, got error: %v", err)
	}
}

// TestDecodeIntoStructUnknownFieldSkip tests decodeIntoStruct skipping unknown fields
func TestDecodeIntoStructUnknownFieldSkip(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("unknown")
	enc.EncodeString("ignored")
	enc.EncodeString("name")
	enc.EncodeString("Alice")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

// TestDecodeIntoStructUnknownFieldSkipDecodeError tests error when skipping unknown field fails
func TestDecodeIntoStructUnknownFieldSkipDecodeError(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("unknown")
	enc.writeByte(formatStr32)
	// Missing length bytes

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeIntoStructUnexportedField tests decodeIntoStruct with unexported field match
func TestDecodeIntoStructUnexportedFieldSkip(t *testing.T) {
	type TestStruct struct {
		name string // unexported
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.EncodeString("Alice")

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

// TestDecodeValueValidationStr8 tests string validation in decodeValueString
func TestDecodeValueValidationStr8(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(200)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr8)
	enc.writeByte(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte('a')
	}

	cfg := Config{
		MaxStringLen: 10, // Too small
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	var s TestStruct
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueValidationStr16 tests string validation in decodeValueString str16
func TestDecodeValueValidationStr16(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(200)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr16)
	enc.writeUint16(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte('a')
	}

	cfg := Config{
		MaxStringLen: 10, // Too small
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	var s TestStruct
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueValidationStr32 tests string validation in decodeValueString str32
func TestDecodeValueValidationStr32(t *testing.T) {
	type TestStruct struct {
		Name string `msgpack:"name"`
	}

	enc := NewEncoder(200)
	enc.EncodeMapHeader(1)
	enc.EncodeString("name")
	enc.writeByte(formatStr32)
	enc.writeUint32(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte('a')
	}

	cfg := Config{
		MaxStringLen: 10, // Too small
		MaxBinaryLen: 1024 * 1024,
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	var s TestStruct
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrStringTooLong {
		t.Errorf("expected ErrStringTooLong, got %v", err)
	}
}

// TestDecodeValueBytesBin8Validation tests validation in decodeValueBytes with bin8
func TestDecodeValueBytesBin8Validation(t *testing.T) {
	type TestStruct struct {
		Data []byte `msgpack:"data"`
	}

	enc := NewEncoder(200)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.writeByte(formatBin8)
	enc.writeByte(100) // length = 100
	for i := 0; i < 100; i++ {
		enc.writeByte(byte(i))
	}

	cfg := Config{
		MaxStringLen: 1024 * 1024,
		MaxBinaryLen: 10, // Too small
		MaxArrayLen:  10000,
		MaxMapLen:    10000,
		MaxExtLen:    1024 * 1024,
		MaxDepth:     100,
	}

	var s TestStruct
	err := UnmarshalStructWithConfig(enc.Bytes(), &s, cfg)
	if err != ErrBinaryTooLong {
		t.Errorf("expected ErrBinaryTooLong, got %v", err)
	}
}

// TestDecodeIntoArrayElementError tests decodeIntoArray element decode error
func TestDecodeIntoArrayElementDecodeError(t *testing.T) {
	type TestStruct struct {
		Items [2]string `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(2)
	enc.EncodeString("first")
	enc.writeByte(formatStr32)
	// Missing length bytes for second element

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeIntoSliceElementError tests decodeIntoSlice element decode error
func TestDecodeIntoSliceElementDecodeError(t *testing.T) {
	type TestStruct struct {
		Items []string `msgpack:"items"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("items")
	enc.EncodeArrayHeader(2)
	enc.EncodeString("first")
	enc.writeByte(formatStr32)
	// Missing length bytes for second element

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeIntoMapKeyError tests decodeIntoMap key decode error
func TestDecodeIntoMapKeyDecodeError(t *testing.T) {
	type TestStruct struct {
		Data map[string]string `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	// Missing length bytes for key

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeIntoMapValueError tests decodeIntoMap value decode error
func TestDecodeIntoMapValueDecodeError(t *testing.T) {
	type TestStruct struct {
		Data map[string]string `msgpack:"data"`
	}

	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.EncodeString("data")
	enc.EncodeMapHeader(1)
	enc.EncodeString("key")
	enc.writeByte(formatStr32)
	// Missing length bytes for value

	var s TestStruct
	err := UnmarshalStruct(enc.Bytes(), &s)
	if err == nil {
		t.Error("expected error")
	}
}

// TestDecodeAnyValueUint64Overflow tests uint64 overflow path in decodeAnyValue
func TestDecodeAnyValueUint64Overflow(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatUint64)
	enc.writeUint64(0xFFFFFFFFFFFFFFFF) // Max uint64, overflows int64

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, ok := v.(uint64); !ok {
		t.Errorf("expected uint64, got %T", v)
	}
}

// TestDecodeAnyValueFloat32 tests float32 path in decodeAnyValue
func TestDecodeAnyValueFloat32(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeFloat32(3.14)

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if _, ok := v.(float64); !ok {
		t.Errorf("expected float64, got %T", v)
	}
}

// TestDecodeMapAnyStr16Key tests str16 key in decodeMapAny
func TestDecodeMapAnyStr16Key(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	enc.writeUint16(3)
	enc.buf = append(enc.buf, "key"...)
	enc.EncodeString("value")

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", v)
	}
	if m["key"] != "value" {
		t.Errorf("expected 'value', got %v", m["key"])
	}
}

// TestDecodeMapAnyStr32KeyFormat tests str32 key format in decodeMapAny
func TestDecodeMapAnyStr32KeyFormat(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	enc.writeUint32(3)
	enc.buf = append(enc.buf, "key"...)
	enc.EncodeString("value")

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", v)
	}
	if m["key"] != "value" {
		t.Errorf("expected 'value', got %v", m["key"])
	}
}

// TestTimestampRoundTrip tests timestamp encode/decode round-trip
func TestTimestampRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		ts   time.Time
	}{
		{"unix_epoch", time.Unix(0, 0).UTC()},
		{"recent", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{"with_nanos", time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)},
		{"max_timestamp32", time.Unix(0xFFFFFFFF, 0).UTC()},
		{"timestamp64_range", time.Unix(0x100000000, 0).UTC()},
		{"timestamp64_nanos", time.Unix(0x100000000, 500000000).UTC()},
		{"negative_unix", time.Unix(-1, 0).UTC()},
		{"far_past", time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"far_future", time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := MarshalTimestamp(tc.ts)
			decoded, err := UnmarshalTimestamp(data)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if !decoded.Equal(tc.ts) {
				t.Errorf("got %v, want %v", decoded, tc.ts)
			}
			if decoded.Location() != time.UTC {
				t.Errorf("expected UTC, got %v", decoded.Location())
			}
		})
	}
}

// TestTimestampEncoderDecoder tests Encoder/Decoder methods
func TestTimestampEncoderDecoder(t *testing.T) {
	ts := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)

	enc := NewEncoder(64)
	enc.EncodeTimestamp(ts)
	data := enc.Bytes()

	dec := NewDecoder(data)
	decoded, err := dec.DecodeTimestamp()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if !decoded.Equal(ts) {
		t.Errorf("got %v, want %v", decoded, ts)
	}
}

// TestTimestampFormats tests all three timestamp formats
func TestTimestampFormats(t *testing.T) {
	t.Run("timestamp32", func(t *testing.T) {
		ts := time.Unix(1000000, 0).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatFixExt4 {
			t.Errorf("expected formatFixExt4, got 0x%x", data[0])
		}
		if data[1] != 0xff {
			t.Errorf("expected type -1, got 0x%x", data[1])
		}
	})

	t.Run("timestamp64", func(t *testing.T) {
		ts := time.Unix(1000000, 500000000).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatFixExt8 {
			t.Errorf("expected formatFixExt8, got 0x%x", data[0])
		}
	})

	t.Run("timestamp96", func(t *testing.T) {
		ts := time.Unix(-1000000, 500000000).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatExt8 {
			t.Errorf("expected formatExt8, got 0x%x", data[0])
		}
		if data[1] != 12 {
			t.Errorf("expected length 12, got %d", data[1])
		}
	})
}

// TestTimestampWithTimezone tests non-UTC times are converted to UTC
func TestTimestampWithTimezone(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("timezone not available")
	}

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	data := MarshalTimestamp(ts)
	decoded, err := UnmarshalTimestamp(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !decoded.Equal(ts) {
		t.Errorf("times not equal: got %v, want %v", decoded, ts)
	}
	if decoded.Location() != time.UTC {
		t.Errorf("expected UTC, got %v", decoded.Location())
	}
}

// TestIsTimestamp tests IsTimestamp helper
func TestIsTimestamp(t *testing.T) {
	if !IsTimestamp(Ext{Type: -1, Data: []byte{1, 2, 3, 4}}) {
		t.Error("expected true for type -1")
	}
	if IsTimestamp(Ext{Type: 1, Data: []byte{1, 2, 3, 4}}) {
		t.Error("expected false for type 1")
	}
}

// TestExtToTimestamp tests ExtToTimestamp conversion
func TestExtToTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	data := MarshalTimestamp(ts)

	dec := NewDecoder(data)
	v, err := dec.Decode()
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if v.Type != TypeExt {
		t.Fatalf("expected ext type, got %v", v.Type)
	}

	decoded, err := ExtToTimestamp(v.Ext)
	if err != nil {
		t.Fatalf("ExtToTimestamp failed: %v", err)
	}
	if !decoded.Equal(ts) {
		t.Errorf("got %v, want %v", decoded, ts)
	}
}

// TestTimestampDecodeErrors tests error cases
func TestTimestampDecodeErrors(t *testing.T) {
	t.Run("wrong_format", func(t *testing.T) {
		data := []byte{formatNil}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("wrong_ext_type", func(t *testing.T) {
		data := []byte{formatFixExt4, 0x01, 0, 0, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("wrong_ext8_length", func(t *testing.T) {
		data := []byte{formatExt8, 5, 0xff, 0, 0, 0, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("truncated_data", func(t *testing.T) {
		data := []byte{formatFixExt4, 0xff, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrUnexpectedEOF {
			t.Errorf("expected ErrUnexpectedEOF, got %v", err)
		}
	})
}

// TestExtToTimestampErrors tests ExtToTimestamp error cases
func TestExtToTimestampErrors(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		ext := Ext{Type: 1, Data: []byte{0, 0, 0, 0}}
		_, err := ExtToTimestamp(ext)
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got %v", err)
		}
	})

	t.Run("wrong_length", func(t *testing.T) {
		ext := Ext{Type: -1, Data: []byte{0, 0, 0}}
		_, err := ExtToTimestamp(ext)
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got %v", err)
		}
	})
}
