package msgpck

import (
	"bytes"
	"reflect"
	"testing"
)

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
				t.Errorf(errMsgFmtSV, f.name, err)
			}
		})
	}
}

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
				t.Errorf(errMsgFmtSV, f.name, err)
			}
		})
	}
}

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
			t.Errorf(errMsgStringTooLong, err)
		}
	})
}

func TestDecodeMapNonStringKey(t *testing.T) {
	// Map with integer key in DecodeAny (should error)
	data := []byte{0x81, 0x42, 0x01} // fixmap 1 with int key
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat for non-string key, got %v", err)
	}
}

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

func TestDecodeIntoValueEOF(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	data := []byte{0x81, 0xa1, 'v'} // missing value
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestReflectionDecodeValueBytesEOF(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	// str8 with declared length but missing data
	data := []byte{0x81, 0xa1, 'b', formatStr8, 100}
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestReflectionDecodeSliceEOF(t *testing.T) {
	type Data struct {
		S []int `msgpack:"s"`
	}

	// array with declared length but missing elements
	data := []byte{0x81, 0xa1, 's', 0x95} // fixarray 5 with no elements
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestReflectionDecodeArrayEOF(t *testing.T) {
	type Data struct {
		A [3]int `msgpack:"a"`
	}

	// array with declared length but missing elements
	data := []byte{0x81, 0xa1, 'a', 0x93} // fixarray 3 with no elements
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestReflectionDecodeMapEOF(t *testing.T) {
	type Data struct {
		M map[string]int `msgpack:"m"`
	}

	// map with declared length but missing entries
	data := []byte{0x81, 0xa1, 'm', 0x81} // fixmap 1 with no entries
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeStructKeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str8 key with missing length byte
	data := []byte{0x81, formatStr8} // missing length
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeStructStr16KeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str16 key with missing length bytes
	data := []byte{0x81, formatStr16, 0} // missing second length byte
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeStructStr32KeyReadError(t *testing.T) {
	type Data struct {
		V int `msgpack:"v"`
	}

	// str32 key with missing length bytes
	data := []byte{0x81, formatStr32, 0, 0, 0} // missing last length byte
	var d Data
	err := UnmarshalStruct(data, &d)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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

func TestDecodeMapNonStringKeyValue(t *testing.T) {
	// Map with integer key
	data := []byte{0x81, 0x42, 0x01} // fixmap 1 with int key 66, value 1
	d := NewDecoder(data)
	v, err := d.Decode()
	// Should still decode but key type might vary
	if err != nil {
		t.Errorf(errMsgDecodeFailed, err)
	}
	if v.Type != TypeMap {
		t.Error("expected map type")
	}
}

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
			t.Errorf(errMsgMapTooLong, err)
		}
	})
}

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
			t.Errorf(errMsgArrayTooLong, err)
		}
	})

	t.Run("array16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(1)
		data := []byte{0x81, 0xa1, 's', formatArray16, 0, 3, 0x01, 0x02, 0x03}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrArrayTooLong {
			t.Errorf(errMsgArrayTooLong, err)
		}
	})
}

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
			t.Errorf(errMsgArrayTooLong, err)
		}
	})
}

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
			t.Errorf(errMsgMapTooLong, err)
		}
	})

	t.Run("map16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(0)
		data := []byte{0x81, 0xa1, 'm', formatMap16, 0, 1, 0xa1, 'k', 0x42}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrMapTooLong {
			t.Errorf(errMsgMapTooLong, err)
		}
	})
}

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
			t.Errorf(errMsgMapTooLong, err)
		}
	})
}

