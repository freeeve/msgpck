package msgpck

import (
	"testing"
)

const errMsgGotLenErr = "got len=%d, err=%v"

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

func TestDecodeAnyStr32(t *testing.T) {
	data := []byte{formatStr32, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	d := NewDecoder(data)
	v, err := d.DecodeAny()
	if err != nil || v != "hello" {
		t.Error("str32 DecodeAny failed")
	}
}

func TestValidationLimitsDecodeAny(t *testing.T) {
	t.Run("string too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(2)
		data := []byte{0xa5, 'h', 'e', 'l', 'l', 'o'} // fixstr "hello" (5 bytes)
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrStringTooLong {
			t.Errorf(errMsgStringTooLong, err)
		}
	})

	t.Run("binary too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxBinaryLen(2)
		data := []byte{formatBin8, 5, 1, 2, 3, 4, 5}
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrBinaryTooLong {
			t.Errorf(errMsgBinaryTooLong, err)
		}
	})

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(2)
		data := []byte{0x95, 1, 2, 3, 4, 5} // fixarray 5 elements
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrArrayTooLong {
			t.Errorf(errMsgArrayTooLong, err)
		}
	})

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(1)
		data := []byte{0x82, 0xa1, 'a', 1, 0xa1, 'b', 2} // fixmap 2 elements
		d := NewDecoderWithConfig(data, cfg)
		_, err := d.DecodeAny()
		if err != ErrMapTooLong {
			t.Errorf(errMsgMapTooLong, err)
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

func TestDecodeAnyInvalidFormat(t *testing.T) {
	// Use a format byte that's not valid (0xc1 is never used in msgpack)
	data := []byte{0xc1}
	d := NewDecoder(data)
	_, err := d.DecodeAny()
	if err != ErrInvalidFormat {
		t.Errorf("expected ErrInvalidFormat, got %v", err)
	}
}

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
			var result any
			err := Unmarshal(tt.data, &result)
			if err != nil {
				t.Errorf("failed to decode %s: %v", tt.name, err)
			}
		})
	}
}

func TestDecodeAnyValueUint64Overflow(t *testing.T) {
	enc := NewEncoder(64)
	enc.writeByte(formatUint64)
	enc.writeUint64(0xFFFFFFFFFFFFFFFF) // Max uint64, overflows int64

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if _, ok := v.(uint64); !ok {
		t.Errorf("expected uint64, got %T", v)
	}
}

func TestDecodeAnyValueFloat32(t *testing.T) {
	enc := NewEncoder(64)
	enc.EncodeFloat32(3.14)

	d := NewDecoder(enc.Bytes())
	v, err := d.DecodeAny()
	if err != nil {
		t.Errorf(errMsgUnexpectedErr, err)
	}
	if _, ok := v.(float64); !ok {
		t.Errorf("expected float64, got %T", v)
	}
}

