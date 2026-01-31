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
	b, err := enc.Encode(&original)
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
		{testPositiveFixint, []byte{0x82, 0xa1, 'u', 0x42, 0xa3, 'u', '6', '4', 0x42}},
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
		{testPositiveFixint, []byte{0x81, 0xa1, 'i', 0x42}, 66},
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
		{testPositiveFixint, []byte{0x81, 0xa1, 'u', 0x42}},
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
	data, err := enc.Encode(&original)
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
		var m any
		_ = Unmarshal(b, &m)
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
}

// TestStructDecoderDecodeWith tests the DecodeWith method for decoder reuse
func TestStructDecoderDecodeWith(t *testing.T) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	enc := GetStructEncoder[Person]()
	dec := GetStructDecoder[Person](false)

	// Create test data
	original := Person{Name: "Alice", Age: 30}
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Use DecodeWith with a user-managed decoder
	d := NewDecoder(nil)
	var result Person
	err = dec.DecodeWith(d, data, &result)
	if err != nil {
		t.Fatalf("DecodeWith failed: %v", err)
	}

	if result.Name != original.Name || result.Age != original.Age {
		t.Errorf("got %+v, want %+v", result, original)
	}

	// Verify decoder can be reused
	data2, _ := enc.Encode(&Person{Name: "Bob", Age: 25})
	var result2 Person
	err = dec.DecodeWith(d, data2, &result2)
	if err != nil {
		t.Fatalf("second DecodeWith failed: %v", err)
	}
	if result2.Name != "Bob" {
		t.Errorf("got name %q, want Bob", result2.Name)
	}
}

// Test nested maps: map[string]map[string][]byte
func TestStructDecoderNestedMaps(t *testing.T) {
	type BitmapFilterData struct {
		Fields map[string]map[string][]byte `msgpack:"fields"`
	}

	original := BitmapFilterData{
		Fields: map[string]map[string][]byte{
			"category": {
				"electronics": []byte{1, 2, 3},
				"books":       []byte{4, 5, 6},
			},
			"status": {
				"active": []byte{7, 8},
			},
		},
	}

	// First test with regular Marshal/Unmarshal
	t.Run("regular Marshal/Unmarshal", func(t *testing.T) {
		data, err := Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result BitmapFilterData
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Verify nested map contents
		if len(result.Fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(result.Fields))
		}
		if cat, ok := result.Fields["category"]; !ok || len(cat) != 2 {
			t.Errorf("category field missing or wrong size")
		}
		if string(result.Fields["category"]["electronics"]) != string([]byte{1, 2, 3}) {
			t.Errorf("electronics data mismatch")
		}
	})

	// Test with struct encoder/decoder
	t.Run("struct encoder/decoder", func(t *testing.T) {
		enc := GetStructEncoder[BitmapFilterData]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[BitmapFilterData](false)
		var result BitmapFilterData
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		// Verify nested map contents
		if len(result.Fields) != 2 {
			t.Errorf("expected 2 fields, got %d", len(result.Fields))
		}
		if cat, ok := result.Fields["category"]; !ok || len(cat) != 2 {
			t.Errorf("category field missing or wrong size: %v", result.Fields)
		}
	})
}

