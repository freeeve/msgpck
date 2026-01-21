package msgpck

import (
	"testing"
)

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
				t.Errorf(errMsgFmtSV, f.name, err)
			}
		})
	}
}

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
				t.Errorf(errMsgFmtSV, tc.name, err)
			}
			if d.I != tc.want {
				t.Errorf("%s: got %d, want %d", tc.name, d.I, tc.want)
			}
		})
	}
}

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
				t.Errorf(errMsgFmtSV, tc.name, err)
			}
		})
	}
}

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

func TestStructDecoderAllIntegerTypes(t *testing.T) {
	type AllInts struct {
		I   int    `msgpack:"i"`
		I64 int64  `msgpack:"i64"`
		I32 int32  `msgpack:"i32"`
		I16 int16  `msgpack:"i16"`
		I8  int8   `msgpack:"i8"`
		U   uint   `msgpack:"u"`
		U64 uint64 `msgpack:"u64"`
		U32 uint32 `msgpack:"u32"`
		U16 uint16 `msgpack:"u16"`
		U8  uint8  `msgpack:"u8"`
	}

	original := AllInts{
		I:   -1000000,
		I64: -9223372036854775807,
		I32: -2147483647,
		I16: -32767,
		I8:  -127,
		U:   1000000,
		U64: 18446744073709551615,
		U32: 4294967295,
		U16: 65535,
		U8:  255,
	}

	enc := GetStructEncoder[AllInts]()
	data, err := enc.EncodeCopy(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	dec := GetStructDecoder[AllInts](false)
	var result AllInts
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailed, err)
	}

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