// TestUnmarshalIntoTypedValues tests reflection-based decoding into typed values.
func TestUnmarshalIntoTypedValues(t *testing.T) {
	t.Run("uint64", func(t *testing.T) {
		data, _ := Marshal(uint64(12345))
		var result uint64
		err := Unmarshal(data, &result)
		if err != nil || result != 12345 {
			t.Errorf("uint64 unmarshal failed: %v, got %d", err, result)
		}
	})

	t.Run("uint32", func(t *testing.T) {
		data, _ := Marshal(uint64(999))
		var result uint32
		err := Unmarshal(data, &result)
		if err != nil || result != 999 {
			t.Errorf("uint32 unmarshal failed: %v, got %d", err, result)
		}
	})

	t.Run("uint16", func(t *testing.T) {
		data, _ := Marshal(uint64(100))
		var result uint16
		err := Unmarshal(data, &result)
		if err != nil || result != 100 {
			t.Errorf("uint16 unmarshal failed: %v, got %d", err, result)
		}
	})

	t.Run("uint8", func(t *testing.T) {
		data, _ := Marshal(uint64(50))
		var result uint8
		err := Unmarshal(data, &result)
		if err != nil || result != 50 {
			t.Errorf("uint8 unmarshal failed: %v, got %d", err, result)
		}
	})

	t.Run("float64", func(t *testing.T) {
		data, _ := Marshal(3.14159)
		var result float64
		err := Unmarshal(data, &result)
		if err != nil || result != 3.14159 {
			t.Errorf("float64 unmarshal failed: %v, got %f", err, result)
		}
	})

	t.Run("float32", func(t *testing.T) {
		data, _ := Marshal(float64(2.5))
		var result float32
		err := Unmarshal(data, &result)
		if err != nil || result != 2.5 {
			t.Errorf("float32 unmarshal failed: %v, got %f", err, result)
		}
	})

	t.Run("bool true", func(t *testing.T) {
		data, _ := Marshal(true)
		var result bool
		err := Unmarshal(data, &result)
		if err != nil || result != true {
			t.Errorf("bool unmarshal failed: %v, got %v", err, result)
		}
	})

	t.Run("bool false", func(t *testing.T) {
		data, _ := Marshal(false)
		var result bool
		err := Unmarshal(data, &result)
		if err != nil || result != false {
			t.Errorf("bool unmarshal failed: %v, got %v", err, result)
		}
	})
}

// TestUnmarshalIntoSlices tests reflection-based slice decoding.
func TestUnmarshalIntoSlices(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		data, _ := Marshal([]int64{1, 2, 3})
		var result []int
		err := Unmarshal(data, &result)
		if err != nil {
			t.Errorf("int slice unmarshal failed: %v", err)
		}
		if len(result) != 3 || result[0] != 1 {
			t.Errorf("int slice mismatch: got %v", result)
		}
	})

	t.Run("float64 slice", func(t *testing.T) {
		data, _ := Marshal([]float64{1.1, 2.2, 3.3})
		var result []float64
		err := Unmarshal(data, &result)
		if err != nil || len(result) != 3 {
			t.Errorf("float64 slice unmarshal failed: %v, got %v", err, result)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		data, _ := Marshal([]string{"a", "b", "c"})
		var result []string
		err := Unmarshal(data, &result)
		if err != nil || len(result) != 3 || result[0] != "a" {
			t.Errorf("string slice unmarshal failed: %v, got %v", err, result)
		}
	})

	t.Run("byte slice from binary", func(t *testing.T) {
		data, _ := Marshal([]byte{1, 2, 3, 4, 5})
		var result []byte
		err := Unmarshal(data, &result)
		if err != nil || len(result) != 5 || result[0] != 1 {
			t.Errorf("byte slice unmarshal failed: %v, got %v", err, result)
		}
	})
}

// TestUnmarshalIntoPrimitivePointers tests decoding into pointer types.
func TestUnmarshalIntoPrimitivePointers(t *testing.T) {
	t.Run("string pointer", func(t *testing.T) {
		data, _ := Marshal("hello")
		var result *string
		err := Unmarshal(data, &result)
		if err != nil || result == nil || *result != "hello" {
			t.Errorf("string pointer unmarshal failed: %v", err)
		}
	})

	t.Run("int64 pointer", func(t *testing.T) {
		data, _ := Marshal(int64(42))
		var result *int64
		err := Unmarshal(data, &result)
		if err != nil || result == nil || *result != 42 {
			t.Errorf("int64 pointer unmarshal failed: %v", err)
		}
	})
}