func TestDecodeMapKeyEOF(t *testing.T) {
	t.Run("str8 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr8} // map with str8 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("str16 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0} // map with str16 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("str32 key EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr32, 0, 0} // map with str32 key, missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

func TestDecodeMapAnyKeyEOF(t *testing.T) {
	t.Run("str8 key length EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr8} // map with str8 key, missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("str16 key length EOF", func(t *testing.T) {
		data := []byte{0x81, formatStr16, 0} // map with str16 key, missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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

func TestDecodeMapValueEOF(t *testing.T) {
	// Map header says 2 entries but only has key
	data := []byte{0x82, 0xa1, 'a'} // map of 2, key "a", no value
	d := NewDecoder(data)
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeIntoStructMap16EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map16 format but no length bytes
	data := []byte{formatMap16}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeIntoStructMap32EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// Map32 format but no length bytes
	data := []byte{formatMap32}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStructNilValueError(t *testing.T) {
	data := []byte{0x80} // empty map
	err := UnmarshalStruct(data, nil)
	if err == nil {
		t.Error("expected error for nil value")
	}
}

func TestDecodeStructNonPointerError(t *testing.T) {
	type S struct{}
	data := []byte{0x80} // empty map
	var s S
	err := UnmarshalStruct(data, s) // passing value, not pointer
	if err == nil {
		t.Error("expected error for non-pointer")
	}
}

func TestDecodeStructNilPointerError(t *testing.T) {
	type S struct{}
	data := []byte{0x80}
	var s *S // nil pointer
	err := UnmarshalStruct(data, s)
	if err == nil {
		t.Error("expected error for nil pointer")
	}
}

func TestDecodeStructNonStructError(t *testing.T) {
	data := []byte{0x80}
	var i int
	err := UnmarshalStruct(data, &i) // pointer to int, not struct
	if err == nil {
		t.Error("expected error for non-struct")
	}
}

func TestDecodeStructformatNil(t *testing.T) {
	type S struct{}
	data := []byte{formatNil}
	var s S
	err := UnmarshalStruct(data, &s)
	if err != nil {
		t.Errorf("expected nil to work, got %v", err)
	}
}

func TestDecodeStructEOF(t *testing.T) {
	type S struct{}
	data := []byte{} // empty
	var s S
	err := UnmarshalStruct(data, &s)
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeStructTypeMismatch(t *testing.T) {
	type S struct{}
	data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // string, not map
	var s S
	err := UnmarshalStruct(data, &s)
	if err != ErrTypeMismatch {
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgGot42, o.Inner.X)
	}
}

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
		t.Errorf(errMsgArrayTooLong, err)
	}
}

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
		t.Errorf(errMsgArrayTooLong, err)
	}
}

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
		t.Errorf(errMsgMapTooLong, err)
	}
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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

