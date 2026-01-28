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