// TestUnmarshalIntFormats tests decoding various integer formats into typed values.
func TestUnmarshalIntFormats(t *testing.T) {
	// Test all integer format variants
	t.Run("int8 format", func(t *testing.T) {
		// Encode as int8 (-128)
		enc := NewEncoder(16)
		enc.writeByte(formatInt8)
		enc.writeByte(0x80) // -128
		var result int
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != -128 {
			t.Errorf("int8 format failed: %v, got %d", err, result)
		}
	})

	t.Run("int16 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatInt16)
		enc.writeUint16(0x8000) // -32768
		var result int
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != -32768 {
			t.Errorf("int16 format failed: %v, got %d", err, result)
		}
	})

	t.Run("int32 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatInt32)
		enc.writeUint32(0xFFFFFF00) // -256
		var result int
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != -256 {
			t.Errorf("int32 format failed: %v, got %d", err, result)
		}
	})

	t.Run("int64 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatInt64)
		enc.writeUint64(0xFFFFFFFFFFFFFF00) // -256
		var result int64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != -256 {
			t.Errorf("int64 format failed: %v, got %d", err, result)
		}
	})

	t.Run("uint8 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatUint8)
		enc.writeByte(200)
		var result uint
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != 200 {
			t.Errorf("uint8 format failed: %v, got %d", err, result)
		}
	})

	t.Run("uint16 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatUint16)
		enc.writeUint16(40000)
		var result uint
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != 40000 {
			t.Errorf("uint16 format failed: %v, got %d", err, result)
		}
	})

	t.Run("uint32 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatUint32)
		enc.writeUint32(3000000000)
		var result uint64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != 3000000000 {
			t.Errorf("uint32 format failed: %v, got %d", err, result)
		}
	})

	t.Run("uint64 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.writeByte(formatUint64)
		enc.writeUint64(0xFFFFFFFFFFFFFFFF)
		var result uint64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != 0xFFFFFFFFFFFFFFFF {
			t.Errorf("uint64 format failed: %v, got %d", err, result)
		}
	})
}

// TestUnmarshalFloatFormats tests decoding float formats.
func TestUnmarshalFloatFormats(t *testing.T) {
	t.Run("float32 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.EncodeFloat32(3.14)
		var result float32
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil {
			t.Errorf("float32 format failed: %v", err)
		}
	})

	t.Run("float64 format", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.EncodeFloat64(3.14159265359)
		var result float64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != 3.14159265359 {
			t.Errorf("float64 format failed: %v, got %f", err, result)
		}
	})

	t.Run("float32 into float64", func(t *testing.T) {
		enc := NewEncoder(16)
		enc.EncodeFloat32(2.5)
		var result float64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || result != float64(float32(2.5)) {
			t.Errorf("float32 into float64 failed: %v, got %f", err, result)
		}
	})
}

// TestDecodeArrayFormats tests decoding large arrays.
func TestDecodeArrayFormats(t *testing.T) {
	t.Run("array16", func(t *testing.T) {
		enc := NewEncoder(1024)
		enc.writeByte(formatArray16)
		enc.writeUint16(100)
		for i := 0; i < 100; i++ {
			enc.EncodeInt(int64(i))
		}
		var result []int64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || len(result) != 100 {
			t.Errorf("array16 failed: %v, len=%d", err, len(result))
		}
	})

	t.Run("array32", func(t *testing.T) {
		enc := NewEncoder(1024)
		enc.writeByte(formatArray32)
		enc.writeUint32(50)
		for i := 0; i < 50; i++ {
			enc.EncodeInt(int64(i))
		}
		var result []int64
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil || len(result) != 50 {
			t.Errorf("array32 failed: %v, len=%d", err, len(result))
		}
	})
}

// TestDecodeMapFormats tests decoding large maps.
func TestDecodeMapFormats(t *testing.T) {
	t.Run("map16", func(t *testing.T) {
		enc := NewEncoder(2048)
		enc.writeByte(formatMap16)
		enc.writeUint16(50)
		for i := 0; i < 50; i++ {
			enc.EncodeString("k" + string(rune('0'+i%10)))
			enc.EncodeInt(int64(i))
		}
		var result map[string]int
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil {
			t.Errorf("map16 failed: %v", err)
		}
	})

	t.Run("map32", func(t *testing.T) {
		enc := NewEncoder(2048)
		enc.writeByte(formatMap32)
		enc.writeUint32(30)
		for i := 0; i < 30; i++ {
			enc.EncodeString("key" + string(rune('a'+i%26)))
			enc.EncodeInt(int64(i))
		}
		var result map[string]int
		err := Unmarshal(enc.Bytes(), &result)
		if err != nil {
			t.Errorf("map32 failed: %v", err)
		}
	})
}