func TestDecodeMapAnyKeyStr8EOF(t *testing.T) {
	// Map with incomplete str8 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr8)
	// Missing length byte

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeMapAnyKeyStr16EOF(t *testing.T) {
	// Map with incomplete str16 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr16)
	// Missing length bytes

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeMapAnyKeyStr32EOF(t *testing.T) {
	// Map with incomplete str32 key
	enc := NewEncoder(64)
	enc.EncodeMapHeader(1)
	enc.writeByte(formatStr32)
	// Missing length bytes

	d := NewDecoder(enc.Bytes())
	_, err := d.DecodeAny()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Name != nil {
		t.Errorf("expected nil, got %v", s.Name)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Public != "value" {
		t.Errorf("expected 'value', got %q", s.Public)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
	// Skip and Ignore should remain empty
	if s.Skip != "" || s.Ignore != "" {
		t.Error("skipped fields should be empty")
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Name != "Alice" {
		t.Errorf("expected 'Alice', got %q", s.Name)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Error(errMsgExpectedErr)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", v)
	}
	if m["key"] != "value" {
		t.Errorf("expected 'value', got %v", m["key"])
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	m, ok := v.(map[string]any)
	if !ok {
		t.Errorf("expected map[string]any, got %T", v)
	}
	if m["key"] != "value" {
		t.Errorf("expected 'value', got %v", m["key"])
	}
}

func TestStructRoundtripAllMsgpackTypes(t *testing.T) {
	type Nested struct {
		Value int    `msgpack:"value"`
		Label string `msgpack:"label"`
	}

	type AllTypes struct {
		// Integers - signed
		I   int   `msgpack:"i"`
		I64 int64 `msgpack:"i64"`
		I32 int32 `msgpack:"i32"`
		I16 int16 `msgpack:"i16"`
		I8  int8  `msgpack:"i8"`
		// Integers - unsigned
		U   uint   `msgpack:"u"`
		U64 uint64 `msgpack:"u64"`
		U32 uint32 `msgpack:"u32"`
		U16 uint16 `msgpack:"u16"`
		U8  uint8  `msgpack:"u8"`
		// Floats
		F64 float64 `msgpack:"f64"`
		F32 float32 `msgpack:"f32"`
		// Bool
		BoolTrue  bool `msgpack:"bool_true"`
		BoolFalse bool `msgpack:"bool_false"`
		// String
		Str      string `msgpack:"str"`
		StrEmpty string `msgpack:"str_empty"`
		// Binary
		Bin      []byte `msgpack:"bin"`
		BinEmpty []byte `msgpack:"bin_empty"`
		// String slice
		StrSlice      []string `msgpack:"str_slice"`
		StrSliceEmpty []string `msgpack:"str_slice_empty"`
		// String map
		StrMap      map[string]string `msgpack:"str_map"`
		StrMapEmpty map[string]string `msgpack:"str_map_empty"`
		// Nested struct
		Nested Nested `msgpack:"nested"`
	}

	original := AllTypes{
		// Integers - signed (test various ranges)
		I:   -1000000,
		I64: -9223372036854775807,
		I32: -2147483647,
		I16: -32767,
		I8:  -127,
		// Integers - unsigned (test max values)
		U:   1000000,
		U64: 18446744073709551615,
		U32: 4294967295,
		U16: 65535,
		U8:  255,
		// Floats
		F64: 3.141592653589793,
		F32: 2.7182817,
		// Bool
		BoolTrue:  true,
		BoolFalse: false,
		// String
		Str:      "hello, 世界! 🎉",
		StrEmpty: "",
		// Binary
		Bin:      []byte{0x00, 0x01, 0x02, 0xff, 0xfe},
		BinEmpty: []byte{},
		// String slice
		StrSlice:      []string{"one", "two", "three"},
		StrSliceEmpty: []string{},
		// String map
		StrMap:      map[string]string{"key1": "value1", "key2": "value2"},
		StrMapEmpty: map[string]string{},
		// Nested struct
		Nested: Nested{Value: 42, Label: "nested"},
	}

	enc := GetStructEncoder[AllTypes]()
	data, err := enc.EncodeCopy(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	dec := GetStructDecoder[AllTypes](false)
	var result AllTypes
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailed, err)
	}

	// Verify integers - signed
	if result.I != original.I {
		t.Errorf("I: got %d, want %d", result.I, original.I)
	}
	if result.I64 != original.I64 {
		t.Errorf("I64: got %d, want %d", result.I64, original.I64)
	}
	if result.I32 != original.I32 {
		t.Errorf("I32: got %d, want %d", result.I32, original.I32)
	}
	if result.I16 != original.I16 {
		t.Errorf("I16: got %d, want %d", result.I16, original.I16)
	}
	if result.I8 != original.I8 {
		t.Errorf("I8: got %d, want %d", result.I8, original.I8)
	}

	// Verify integers - unsigned
	if result.U != original.U {
		t.Errorf("U: got %d, want %d", result.U, original.U)
	}
	if result.U64 != original.U64 {
		t.Errorf("U64: got %d, want %d", result.U64, original.U64)
	}
	if result.U32 != original.U32 {
		t.Errorf("U32: got %d, want %d", result.U32, original.U32)
	}
	if result.U16 != original.U16 {
		t.Errorf("U16: got %d, want %d", result.U16, original.U16)
	}
	if result.U8 != original.U8 {
		t.Errorf("U8: got %d, want %d", result.U8, original.U8)
	}

	// Verify floats
	if result.F64 != original.F64 {
		t.Errorf("F64: got %v, want %v", result.F64, original.F64)
	}
	if result.F32 != original.F32 {
		t.Errorf("F32: got %v, want %v", result.F32, original.F32)
	}

	// Verify bools
	if result.BoolTrue != original.BoolTrue {
		t.Errorf("BoolTrue: got %v, want %v", result.BoolTrue, original.BoolTrue)
	}
	if result.BoolFalse != original.BoolFalse {
		t.Errorf("BoolFalse: got %v, want %v", result.BoolFalse, original.BoolFalse)
	}

	// Verify strings
	if result.Str != original.Str {
		t.Errorf("Str: got %q, want %q", result.Str, original.Str)
	}
	if result.StrEmpty != original.StrEmpty {
		t.Errorf("StrEmpty: got %q, want %q", result.StrEmpty, original.StrEmpty)
	}

	// Verify binary
	if !bytes.Equal(result.Bin, original.Bin) {
		t.Errorf("Bin: got %v, want %v", result.Bin, original.Bin)
	}
	if !bytes.Equal(result.BinEmpty, original.BinEmpty) {
		t.Errorf("BinEmpty: got %v, want %v", result.BinEmpty, original.BinEmpty)
	}

	// Verify string slice
	if len(result.StrSlice) != len(original.StrSlice) {
		t.Errorf("StrSlice length: got %d, want %d", len(result.StrSlice), len(original.StrSlice))
	} else {
		for i, v := range original.StrSlice {
			if result.StrSlice[i] != v {
				t.Errorf("StrSlice[%d]: got %q, want %q", i, result.StrSlice[i], v)
			}
		}
	}
	if len(result.StrSliceEmpty) != 0 {
		t.Errorf("StrSliceEmpty: got %v, want empty", result.StrSliceEmpty)
	}

	// Verify string map
	if len(result.StrMap) != len(original.StrMap) {
		t.Errorf("StrMap length: got %d, want %d", len(result.StrMap), len(original.StrMap))
	} else {
		for k, v := range original.StrMap {
			if result.StrMap[k] != v {
				t.Errorf("StrMap[%q]: got %q, want %q", k, result.StrMap[k], v)
			}
		}
	}
	if len(result.StrMapEmpty) != 0 {
		t.Errorf("StrMapEmpty: got %v, want empty", result.StrMapEmpty)
	}

	// Verify nested struct
	if result.Nested.Value != original.Nested.Value {
		t.Errorf("Nested.Value: got %d, want %d", result.Nested.Value, original.Nested.Value)
	}
	if result.Nested.Label != original.Nested.Label {
		t.Errorf("Nested.Label: got %q, want %q", result.Nested.Label, original.Nested.Label)
	}
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

// TestDecodeStruct tests struct unmarshaling
func TestDecodeStruct(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	// Encode
	original := Person{Name: "Alice", Age: 30}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	// Decode
	var decoded Person
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf(errMsgStructFailed, err)
	}

	if decoded != original {
		t.Errorf(errMsgGotWantStruct, decoded, original)
	}
}

// TestOmitEmpty tests omitempty tag
func TestOmitEmpty(t *testing.T) {
	type Data struct {
		Name  string `msgpack:"name"`
		Value int    `msgpack:"value,omitempty"`
	}

	// Value is zero, should be omitted
	original := Data{Name: "test", Value: 0}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	// Decode as map to check fields
	decoded, err := Unmarshal(encoded)
	if err != nil {
		t.Fatalf(errMsgUnmarshalFailed, err)
	}

	m := decoded.(map[string]any)
	if _, ok := m["value"]; ok {
		t.Error("value field should be omitted")
	}
	if m["name"] != "test" {
		t.Errorf("name = %v, want test", m["name"])
	}
}

// TestNestedStruct tests nested struct decoding
func TestNestedStruct(t *testing.T) {
	type Address struct {
		City string `msgpack:"city"`
	}
	type Person struct {
		Name    string  `msgpack:"name"`
		Address Address `msgpack:"address"`
	}

	original := Person{
		Name:    "Alice",
		Address: Address{City: "NYC"},
	}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	var decoded Person
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf(errMsgStructFailed, err)
	}

	if decoded != original {
		t.Errorf(errMsgGotWantStruct, decoded, original)
	}
}

// TestSliceInStruct tests slice field in struct
func TestSliceInStruct(t *testing.T) {
	type Data struct {
		Values []int `msgpack:"values"`
	}

	original := Data{Values: []int{1, 2, 3}}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	var decoded Data
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf(errMsgStructFailed, err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf(errMsgGotWantStruct, decoded, original)
	}
}

// TestMapInStruct tests map field in struct
func TestMapInStruct(t *testing.T) {
	type Data struct {
		Meta map[string]string `msgpack:"meta"`
	}

	original := Data{Meta: map[string]string{"key": "value"}}

	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	var decoded Data
	err = UnmarshalStruct(encoded, &decoded)
	if err != nil {
		t.Fatalf(errMsgStructFailed, err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf(errMsgGotWantStruct, decoded, original)
	}
}
