package msgpck

import (
	"testing"
)

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