// TestUnmarshalReflectionIntFormats tests reflection-based int decoding with all formats.
func TestUnmarshalReflectionIntFormats(t *testing.T) {
	tests := []struct {
		name   string
		encode func(*Encoder)
		want   int64
	}{
		{"int8", func(e *Encoder) { e.writeByte(formatInt8); e.writeByte(0x80) }, -128},
		{"int16", func(e *Encoder) { e.writeByte(formatInt16); e.writeUint16(0x8000) }, -32768},
		{"int32", func(e *Encoder) { e.writeByte(formatInt32); e.writeUint32(0xFFFFFF00) }, -256},
		{"int64", func(e *Encoder) { e.writeByte(formatInt64); e.writeUint64(0xFFFFFFFFFFFFFC18) }, -1000},
		{"uint8 as int", func(e *Encoder) { e.writeByte(formatUint8); e.writeByte(200) }, 200},
		{"uint16 as int", func(e *Encoder) { e.writeByte(formatUint16); e.writeUint16(40000) }, 40000},
		{"uint32 as int", func(e *Encoder) { e.writeByte(formatUint32); e.writeUint32(100000) }, 100000},
		{"uint64 as int", func(e *Encoder) { e.writeByte(formatUint64); e.writeUint64(500000) }, 500000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(16)
			tc.encode(e)
			var result int64
			err := Unmarshal(e.Bytes(), &result)
			if err != nil || result != tc.want {
				t.Errorf("%s: got %d, want %d, err=%v", tc.name, result, tc.want, err)
			}
		})
	}
}

// TestUnmarshalReflectionUintFormats tests reflection-based uint decoding with all formats.
func TestUnmarshalReflectionUintFormats(t *testing.T) {
	tests := []struct {
		name   string
		encode func(*Encoder)
		want   uint64
	}{
		{"uint8", func(e *Encoder) { e.writeByte(formatUint8); e.writeByte(200) }, 200},
		{"uint16", func(e *Encoder) { e.writeByte(formatUint16); e.writeUint16(40000) }, 40000},
		{"uint32", func(e *Encoder) { e.writeByte(formatUint32); e.writeUint32(3000000000) }, 3000000000},
		{"uint64", func(e *Encoder) { e.writeByte(formatUint64); e.writeUint64(0xFFFFFFFFFFFFFFFF) }, 0xFFFFFFFFFFFFFFFF},
		{"int8 as uint", func(e *Encoder) { e.writeByte(formatInt8); e.writeByte(100) }, 100},
		{"int16 as uint", func(e *Encoder) { e.writeByte(formatInt16); e.writeUint16(10000) }, 10000},
		{"int32 as uint", func(e *Encoder) { e.writeByte(formatInt32); e.writeUint32(100000) }, 100000},
		{"int64 as uint", func(e *Encoder) { e.writeByte(formatInt64); e.writeUint64(500000) }, 500000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(16)
			tc.encode(e)
			var result uint64
			err := Unmarshal(e.Bytes(), &result)
			if err != nil || result != tc.want {
				t.Errorf("%s: got %d, want %d, err=%v", tc.name, result, tc.want, err)
			}
		})
	}
}

// TestUnmarshalBinarySlice tests decoding binary data into []byte.
func TestUnmarshalBinarySlice(t *testing.T) {
	tests := []struct {
		name   string
		encode func(*Encoder)
		want   int
	}{
		{"bin8", func(e *Encoder) { e.writeByte(formatBin8); e.writeByte(5); e.writeBytes([]byte{1, 2, 3, 4, 5}) }, 5},
		{"bin16", func(e *Encoder) { e.writeByte(formatBin16); e.writeUint16(3); e.writeBytes([]byte{1, 2, 3}) }, 3},
		{"bin32", func(e *Encoder) { e.writeByte(formatBin32); e.writeUint32(2); e.writeBytes([]byte{1, 2}) }, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEncoder(64)
			tc.encode(e)
			var result []byte
			err := Unmarshal(e.Bytes(), &result)
			if err != nil || len(result) != tc.want {
				t.Errorf("%s: got len=%d, want %d, err=%v", tc.name, len(result), tc.want, err)
			}
		})
	}
}

