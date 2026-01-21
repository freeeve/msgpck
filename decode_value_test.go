package msgpck

import (
	"testing"
)

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
				t.Errorf(errMsgFmtSV, f.name, err)
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
				t.Errorf(errMsgFmtSV, f.name, err)
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

func TestDecodeValueInvalidFormat(t *testing.T) {
	data := []byte{0xc1} // never used format
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

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
				t.Errorf(errMsgUnexpectedEOF, err)
			}
		})
	}
}

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
				t.Errorf(errMsgFmtSV, tc.name, err)
			}
			if v.Type == TypeNil {
				t.Errorf("%s: unexpected nil", tc.name)
			}
		})
	}
}

func TestDecodeValueExtTooLong(t *testing.T) {
	cfg := DefaultConfig().WithMaxExtLen(1)
	data := append([]byte{formatExt8, 5, 1}, make([]byte, 5)...)
	d := NewDecoderWithConfig(data, cfg)
	_, err := d.Decode()
	if err != ErrExtTooLong {
		t.Errorf("expected ErrExtTooLong, got %v", err)
	}
}

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

func TestDecodeValueBytesEOFPaths(t *testing.T) {
	type Data struct {
		B []byte `msgpack:"b"`
	}

	t.Run("str16 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatStr16, 0} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("str32 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatStr32, 0, 0} // missing length bytes
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("bin8 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin8} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("bin16 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin16, 0} // missing length byte
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})

	t.Run("bin32 EOF", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'b', formatBin32, 0, 0} // missing length bytes
		var d Data
		err := UnmarshalStruct(data, &d)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

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
			t.Errorf(errMsgStringTooLong, err)
		}
	})

	t.Run("bin16 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxBinaryLen(1)
		data := []byte{0x81, 0xa1, 'b', formatBin16, 0, 5, 1, 2, 3, 4, 5}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrBinaryTooLong {
			t.Errorf(errMsgBinaryTooLong, err)
		}
	})
}

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
			t.Errorf(errMsgStringTooLong, err)
		}
	})

	t.Run("str32 too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(1)
		data := []byte{0x81, 0xa1, 's', formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
		var d Data
		err := UnmarshalStructWithConfig(data, &d, cfg)
		if err != ErrStringTooLong {
			t.Errorf(errMsgStringTooLong, err)
		}
	})
}

func TestDecodeValueStringReadBytesEOF(t *testing.T) {
	// fixstr says 5 bytes but only 2 provided
	data := []byte{0xa5, 'h', 'i'}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueBinaryReadBytesEOF(t *testing.T) {
	// bin8 says 5 bytes but only 2 provided
	data := []byte{formatBin8, 5, 0x01, 0x02}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueStr16EOF(t *testing.T) {
	// str16 but only 1 byte for length
	data := []byte{formatStr16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueStr32EOF(t *testing.T) {
	// str32 but only 2 bytes for length
	data := []byte{formatStr32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueBin16EOF(t *testing.T) {
	// bin16 but only 1 byte for length
	data := []byte{formatBin16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueBin32EOF(t *testing.T) {
	// bin32 but only 2 bytes for length
	data := []byte{formatBin32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueArray16EOF(t *testing.T) {
	// array16 but only 1 byte for length
	data := []byte{formatArray16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueArray32EOF(t *testing.T) {
	// array32 but only 2 bytes for length
	data := []byte{formatArray32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueMap16EOF(t *testing.T) {
	// map16 but only 1 byte for length
	data := []byte{formatMap16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueMap32EOF(t *testing.T) {
	// map32 but only 2 bytes for length
	data := []byte{formatMap32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueExt8EOF(t *testing.T) {
	// ext8 but no length byte
	data := []byte{formatExt8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueExt16EOF(t *testing.T) {
	// ext16 but only 1 byte for length
	data := []byte{formatExt16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueExt32EOF(t *testing.T) {
	// ext32 but only 2 bytes for length
	data := []byte{formatExt32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueUint8EOF(t *testing.T) {
	data := []byte{formatUint8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueUint16EOF(t *testing.T) {
	data := []byte{formatUint16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueUint32EOF(t *testing.T) {
	data := []byte{formatUint32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueUint64EOF(t *testing.T) {
	data := []byte{formatUint64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueInt8EOF(t *testing.T) {
	data := []byte{formatInt8}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueInt16EOF(t *testing.T) {
	data := []byte{formatInt16, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueInt32EOF(t *testing.T) {
	data := []byte{formatInt32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueInt64EOF(t *testing.T) {
	data := []byte{formatInt64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueFloat32EOF(t *testing.T) {
	data := []byte{formatFloat32, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueFloat64EOF(t *testing.T) {
	data := []byte{formatFloat64, 0x00, 0x00, 0x00, 0x00}
	d := NewDecoder(data)
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

func TestDecodeValueStringReadBytesError(t *testing.T) {
	// str16 with length but not enough data
	enc := NewEncoder(64)
	enc.writeByte(formatStr16)
	enc.writeUint16(100) // says 100 bytes but only have a few

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueBinaryReadBytesError(t *testing.T) {
	// bin16 with length but not enough data
	enc := NewEncoder(64)
	enc.writeByte(formatBin16)
	enc.writeUint16(100) // says 100 bytes but only have a few

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueStringStr8ReadBytesEOF(t *testing.T) {
	// str8 says 50 bytes but only provides 5
	enc := NewEncoder(64)
	enc.writeByte(formatStr8)
	enc.writeByte(50)
	enc.writeBytes([]byte("hello"))

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

func TestDecodeValueBinaryBin8ReadBytesEOF(t *testing.T) {
	// bin8 says 50 bytes but only provides 5
	enc := NewEncoder(64)
	enc.writeByte(formatBin8)
	enc.writeByte(50)
	enc.writeBytes([]byte{1, 2, 3, 4, 5})

	d := NewDecoder(enc.Bytes())
	_, err := d.Decode()
	if err != ErrUnexpectedEOF {
		t.Errorf(errMsgUnexpectedEOF, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}

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
		t.Errorf(errMsgTypeMismatch, err)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Error(errMsgEOFError)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 1000 {
		t.Errorf("expected 1000, got %d", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 100000 {
		t.Errorf("expected 100000, got %d", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != -50 {
		t.Errorf("expected -50, got %d", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 1000 {
		t.Errorf("expected 1000, got %d", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 100000 {
		t.Errorf("expected 100000, got %d", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 42.0 {
		t.Errorf("expected 42.0, got %f", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 1000.0 {
		t.Errorf("expected 1000.0, got %f", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != 100000.0 {
		t.Errorf("expected 100000.0, got %f", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if s.Value != -50.0 {
		t.Errorf("expected -50.0, got %f", s.Value)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgUnexpectedErr, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgStringTooLong, err)
	}
}

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
		t.Errorf(errMsgBinaryTooLong, err)
	}
}