// Test generic struct with type parameter - comprehensive tests
func TestStructDecoderGenericTypeParam(t *testing.T) {
	type SortColumnData[T any] struct {
		Values   []T    `msgpack:"values"`
		MaxDocID uint32 `msgpack:"max_doc_id"`
	}

	t.Run("int64 values", func(t *testing.T) {
		original := SortColumnData[int64]{
			Values:   []int64{100, 200, 300, 400, 500},
			MaxDocID: 1000,
		}

		// Test with regular Marshal/Unmarshal
		data, err := Marshal(original)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var result SortColumnData[int64]
		err = Unmarshal(data, &result)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if result.MaxDocID != 1000 {
			t.Errorf(errMsgMaxDocID1000, result.MaxDocID)
		}
		if len(result.Values) != 5 || result.Values[4] != 500 {
			t.Errorf(errMsgValuesMismatch, result.Values)
		}
	})

	t.Run("int64 struct encoder/decoder", func(t *testing.T) {
		original := SortColumnData[int64]{
			Values:   []int64{100, 200, 300, 400, 500},
			MaxDocID: 1000,
		}

		enc := GetStructEncoder[SortColumnData[int64]]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[SortColumnData[int64]](false)
		var result SortColumnData[int64]
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 1000 {
			t.Errorf(errMsgMaxDocID1000, result.MaxDocID)
		}
		if len(result.Values) != 5 || result.Values[4] != 500 {
			t.Errorf(errMsgValuesMismatch, result.Values)
		}
	})

	t.Run("float64 struct encoder/decoder", func(t *testing.T) {
		original := SortColumnData[float64]{
			Values:   []float64{1.1, 2.2, 3.3},
			MaxDocID: 500,
		}

		enc := GetStructEncoder[SortColumnData[float64]]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[SortColumnData[float64]](false)
		var result SortColumnData[float64]
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 500 {
			t.Errorf(errMsgMaxDocID500, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
	})

	t.Run("string struct encoder/decoder", func(t *testing.T) {
		original := SortColumnData[string]{
			Values:   []string{"apple", "banana", "cherry"},
			MaxDocID: 100,
		}

		enc := GetStructEncoder[SortColumnData[string]]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[SortColumnData[string]](false)
		var result SortColumnData[string]
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 100 {
			t.Errorf("MaxDocID: got %d, want 100", result.MaxDocID)
		}
		if len(result.Values) != 3 || result.Values[0] != "apple" {
			t.Errorf(errMsgValuesMismatch, result.Values)
		}
	})
}

// Test the exact pattern from roaringsearch sortColumnData
func TestStructDecoderSortColumnDataPattern(t *testing.T) {
	// Simulating cmp.Ordered constraint with concrete types
	type sortColumnData[T int64 | float64 | string] struct {
		Values   []T    `msgpack:"values"`
		MaxDocID uint32 `msgpack:"max_doc_id"`
	}

	t.Run("int64 roundtrip", func(t *testing.T) {
		original := sortColumnData[int64]{
			Values:   []int64{500, 200, 1000},
			MaxDocID: 1000,
		}

		enc := GetStructEncoder[sortColumnData[int64]]()
		dec := GetStructDecoder[sortColumnData[int64]](false)

		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		var result sortColumnData[int64]
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 1000 {
			t.Errorf(errMsgMaxDocID1000, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
		if len(result.Values) > 0 && result.Values[0] != 500 {
			t.Errorf("Values[0]: got %d, want 500", result.Values[0])
		}
	})

	t.Run("float64 roundtrip", func(t *testing.T) {
		original := sortColumnData[float64]{
			Values:   []float64{1.5, 2.5, 3.5},
			MaxDocID: 500,
		}

		enc := GetStructEncoder[sortColumnData[float64]]()
		dec := GetStructDecoder[sortColumnData[float64]](false)

		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		var result sortColumnData[float64]
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 500 {
			t.Errorf(errMsgMaxDocID500, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
	})
}

// Test calling struct decoder from within a generic function (like roaringsearch does)
// Uses package-level generic functions since Go doesn't allow generic function literals

func TestStructDecoderGenericFunctionContext(t *testing.T) {
	t.Run("int64 via generic function", func(t *testing.T) {
		original := testSortColumnData[int64]{
			Values:   []int64{500, 200, 1000},
			MaxDocID: 1000,
		}

		data, err := encodeTestSortColumn(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		result, err := decodeTestSortColumn[int64](data)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 1000 {
			t.Errorf(errMsgMaxDocID1000, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
		if len(result.Values) > 0 && result.Values[0] != 500 {
			t.Errorf("Values[0]: got %d, want 500", result.Values[0])
		}
		if len(result.Values) > 1 && result.Values[1] != 200 {
			t.Errorf("Values[1]: got %d, want 200", result.Values[1])
		}
	})

	t.Run("float64 via generic function", func(t *testing.T) {
		original := testSortColumnData[float64]{
			Values:   []float64{1.5, 2.5, 3.5},
			MaxDocID: 500,
		}

		data, err := encodeTestSortColumn(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		result, err := decodeTestSortColumn[float64](data)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 500 {
			t.Errorf(errMsgMaxDocID500, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
	})

	t.Run("uint16 via generic function (roaringsearch pattern)", func(t *testing.T) {
		original := testSortColumnData[uint16]{
			Values:   []uint16{100, 200, 500},
			MaxDocID: 1000,
		}

		data, err := encodeTestSortColumn(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		result, err := decodeTestSortColumn[uint16](data)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.MaxDocID != 1000 {
			t.Errorf(errMsgMaxDocID1000, result.MaxDocID)
		}
		if len(result.Values) != 3 {
			t.Errorf(errMsgValuesLength3, len(result.Values))
		}
		if len(result.Values) > 0 && result.Values[0] != 100 {
			t.Errorf("Values[0]: got %d, want 100", result.Values[0])
		}
		if len(result.Values) > 2 && result.Values[2] != 500 {
			t.Errorf("Values[2]: got %d, want 500", result.Values[2])
		}
	})
}

// Helper types for generic function context tests
type testSortColumnData[T any] struct {
	Values   []T    `msgpack:"values"`
	MaxDocID uint32 `msgpack:"max_doc_id"`
}

func decodeTestSortColumn[T any](data []byte) (testSortColumnData[T], error) {
	dec := GetStructDecoder[testSortColumnData[T]](false)
	var result testSortColumnData[T]
	err := dec.Decode(data, &result)
	return result, err
}

func encodeTestSortColumn[T any](s *testSortColumnData[T]) ([]byte, error) {
	enc := GetStructEncoder[testSortColumnData[T]]()
	return enc.Encode(s)
}

// TestStructDecoderPointerFields tests decoding of struct fields that are pointers.
func TestStructDecoderPointerFields(t *testing.T) {
	type PersonWithPointers struct {
		Name    string   `msgpack:"name"`
		Address *string  `msgpack:"address"`
		City    *string  `msgpack:"city"`
		Age     *int     `msgpack:"age"`
		Score   *float64 `msgpack:"score"`
		Active  *bool    `msgpack:"active"`
	}

	t.Run("non-nil pointer fields", func(t *testing.T) {
		addr := "123 Main St"
		city := "Boston"
		age := 30
		score := 95.5
		active := true

		original := PersonWithPointers{
			Name:    "Alice",
			Address: &addr,
			City:    &city,
			Age:     &age,
			Score:   &score,
			Active:  &active,
		}

		enc := GetStructEncoder[PersonWithPointers]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[PersonWithPointers](false)
		var result PersonWithPointers
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.Name != "Alice" {
			t.Errorf("Name: got %q, want Alice", result.Name)
		}
		if result.Address == nil {
			t.Error("Address: got nil, want pointer")
		} else if *result.Address != "123 Main St" {
			t.Errorf("Address: got %q, want 123 Main St", *result.Address)
		}
		if result.City == nil {
			t.Error("City: got nil, want pointer")
		} else if *result.City != "Boston" {
			t.Errorf("City: got %q, want Boston", *result.City)
		}
		if result.Age == nil {
			t.Error("Age: got nil, want pointer")
		} else if *result.Age != 30 {
			t.Errorf("Age: got %d, want 30", *result.Age)
		}
		if result.Score == nil {
			t.Error("Score: got nil, want pointer")
		} else if *result.Score != 95.5 {
			t.Errorf("Score: got %f, want 95.5", *result.Score)
		}
		if result.Active == nil {
			t.Error("Active: got nil, want pointer")
		} else if !*result.Active {
			t.Errorf("Active: got false, want true")
		}
	})

	t.Run("nil pointer fields", func(t *testing.T) {
		original := PersonWithPointers{
			Name: "Bob",
			// All pointer fields are nil
		}

		enc := GetStructEncoder[PersonWithPointers]()
		data, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf(errMsgEncodeFailed, err)
		}

		dec := GetStructDecoder[PersonWithPointers](false)
		var result PersonWithPointers
		err = dec.Decode(data, &result)
		if err != nil {
			t.Fatalf(errMsgDecodeFailedCap, err)
		}

		if result.Name != "Bob" {
			t.Errorf("Name: got %q, want Bob", result.Name)
		}
		if result.Address != nil {
			t.Errorf("Address: got %v, want nil", result.Address)
		}
		if result.City != nil {
			t.Errorf("City: got %v, want nil", result.City)
		}
		if result.Age != nil {
			t.Errorf("Age: got %v, want nil", result.Age)
		}
		if result.Score != nil {
			t.Errorf("Score: got %v, want nil", result.Score)
		}
		if result.Active != nil {
			t.Errorf("Active: got %v, want nil", result.Active)
		}
	})
}

// TestStructDecoderAllPointerTypes tests all supported pointer types.
func TestStructDecoderAllPointerTypes(t *testing.T) {
	type AllPointers struct {
		PtrString  *string  `msgpack:"ptr_string"`
		PtrInt     *int     `msgpack:"ptr_int"`
		PtrInt64   *int64   `msgpack:"ptr_int64"`
		PtrInt32   *int32   `msgpack:"ptr_int32"`
		PtrInt16   *int16   `msgpack:"ptr_int16"`
		PtrInt8    *int8    `msgpack:"ptr_int8"`
		PtrUint    *uint    `msgpack:"ptr_uint"`
		PtrUint64  *uint64  `msgpack:"ptr_uint64"`
		PtrUint32  *uint32  `msgpack:"ptr_uint32"`
		PtrUint16  *uint16  `msgpack:"ptr_uint16"`
		PtrUint8   *uint8   `msgpack:"ptr_uint8"`
		PtrFloat64 *float64 `msgpack:"ptr_float64"`
		PtrFloat32 *float32 `msgpack:"ptr_float32"`
		PtrBool    *bool    `msgpack:"ptr_bool"`
	}

	// Create values for all pointer fields
	s := "test"
	i := -42
	i64 := int64(-1000000)
	i32 := int32(-50000)
	i16 := int16(-1000)
	i8 := int8(-100)
	u := uint(42)
	u64 := uint64(1000000)
	u32 := uint32(50000)
	u16 := uint16(1000)
	u8 := uint8(200)
	f64 := 3.14159
	f32 := float32(2.71828)
	b := true

	original := AllPointers{
		PtrString:  &s,
		PtrInt:     &i,
		PtrInt64:   &i64,
		PtrInt32:   &i32,
		PtrInt16:   &i16,
		PtrInt8:    &i8,
		PtrUint:    &u,
		PtrUint64:  &u64,
		PtrUint32:  &u32,
		PtrUint16:  &u16,
		PtrUint8:   &u8,
		PtrFloat64: &f64,
		PtrFloat32: &f32,
		PtrBool:    &b,
	}

	enc := GetStructEncoder[AllPointers]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[AllPointers](false)
	var result AllPointers
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	// Verify all pointer fields
	if result.PtrString == nil || *result.PtrString != s {
		t.Errorf("PtrString: got %v, want %q", result.PtrString, s)
	}
	if result.PtrInt == nil || *result.PtrInt != i {
		t.Errorf("PtrInt: got %v, want %d", result.PtrInt, i)
	}
	if result.PtrInt64 == nil || *result.PtrInt64 != i64 {
		t.Errorf("PtrInt64: got %v, want %d", result.PtrInt64, i64)
	}
	if result.PtrInt32 == nil || *result.PtrInt32 != i32 {
		t.Errorf("PtrInt32: got %v, want %d", result.PtrInt32, i32)
	}
	if result.PtrInt16 == nil || *result.PtrInt16 != i16 {
		t.Errorf("PtrInt16: got %v, want %d", result.PtrInt16, i16)
	}
	if result.PtrInt8 == nil || *result.PtrInt8 != i8 {
		t.Errorf("PtrInt8: got %v, want %d", result.PtrInt8, i8)
	}
	if result.PtrUint == nil || *result.PtrUint != u {
		t.Errorf("PtrUint: got %v, want %d", result.PtrUint, u)
	}
	if result.PtrUint64 == nil || *result.PtrUint64 != u64 {
		t.Errorf("PtrUint64: got %v, want %d", result.PtrUint64, u64)
	}
	if result.PtrUint32 == nil || *result.PtrUint32 != u32 {
		t.Errorf("PtrUint32: got %v, want %d", result.PtrUint32, u32)
	}
	if result.PtrUint16 == nil || *result.PtrUint16 != u16 {
		t.Errorf("PtrUint16: got %v, want %d", result.PtrUint16, u16)
	}
	if result.PtrUint8 == nil || *result.PtrUint8 != u8 {
		t.Errorf("PtrUint8: got %v, want %d", result.PtrUint8, u8)
	}
	if result.PtrFloat64 == nil || *result.PtrFloat64 != f64 {
		t.Errorf("PtrFloat64: got %v, want %f", result.PtrFloat64, f64)
	}
	if result.PtrFloat32 == nil || *result.PtrFloat32 != f32 {
		t.Errorf("PtrFloat32: got %v, want %f", result.PtrFloat32, f32)
	}
	if result.PtrBool == nil || *result.PtrBool != b {
		t.Errorf("PtrBool: got %v, want %v", result.PtrBool, b)
	}
}

// TestStructDecoderAllSliceTypes tests all supported slice types for coverage.
func TestStructDecoderAllSliceTypes(t *testing.T) {
	type AllSlices struct {
		StringSlice  []string  `msgpack:"string_slice"`
		Int64Slice   []int64   `msgpack:"int64_slice"`
		IntSlice     []int     `msgpack:"int_slice"`
		Int32Slice   []int32   `msgpack:"int32_slice"`
		Int16Slice   []int16   `msgpack:"int16_slice"`
		Int8Slice    []int8    `msgpack:"int8_slice"`
		Uint64Slice  []uint64  `msgpack:"uint64_slice"`
		UintSlice    []uint    `msgpack:"uint_slice"`
		Uint32Slice  []uint32  `msgpack:"uint32_slice"`
		Uint16Slice  []uint16  `msgpack:"uint16_slice"`
		Float64Slice []float64 `msgpack:"float64_slice"`
		Float32Slice []float32 `msgpack:"float32_slice"`
		ByteSlice    []byte    `msgpack:"byte_slice"`
	}

	original := AllSlices{
		StringSlice:  []string{"a", "b", "c"},
		Int64Slice:   []int64{-1, -2, -3},
		IntSlice:     []int{1, 2, 3},
		Int32Slice:   []int32{-100, -200, -300},
		Int16Slice:   []int16{-10, -20, -30},
		Int8Slice:    []int8{-1, -2, -3},
		Uint64Slice:  []uint64{100, 200, 300},
		UintSlice:    []uint{10, 20, 30},
		Uint32Slice:  []uint32{1000, 2000, 3000},
		Uint16Slice:  []uint16{100, 200, 300},
		Float64Slice: []float64{1.1, 2.2, 3.3},
		Float32Slice: []float32{1.5, 2.5, 3.5},
		ByteSlice:    []byte{0x01, 0x02, 0x03},
	}

	enc := GetStructEncoder[AllSlices]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[AllSlices](false)
	var result AllSlices
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	// Verify all slice fields
	if len(result.StringSlice) != 3 || result.StringSlice[0] != "a" {
		t.Errorf("StringSlice mismatch: got %v", result.StringSlice)
	}
	if len(result.Int64Slice) != 3 || result.Int64Slice[0] != -1 {
		t.Errorf("Int64Slice mismatch: got %v", result.Int64Slice)
	}
	if len(result.IntSlice) != 3 || result.IntSlice[0] != 1 {
		t.Errorf("IntSlice mismatch: got %v", result.IntSlice)
	}
	if len(result.Int32Slice) != 3 || result.Int32Slice[0] != -100 {
		t.Errorf("Int32Slice mismatch: got %v", result.Int32Slice)
	}
	if len(result.Int16Slice) != 3 || result.Int16Slice[0] != -10 {
		t.Errorf("Int16Slice mismatch: got %v", result.Int16Slice)
	}
	if len(result.Int8Slice) != 3 || result.Int8Slice[0] != -1 {
		t.Errorf("Int8Slice mismatch: got %v", result.Int8Slice)
	}
	if len(result.Uint64Slice) != 3 || result.Uint64Slice[0] != 100 {
		t.Errorf("Uint64Slice mismatch: got %v", result.Uint64Slice)
	}
	if len(result.UintSlice) != 3 || result.UintSlice[0] != 10 {
		t.Errorf("UintSlice mismatch: got %v", result.UintSlice)
	}
	if len(result.Uint32Slice) != 3 || result.Uint32Slice[0] != 1000 {
		t.Errorf("Uint32Slice mismatch: got %v", result.Uint32Slice)
	}
	if len(result.Uint16Slice) != 3 || result.Uint16Slice[0] != 100 {
		t.Errorf("Uint16Slice mismatch: got %v", result.Uint16Slice)
	}
	if len(result.Float64Slice) != 3 || result.Float64Slice[0] != 1.1 {
		t.Errorf("Float64Slice mismatch: got %v", result.Float64Slice)
	}
	if len(result.Float32Slice) != 3 || result.Float32Slice[0] != 1.5 {
		t.Errorf("Float32Slice mismatch: got %v", result.Float32Slice)
	}
	if len(result.ByteSlice) != 3 || result.ByteSlice[0] != 0x01 {
		t.Errorf("ByteSlice mismatch: got %v", result.ByteSlice)
	}
}

// TestStructDecoderAllMapTypes tests all supported map types for coverage.
func TestStructDecoderAllMapTypes(t *testing.T) {
	type AllMaps struct {
		StringMap  map[string]string  `msgpack:"string_map"`
		IntMap     map[string]int     `msgpack:"int_map"`
		Int64Map   map[string]int64   `msgpack:"int64_map"`
		Int32Map   map[string]int32   `msgpack:"int32_map"`
		Uint64Map  map[string]uint64  `msgpack:"uint64_map"`
		Uint32Map  map[string]uint32  `msgpack:"uint32_map"`
		Float64Map map[string]float64 `msgpack:"float64_map"`
		BoolMap    map[string]bool    `msgpack:"bool_map"`
		AnyMap     map[string]any     `msgpack:"any_map"`
	}

	original := AllMaps{
		StringMap:  map[string]string{"a": "x", "b": "y"},
		IntMap:     map[string]int{"i1": 1, "i2": 2},
		Int64Map:   map[string]int64{"i64a": -100, "i64b": 200},
		Int32Map:   map[string]int32{"i32a": -50, "i32b": 50},
		Uint64Map:  map[string]uint64{"u64a": 1000, "u64b": 2000},
		Uint32Map:  map[string]uint32{"u32a": 100, "u32b": 200},
		Float64Map: map[string]float64{"f1": 1.5, "f2": 2.5},
		BoolMap:    map[string]bool{"t": true, "f": false},
		AnyMap:     map[string]any{"str": "hello", "num": int64(42), "bool": true},
	}

	enc := GetStructEncoder[AllMaps]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[AllMaps](false)
	var result AllMaps
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	// Verify all map fields
	if result.StringMap["a"] != "x" {
		t.Errorf("StringMap mismatch: got %v", result.StringMap)
	}
	if result.IntMap["i1"] != 1 {
		t.Errorf("IntMap mismatch: got %v", result.IntMap)
	}
	if result.Int64Map["i64a"] != -100 {
		t.Errorf("Int64Map mismatch: got %v", result.Int64Map)
	}
	if result.Int32Map["i32a"] != -50 {
		t.Errorf("Int32Map mismatch: got %v", result.Int32Map)
	}
	if result.Uint64Map["u64a"] != 1000 {
		t.Errorf("Uint64Map mismatch: got %v", result.Uint64Map)
	}
	if result.Uint32Map["u32a"] != 100 {
		t.Errorf("Uint32Map mismatch: got %v", result.Uint32Map)
	}
	if result.Float64Map["f1"] != 1.5 {
		t.Errorf("Float64Map mismatch: got %v", result.Float64Map)
	}
	if result.BoolMap["t"] != true {
		t.Errorf("BoolMap mismatch: got %v", result.BoolMap)
	}
	if result.AnyMap["str"] != "hello" {
		t.Errorf("AnyMap mismatch: got %v", result.AnyMap)
	}
}

// TestStructDecoderNestedStruct tests nested struct decoding for coverage.
func TestStructDecoderNestedStruct(t *testing.T) {
	type Inner struct {
		Value  int     `msgpack:"value"`
		Name   string  `msgpack:"name"`
		Active bool    `msgpack:"active"`
		Score  float64 `msgpack:"score"`
	}

	type Outer struct {
		ID    int64  `msgpack:"id"`
		Inner Inner  `msgpack:"inner"`
		Label string `msgpack:"label"`
	}

	original := Outer{
		ID: 123,
		Inner: Inner{
			Value:  42,
			Name:   "test",
			Active: true,
			Score:  3.14,
		},
		Label: "outer",
	}

	enc := GetStructEncoder[Outer]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[Outer](false)
	var result Outer
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	if result.ID != 123 {
		t.Errorf("ID mismatch: got %d", result.ID)
	}
	if result.Inner.Value != 42 {
		t.Errorf("Inner.Value mismatch: got %d", result.Inner.Value)
	}
	if result.Inner.Name != "test" {
		t.Errorf("Inner.Name mismatch: got %s", result.Inner.Name)
	}
	if result.Inner.Active != true {
		t.Errorf("Inner.Active mismatch: got %v", result.Inner.Active)
	}
	if result.Inner.Score != 3.14 {
		t.Errorf("Inner.Score mismatch: got %f", result.Inner.Score)
	}
	if result.Label != "outer" {
		t.Errorf("Label mismatch: got %s", result.Label)
	}
}

// TestStructDecoderDeepNested tests deeply nested struct decoding.
func TestStructDecoderDeepNested(t *testing.T) {
	type Level3 struct {
		Value int    `msgpack:"value"`
		Name  string `msgpack:"name"`
	}

	type Level2 struct {
		L3    Level3 `msgpack:"l3"`
		Count uint32 `msgpack:"count"`
	}

	type Level1 struct {
		L2   Level2  `msgpack:"l2"`
		Rate float32 `msgpack:"rate"`
	}

	original := Level1{
		L2: Level2{
			L3: Level3{
				Value: 999,
				Name:  "deep",
			},
			Count: 42,
		},
		Rate: 3.5,
	}

	enc := GetStructEncoder[Level1]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[Level1](false)
	var result Level1
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	if result.L2.L3.Value != 999 {
		t.Errorf("L2.L3.Value mismatch: got %d", result.L2.L3.Value)
	}
	if result.L2.L3.Name != "deep" {
		t.Errorf("L2.L3.Name mismatch: got %s", result.L2.L3.Name)
	}
}

// TestStructDecoderNestedMapAny tests map[string]any within struct.
func TestStructDecoderNestedMapAny(t *testing.T) {
	type Data struct {
		Meta map[string]any `msgpack:"meta"`
	}

	original := Data{
		Meta: map[string]any{
			"string": "hello",
			"number": int64(42),
			"bool":   true,
			"float":  3.14,
			"nested": map[string]any{
				"inner": "value",
			},
			"array": []any{int64(1), int64(2), int64(3)},
		},
	}

	enc := GetStructEncoder[Data]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	dec := GetStructDecoder[Data](false)
	var result Data
	err = dec.Decode(data, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	if result.Meta["string"] != "hello" {
		t.Errorf("Meta[string] mismatch: got %v", result.Meta["string"])
	}
	if result.Meta["number"] != int64(42) {
		t.Errorf("Meta[number] mismatch: got %v", result.Meta["number"])
	}
}

// TestStructDecoderZeroCopyMode tests zero-copy decoding mode.
func TestStructDecoderZeroCopyMode(t *testing.T) {
	type Data struct {
		Name   string            `msgpack:"name"`
		Values []string          `msgpack:"values"`
		Meta   map[string]string `msgpack:"meta"`
	}

	original := Data{
		Name:   "test",
		Values: []string{"a", "b", "c"},
		Meta:   map[string]string{"key": "value"},
	}

	enc := GetStructEncoder[Data]()
	data, err := enc.Encode(&original)
	if err != nil {
		t.Fatalf(errMsgEncodeFailed, err)
	}

	// Keep a copy of the data since zero-copy references it
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	dec := GetStructDecoder[Data](true) // zero-copy mode
	var result Data
	err = dec.Decode(dataCopy, &result)
	if err != nil {
		t.Fatalf(errMsgDecodeFailedCap, err)
	}

	if result.Name != "test" {
		t.Errorf("Name mismatch: got %s", result.Name)
	}
	if len(result.Values) != 3 {
		t.Errorf("Values length mismatch: got %d", len(result.Values))
	}
	if result.Meta["key"] != "value" {
		t.Errorf("Meta mismatch: got %v", result.Meta)
	}
}

// TestStructDecoderMapAnyWithVariousFormats tests decoding map[string]any with various msgpack formats.
// This exercises the typed_decode.go functions through struct decoder path.
func TestStructDecoderMapAnyWithVariousFormats(t *testing.T) {
	type Data struct {
		Meta map[string]any `msgpack:"meta"`
	}

	t.Run("str8 in map value", func(t *testing.T) {
		// Create a string > 31 bytes to trigger str8 format
		longStr := "this is a string longer than 31 bytes to trigger str8 format"
		original := Data{Meta: map[string]any{"key": longStr}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil || result.Meta["key"] != longStr {
			t.Errorf("str8: got %v, err=%v", result.Meta["key"], err)
		}
	})

	t.Run("str16 in map value", func(t *testing.T) {
		// Create a string > 255 bytes to trigger str16 format
		longStr := make([]byte, 300)
		for i := range longStr {
			longStr[i] = 'a'
		}
		original := Data{Meta: map[string]any{"key": string(longStr)}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil || result.Meta["key"] != string(longStr) {
			t.Errorf("str16: got len=%d, err=%v", len(result.Meta["key"].(string)), err)
		}
	})

	t.Run("bin8 in map value", func(t *testing.T) {
		original := Data{Meta: map[string]any{"key": []byte{1, 2, 3, 4, 5}}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		b, ok := result.Meta["key"].([]byte)
		if err != nil || !ok || len(b) != 5 {
			t.Errorf("bin8: got %v, err=%v", result.Meta["key"], err)
		}
	})

	t.Run("bin16 in map value", func(t *testing.T) {
		// Create binary > 255 bytes to trigger bin16 format
		bin := make([]byte, 300)
		for i := range bin {
			bin[i] = byte(i % 256)
		}
		original := Data{Meta: map[string]any{"key": bin}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		b, ok := result.Meta["key"].([]byte)
		if err != nil || !ok || len(b) != 300 {
			t.Errorf("bin16: got len=%d, err=%v", len(b), err)
		}
	})

	t.Run("array16 in map value", func(t *testing.T) {
		// Create array > 15 elements to trigger array16 format
		arr := make([]any, 20)
		for i := range arr {
			arr[i] = int64(i)
		}
		original := Data{Meta: map[string]any{"key": arr}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		a, ok := result.Meta["key"].([]any)
		if err != nil || !ok || len(a) != 20 {
			t.Errorf("array16: got len=%d, err=%v", len(a), err)
		}
	})

	t.Run("map16 in map value", func(t *testing.T) {
		// Create map > 15 entries to trigger map16 format
		m := make(map[string]any)
		for i := 0; i < 20; i++ {
			m[string('a'+byte(i))] = int64(i)
		}
		original := Data{Meta: map[string]any{"key": m}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		inner, ok := result.Meta["key"].(map[string]any)
		if err != nil || !ok || len(inner) != 20 {
			t.Errorf("map16: got len=%d, err=%v", len(inner), err)
		}
	})

	t.Run("nested arrays", func(t *testing.T) {
		original := Data{Meta: map[string]any{"key": []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		a, ok := result.Meta["key"].([]any)
		if err != nil || !ok || len(a) != 2 {
			t.Errorf("nested arrays: got %v, err=%v", result.Meta["key"], err)
		}
	})

	t.Run("nested maps", func(t *testing.T) {
		original := Data{Meta: map[string]any{"outer": map[string]any{"inner": int64(42)}}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		outer, ok := result.Meta["outer"].(map[string]any)
		if err != nil || !ok || outer["inner"] != int64(42) {
			t.Errorf("nested maps: got %v, err=%v", result.Meta, err)
		}
	})

	t.Run("all int formats", func(t *testing.T) {
		original := Data{Meta: map[string]any{
			"uint8":  uint64(200),
			"uint16": uint64(40000),
			"uint32": uint64(100000),
			"int8":   int64(-100),
			"int16":  int64(-1000),
			"int32":  int64(-100000),
		}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil {
			t.Errorf("all int formats: err=%v", err)
		}
	})

	t.Run("float formats", func(t *testing.T) {
		original := Data{Meta: map[string]any{
			"f32": float64(3.14),
			"f64": float64(2.71828),
		}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil {
			t.Errorf("float formats: err=%v", err)
		}
	})

	t.Run("nil value", func(t *testing.T) {
		original := Data{Meta: map[string]any{"key": nil}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil || result.Meta["key"] != nil {
			t.Errorf("nil value: got %v, err=%v", result.Meta["key"], err)
		}
	})

	t.Run("bool values", func(t *testing.T) {
		original := Data{Meta: map[string]any{"t": true, "f": false}}
		enc := GetStructEncoder[Data]()
		data, _ := enc.Encode(&original)

		dec := GetStructDecoder[Data](false)
		var result Data
		err := dec.Decode(data, &result)
		if err != nil || result.Meta["t"] != true || result.Meta["f"] != false {
			t.Errorf("bool values: got %v, err=%v", result.Meta, err)
		}
	})
}

// TestStructDecoderMapAnyZeroCopy tests decoding map[string]any in zero-copy mode.
func TestStructDecoderMapAnyZeroCopy(t *testing.T) {
	type Data struct {
		Meta map[string]any `msgpack:"meta"`
	}

	longStr := "this is a string that triggers str8 format for testing"
	original := Data{Meta: map[string]any{
		"str":   longStr,
		"bin":   []byte{1, 2, 3, 4, 5},
		"arr":   []any{int64(1), int64(2), int64(3)},
		"map":   map[string]any{"nested": "value"},
		"int":   int64(42),
		"float": float64(3.14),
	}}

	enc := GetStructEncoder[Data]()
	data, _ := enc.Encode(&original)

	// Keep a copy for zero-copy mode
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	dec := GetStructDecoder[Data](true) // zero-copy mode
	var result Data
	err := dec.Decode(dataCopy, &result)
	if err != nil {
		t.Fatalf("zero-copy decode failed: %v", err)
	}

	if result.Meta["str"] != longStr {
		t.Errorf("str mismatch")
	}
	if len(result.Meta["bin"].([]byte)) != 5 {
		t.Errorf("bin mismatch")
	}
	if len(result.Meta["arr"].([]any)) != 3 {
		t.Errorf("arr mismatch")
	}
}

// TestStructDecoderFloatFromIntFormats tests decoding various int formats into float64 field.
func TestStructDecoderFloatFromIntFormats(t *testing.T) {
	type Data struct {
		F float64 `msgpack:"f"`
	}

	tests := []struct {
		name     string
		buildMsg func(*Encoder)
		want     float64
	}{
		{"positive fixint", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(42) // positive fixint
		}, 42.0},
		{"negative fixint", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(0xe0) // -32
		}, -32.0},
		{"uint8", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatUint8)
			e.writeByte(200)
		}, 200.0},
		{"uint16", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatUint16)
			e.writeUint16(40000)
		}, 40000.0},
		{"uint32", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatUint32)
			e.writeUint32(100000)
		}, 100000.0},
		{"int8", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatInt8)
			e.writeByte(0x80) // -128
		}, -128.0},
		{"int16", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatInt16)
			e.writeUint16(0x8000) // -32768
		}, -32768.0},
		{"int32", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatInt32)
			e.writeUint32(0xFFFFFF00) // -256
		}, -256.0},
		{"int64", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatInt64)
			e.writeUint64(0xFFFFFFFFFFFFFFFE) // -2
		}, -2.0},
		{"float32", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.EncodeFloat32(3.14)
		}, 3.14},
		{"float64", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.EncodeFloat64(2.71828)
		}, 2.71828},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(64)
			tc.buildMsg(e)
			var result Data
			err := dec.Decode(e.Bytes(), &result)
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			diff := result.F - tc.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("got %v, want %v", result.F, tc.want)
			}
		})
	}
}

// TestStructDecoderFloat32FieldFromIntFormats tests decoding various formats into float32 field.
func TestStructDecoderFloat32FieldFromIntFormats(t *testing.T) {
	type Data struct {
		F float32 `msgpack:"f"`
	}

	tests := []struct {
		name     string
		buildMsg func(*Encoder)
		want     float32
	}{
		{"positive fixint", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(42)
		}, 42.0},
		{"uint8", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatUint8)
			e.writeByte(200)
		}, 200.0},
		{"int8", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.writeByte(formatInt8)
			e.writeByte(0x80)
		}, -128.0},
		{"float32", func(e *Encoder) {
			e.writeByte(fixmapPrefix | 1)
			e.writeByte(fixstrPrefix | 1)
			e.writeByte('f')
			e.EncodeFloat32(3.14)
		}, 3.14},
	}

	dec := GetStructDecoder[Data](false)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(64)
			tc.buildMsg(e)
			var result Data
			err := dec.Decode(e.Bytes(), &result)
			if err != nil {
				t.Fatalf("err=%v", err)
			}
			diff := result.F - tc.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("got %v, want %v", result.F, tc.want)
			}
		})
	}
}

// TestLargeArrayAndMapEncodings tests array32 and map32 format encoding.
func TestLargeArrayAndMapEncodings(t *testing.T) {
	t.Run("array32 encoding", func(t *testing.T) {
		e := NewEncoder(1024)
		// Create array with > 65535 elements would be too slow
		// Instead, directly test the format by manually constructing
		e.writeByte(formatArray32)
		e.writeUint32(3)
		e.EncodeInt(1)
		e.EncodeInt(2)
		e.EncodeInt(3)

		var result []int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 3 {
			t.Errorf("got len=%d, err=%v", len(result), err)
		}
	})

	t.Run("map32 encoding", func(t *testing.T) {
		e := NewEncoder(1024)
		e.writeByte(formatMap32)
		e.writeUint32(2)
		e.EncodeString("a")
		e.EncodeInt(1)
		e.EncodeString("b")
		e.EncodeInt(2)

		var result map[string]int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 2 {
			t.Errorf("got len=%d, err=%v", len(result), err)
		}
	})
}