// TestUnmarshalSliceWithIntFormats tests decoding arrays with various int element formats.
func TestUnmarshalSliceWithIntFormats(t *testing.T) {
	e := NewEncoder(256)
	e.writeByte(fixarrayPrefix | 8) // 8 elements
	e.writeByte(42)                 // positive fixint
	e.writeByte(0xe0)               // negative fixint (-32)
	e.writeByte(formatInt8)
	e.writeByte(0x80) // -128
	e.writeByte(formatInt16)
	e.writeUint16(0x8000) // -32768
	e.writeByte(formatUint8)
	e.writeByte(200)
	e.writeByte(formatUint16)
	e.writeUint16(40000)
	e.writeByte(formatInt32)
	e.writeUint32(100)
	e.writeByte(formatUint32)
	e.writeUint32(100000)

	var result []int64
	err := Unmarshal(e.Bytes(), &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(result) != 8 {
		t.Errorf("got len=%d, want 8", len(result))
	}
	if result[0] != 42 {
		t.Errorf("result[0]: got %d, want 42", result[0])
	}
	if result[1] != -32 {
		t.Errorf("result[1]: got %d, want -32", result[1])
	}
	if result[2] != -128 {
		t.Errorf("result[2]: got %d, want -128", result[2])
	}
}

// TestDecodeAnyValueFormatBranches exercises all format branches in decodeAnyValue.
func TestDecodeAnyValueFormatBranches(t *testing.T) {
	t.Run("str8 via decodeAnyStr", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatStr8)
		e.writeByte(5)
		e.writeBytes([]byte("hello"))
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != "hello" {
			t.Errorf("str8: got %v, err=%v", v, err)
		}
	})

	t.Run("str16 via decodeAnyStr", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatStr16)
		e.writeUint16(5)
		e.writeBytes([]byte("world"))
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != "world" {
			t.Errorf("str16: got %v, err=%v", v, err)
		}
	})

	t.Run("str32 via decodeAnyStr", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatStr32)
		e.writeUint32(3)
		e.writeBytes([]byte("abc"))
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != "abc" {
			t.Errorf("str32: got %v, err=%v", v, err)
		}
	})

	t.Run("bin8 via decodeAnyBin", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatBin8)
		e.writeByte(3)
		e.writeBytes([]byte{1, 2, 3})
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		b, ok := v.([]byte)
		if err != nil || !ok || len(b) != 3 {
			t.Errorf("bin8: got %v, err=%v", v, err)
		}
	})

	t.Run("bin16 via decodeAnyBin", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatBin16)
		e.writeUint16(2)
		e.writeBytes([]byte{4, 5})
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		b, ok := v.([]byte)
		if err != nil || !ok || len(b) != 2 {
			t.Errorf("bin16: got %v, err=%v", v, err)
		}
	})

	t.Run("bin32 via decodeAnyBin", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatBin32)
		e.writeUint32(1)
		e.writeBytes([]byte{9})
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		b, ok := v.([]byte)
		if err != nil || !ok || len(b) != 1 {
			t.Errorf("bin32: got %v, err=%v", v, err)
		}
	})

	t.Run("uint8 via decodeAnyUint", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatUint8)
		e.writeByte(200)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(200) {
			t.Errorf("uint8: got %v, err=%v", v, err)
		}
	})

	t.Run("uint16 via decodeAnyUint", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatUint16)
		e.writeUint16(40000)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(40000) {
			t.Errorf("uint16: got %v, err=%v", v, err)
		}
	})

	t.Run("uint32 via decodeAnyUint", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatUint32)
		e.writeUint32(3000000000)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(3000000000) {
			t.Errorf("uint32: got %v, err=%v", v, err)
		}
	})

	t.Run("uint64 within int64 range via decodeAnyUint", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatUint64)
		e.writeUint64(5000000000)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(5000000000) {
			t.Errorf("uint64: got %v, err=%v", v, err)
		}
	})

	t.Run("uint64 overflow returns uint64 via decodeAnyUint", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatUint64)
		e.writeUint64(0xFFFFFFFFFFFFFFFF)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != uint64(0xFFFFFFFFFFFFFFFF) {
			t.Errorf("uint64 overflow: got %v (type %T), err=%v", v, v, err)
		}
	})

	t.Run("int8 via decodeAnyInt", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatInt8)
		e.writeByte(0x80) // -128
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(-128) {
			t.Errorf("int8: got %v, err=%v", v, err)
		}
	})

	t.Run("int16 via decodeAnyInt", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatInt16)
		e.writeUint16(0x8000) // -32768
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(-32768) {
			t.Errorf("int16: got %v, err=%v", v, err)
		}
	})

	t.Run("int32 via decodeAnyInt", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatInt32)
		e.writeUint32(0xFFFFFFFF) // -1
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(-1) {
			t.Errorf("int32: got %v, err=%v", v, err)
		}
	})

	t.Run("int64 via decodeAnyInt", func(t *testing.T) {
		e := NewEncoder(16)
		e.writeByte(formatInt64)
		e.writeUint64(0xFFFFFFFFFFFFFFFE) // -2
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != int64(-2) {
			t.Errorf("int64: got %v, err=%v", v, err)
		}
	})

	t.Run("array16 via decodeAnyArray", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatArray16)
		e.writeUint16(2)
		e.writeByte(1)
		e.writeByte(2)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		arr, ok := v.([]any)
		if err != nil || !ok || len(arr) != 2 {
			t.Errorf("array16: got %v, err=%v", v, err)
		}
	})

	t.Run("array32 via decodeAnyArray", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatArray32)
		e.writeUint32(3)
		e.writeByte(1)
		e.writeByte(2)
		e.writeByte(3)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		arr, ok := v.([]any)
		if err != nil || !ok || len(arr) != 3 {
			t.Errorf("array32: got %v, err=%v", v, err)
		}
	})

	t.Run("map16 via decodeAnyMap", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatMap16)
		e.writeUint16(1)
		e.writeByte(fixstrPrefix | 1) // key "k"
		e.writeByte('k')
		e.writeByte(42)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		m, ok := v.(map[string]any)
		if err != nil || !ok || m["k"] != int64(42) {
			t.Errorf("map16: got %v, err=%v", v, err)
		}
	})

	t.Run("map32 via decodeAnyMap", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatMap32)
		e.writeUint32(1)
		e.writeByte(fixstrPrefix | 1) // key "x"
		e.writeByte('x')
		e.writeByte(100)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		m, ok := v.(map[string]any)
		if err != nil || !ok || m["x"] != int64(100) {
			t.Errorf("map32: got %v, err=%v", v, err)
		}
	})

	t.Run("float32 via decodeAnyValue", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat32(3.14)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		f, ok := v.(float64)
		if err != nil || !ok {
			t.Errorf("float32: got %v, err=%v", v, err)
		}
		if f < 3.13 || f > 3.15 {
			t.Errorf("float32: got %v, want ~3.14", f)
		}
	})

	t.Run("float64 via decodeAnyValue", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat64(2.71828)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		f, ok := v.(float64)
		if err != nil || !ok {
			t.Errorf("float64: got %v, err=%v", v, err)
		}
		if f < 2.71 || f > 2.72 {
			t.Errorf("float64: got %v, want ~2.718", f)
		}
	})

	t.Run("nil via decodeAnyValue", func(t *testing.T) {
		e := NewEncoder(8)
		e.writeByte(formatNil)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != nil {
			t.Errorf("nil: got %v, err=%v", v, err)
		}
	})

	t.Run("true via decodeAnyValue", func(t *testing.T) {
		e := NewEncoder(8)
		e.writeByte(formatTrue)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != true {
			t.Errorf("true: got %v, err=%v", v, err)
		}
	})

	t.Run("false via decodeAnyValue", func(t *testing.T) {
		e := NewEncoder(8)
		e.writeByte(formatFalse)
		d := NewDecoder(e.Bytes())
		v, err := d.DecodeAny()
		if err != nil || v != false {
			t.Errorf("false: got %v, err=%v", v, err)
		}
	})
}

