package msgpck

import (
	"testing"
)

func TestTypedDecodeExtra(t *testing.T) {
	t.Run("UnmarshalMap", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(1)
		e.EncodeString("key")
		e.EncodeString("value")
		b := make([]byte, len(e.Bytes()))
		copy(b, e.Bytes())

		var m map[string]any
		err := Unmarshal(b, &m)
		if err != nil || m["key"] != "value" {
			t.Error("UnmarshalMap failed")
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

		var m map[string]any
		err := Unmarshal(b, &m)
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

func TestTypedDecodeMapFormats(t *testing.T) {
	t.Run("map32 in typed decode", func(t *testing.T) {
		data := []byte{
			formatMap32, 0, 0, 0, 1, // map32 length 1
			0xa1, 'k', // fixstr "k"
			0x01, // positive fixint 1
		}

		var m map[string]any
		err := Unmarshal(data, &m)
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

		var m map[string]any
		err := Unmarshal(data, &m)
		if err != nil {
			t.Fatal(err)
		}
		inner := m["k"].(map[string]any)
		if inner["n"] != int64(1) {
			t.Error("nested map32 failed")
		}
	})
}

func TestTypedDecodeMapStringAnyFormats(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var m map[string]any
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("map16 in Unmarshal failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var m map[string]any
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("map32 in Unmarshal failed")
		}
	})
}

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
			// Test via Unmarshal into map which uses decodeAnyValue
			mapData := append([]byte{0x81, 0xa1, 'v'}, tc.data...)
			var m map[string]any
			err := Unmarshal(mapData, &m)
			if err != nil {
				t.Errorf("%s failed: %v", tc.name, err)
			}
			if m["v"] == nil && tc.name != "nil" {
				t.Errorf("%s: value is nil", tc.name)
			}
		})
	}
}

func TestTypedDecodeNilMap(t *testing.T) {
	data := []byte{formatNil}
	var m map[string]any
	err := Unmarshal(data, &m)
	if err != nil || m != nil {
		t.Error("nil map should return nil")
	}
}

func TestTypedDecodeMapStringString(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		data := []byte{formatMap16, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("map16 failed")
		}
	})

	t.Run("map32", func(t *testing.T) {
		data := []byte{formatMap32, 0, 0, 0, 1, 0xa1, 'k', 0xa1, 'v'}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("map32 failed")
		}
	})

	t.Run("nil", func(t *testing.T) {
		data := []byte{formatNil}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m != nil {
			t.Error("nil map failed")
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		data := []byte{0x42}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != ErrTypeMismatch {
			t.Error("expected ErrTypeMismatch")
		}
	})

	t.Run("str8 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr8, 1, 'v'}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("str8 value failed")
		}
	})

	t.Run("str16 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr16, 0, 1, 'v'}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("str16 value failed")
		}
	})

	t.Run("str32 value", func(t *testing.T) {
		data := []byte{0x81, 0xa1, 'k', formatStr32, 0, 0, 0, 1, 'v'}
		var m map[string]string
		err := Unmarshal(data, &m)
		if err != nil || m["k"] != "v" {
			t.Error("str32 value failed")
		}
	})
}

// TestTypedDecode tests typed decode functions
func TestTypedDecode(t *testing.T) {
	t.Run("UnmarshalMapStringAny", func(t *testing.T) {
		data, _ := Marshal(map[string]any{"a": int64(1)})
		var m map[string]any
		err := Unmarshal(data, &m)
		if err != nil || m["a"] != int64(1) {
			t.Error("decode failed")
		}
	})

	t.Run("UnmarshalMapStringString", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeMapHeader(2)
		e.EncodeString("key1")
		e.EncodeString("value1")
		e.EncodeString("key2")
		e.EncodeString("value2")

		var m map[string]string
		err := Unmarshal(e.Bytes(), &m)
		if err != nil || m["key1"] != "value1" || m["key2"] != "value2" {
			t.Error("UnmarshalMapStringString failed")
		}
	})
}
