package msgpck

import (
	"testing"
)

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
				t.Errorf(errMsgUnexpectedEOF, err)
			}
		})
	}
}

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
			t.Error(errMsgExpectedEOF)
		}
	})
}

func TestDecoderPeekByteEOF(t *testing.T) {
	data := []byte{}
	d := NewDecoder(data)
	_, err := d.peekByte()
	if err != ErrUnexpectedEOF {
		t.Error(errMsgExpectedEOF)
	}
}

func TestDecoderReadFloatEOF(t *testing.T) {
	t.Run("readFloat32 EOF", func(t *testing.T) {
		data := []byte{0x00, 0x00} // only 2 bytes, need 4
		d := NewDecoder(data)
		_, err := d.readFloat32()
		if err != ErrUnexpectedEOF {
			t.Error(errMsgExpectedEOF)
		}
	})

	t.Run("readFloat64 EOF", func(t *testing.T) {
		data := []byte{0x00, 0x00, 0x00, 0x00} // only 4 bytes, need 8
		d := NewDecoder(data)
		_, err := d.readFloat64()
		if err != ErrUnexpectedEOF {
			t.Error(errMsgExpectedEOF)
		}
	})
}

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

func TestReadBytesEOF(t *testing.T) {
	data := []byte{0xa5, 'h', 'e'} // fixstr 5 with only 2 bytes
	d := NewDecoder(data)
	_, err := d.readStringBytes()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeExtEOF(t *testing.T) {
	t.Run("ext8 length EOF", func(t *testing.T) {
		data := []byte{formatExt8} // ext8 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("ext16 length EOF", func(t *testing.T) {
		data := []byte{formatExt16, 0} // ext16 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("ext32 length EOF", func(t *testing.T) {
		data := []byte{formatExt32, 0, 0} // ext32 with missing length
		d := NewDecoder(data)
		_, err := d.Decode()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

func TestDecodeStringAnyEOF(t *testing.T) {
	t.Run("str16 length EOF", func(t *testing.T) {
		data := []byte{formatStr16, 0} // str16 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("str32 length EOF", func(t *testing.T) {
		data := []byte{formatStr32, 0, 0} // str32 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

func TestDecodeBinaryAnyEOF(t *testing.T) {
	t.Run("bin16 length EOF", func(t *testing.T) {
		data := []byte{formatBin16, 0} // bin16 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("bin32 length EOF", func(t *testing.T) {
		data := []byte{formatBin32, 0, 0} // bin32 with missing length
		d := NewDecoder(data)
		_, err := d.DecodeAny()
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

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
		t.Errorf(errMsgGot42, s.A)
	}
}

func TestDecodeStringKeyStr8EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str8 format but no length
	data := []byte{0x81, formatStr8}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStringKeyStr16EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str16 format but no length
	data := []byte{0x81, formatStr16}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStringKeyStr32EOF(t *testing.T) {
	type S struct {
		A int `msgpack:"a"`
	}

	// str32 format but no length
	data := []byte{0x81, formatStr32}
	var s S
	err := UnmarshalStruct(data, &s)
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

func TestDecodeExtDataEOF(t *testing.T) {
	// fixext1 (1 byte data) but no data after type
	data := []byte{formatFixExt1, 0x01} // type=1, no data
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeExtTypeEOF(t *testing.T) {
	// ext8 format with length but no type
	data := []byte{formatExt8, 0x01} // length=1, no type
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeArrayValueEOF(t *testing.T) {
	// Array header says 3 but only 2 values
	data := []byte{0x93, 0x01, 0x02} // array of 3, values 1, 2
	d := NewDecoder(data)
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestReadBytesInsufficientData(t *testing.T) {
	d := NewDecoder([]byte{0x01, 0x02})
	_, err := d.readBytes(5) // only 2 bytes available
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

func TestDecodeStringAnyReadBytesEOF(t *testing.T) {
	// str8 says 10 bytes but only 5 provided
	data := []byte{formatStr8, 10, 'a', 'b', 'c', 'd', 'e'}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

func TestDecodeBinaryAnyReadBytesEOF(t *testing.T) {
	// bin8 says 10 bytes but only 5 provided
	data := []byte{formatBin8, 10, 'a', 'b', 'c', 'd', 'e'}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStringReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeBinaryReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStringStr16ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeStringStr32ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatStr32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, "short"...)
	// Only 5 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeBinaryBin16ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

func TestDecodeBinaryBin32ReadBytesError(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatBin32)
	enc.writeUint32(50) // length = 50
	enc.buf = append(enc.buf, []byte{1, 2, 3}...)
	// Only 3 bytes, expected 50

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err == nil {
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

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

// TestZeroCopyDecode tests that string data points into source buffer
func TestZeroCopyDecode(t *testing.T) {
	data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // fixstr "hello"

	d := NewDecoder(data)
	v, err := d.Decode()
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if v.Type != TypeString {
		t.Fatalf("expected TypeString, got %v", v.Type)
	}

	// Check that Bytes points into the original data
	if &v.Bytes[0] != &data[1] {
		t.Error("string bytes not pointing into source buffer (not zero-copy)")
	}
}

// TestUnexpectedEOF tests handling of truncated data
func TestUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"truncated uint16", []byte{formatUint16, 0x01}},
		{"truncated string", []byte{0xa5, 'h', 'e'}}, // fixstr 5, only 2 chars
		{"truncated array", []byte{0x92, 0x01}},      // fixarray 2, only 1 element
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(tt.data)
			_, err := d.Decode()
			if err != ErrUnexpectedEOF {
				t.Errorf(errMsgUnexpectedEOF, err)
			}
		})
	}
}