// TestDecodeMapStringString tests decoding map[string]string with various formats.
func TestDecodeMapStringString(t *testing.T) {
	t.Run("fixmap", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(fixmapPrefix | 2)
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('a')
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('x')
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('b')
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('y')

		var result map[string]string
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || result["a"] != "x" || result["b"] != "y" {
			t.Errorf("got %v, err=%v", result, err)
		}
	})

	t.Run("nil map", func(t *testing.T) {
		e := NewEncoder(8)
		e.writeByte(formatNil)
		var result map[string]string
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || result != nil {
			t.Errorf("got %v, err=%v", result, err)
		}
	})

	t.Run("nil value in map returns error", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(fixmapPrefix | 1)
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('k')
		e.writeByte(formatNil)

		var result map[string]string
		err := Unmarshal(e.Bytes(), &result)
		// map[string]string expects string values, nil returns type mismatch
		if err != ErrTypeMismatch {
			t.Errorf("expected ErrTypeMismatch, got err=%v", err)
		}
	})
}

// TestDecodeFloatFormats tests decoding float formats into float64.
func TestDecodeFloatFormats(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat32(3.14)
		var result float64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil {
			t.Errorf("err=%v", err)
		}
		if result < 3.13 || result > 3.15 {
			t.Errorf("got %v, want ~3.14", result)
		}
	})

	t.Run("float64", func(t *testing.T) {
		e := NewEncoder(16)
		e.EncodeFloat64(2.71828)
		var result float64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil {
			t.Errorf("err=%v", err)
		}
		if result < 2.71 || result > 2.72 {
			t.Errorf("got %v, want ~2.718", result)
		}
	})
}

// TestLargeArrayMapHeaders tests encoding/decoding of array16, array32, map16, map32.
func TestLargeArrayMapHeaders(t *testing.T) {
	t.Run("array16 header", func(t *testing.T) {
		e := NewEncoder(32)
		e.EncodeArrayHeader(300)
		for i := 0; i < 300; i++ {
			e.EncodeInt(int64(i))
		}

		var result []int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 300 {
			t.Errorf(errMsgGotLenErr, len(result), err)
		}
	})

	t.Run("array32 header", func(t *testing.T) {
		e := NewEncoder(8)
		// Directly write array32 format
		e.writeByte(formatArray32)
		e.writeUint32(3)
		e.writeByte(1)
		e.writeByte(2)
		e.writeByte(3)

		var result []int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 3 {
			t.Errorf(errMsgGotLenErr, len(result), err)
		}
	})

	t.Run("map16 header", func(t *testing.T) {
		e := NewEncoder(1024)
		e.EncodeMapHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("k" + string(rune('a'+i)))
			e.EncodeInt(int64(i))
		}

		var result map[string]int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 20 {
			t.Errorf(errMsgGotLenErr, len(result), err)
		}
	})

	t.Run("map32 header", func(t *testing.T) {
		e := NewEncoder(64)
		e.writeByte(formatMap32)
		e.writeUint32(2)
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('a')
		e.writeByte(1)
		e.writeByte(fixstrPrefix | 1)
		e.writeByte('b')
		e.writeByte(2)

		var result map[string]int64
		err := Unmarshal(e.Bytes(), &result)
		if err != nil || len(result) != 2 {
			t.Errorf(errMsgGotLenErr, len(result), err)
		}
	})
}
