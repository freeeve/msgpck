package msgpck

import (
	"testing"
)

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
	err := Unmarshal(b, &d)
	if err != nil || d.Name != "test" || d.Value != 42 || d.Skip != 999 {
		t.Error("skip tagged field failed")
	}
}

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

// TestSecurityLimits tests that security limits are enforced
func TestSecurityLimits(t *testing.T) {
	t.Run("string too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxStringLen(10)
		// str8 with length 100
		data := []byte{formatStr8, 100}
		data = append(data, make([]byte, 100)...)

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrStringTooLong {
			t.Errorf("expected ErrStringTooLong, got %v", err)
		}
	})

	t.Run("array too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxArrayLen(5)
		// array16 with length 1000
		data := []byte{formatArray16, 0x03, 0xe8} // 1000

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrArrayTooLong {
			t.Errorf("expected ErrArrayTooLong, got %v", err)
		}
	})

	t.Run("map too long", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxMapLen(5)
		// map16 with length 1000
		data := []byte{formatMap16, 0x03, 0xe8}

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrMapTooLong {
			t.Errorf("expected ErrMapTooLong, got %v", err)
		}
	})

	t.Run("max depth exceeded", func(t *testing.T) {
		cfg := DefaultConfig().WithMaxDepth(3)
		// Deeply nested arrays: [[[[1]]]]
		data := []byte{
			0x91, // fixarray 1
			0x91, // fixarray 1
			0x91, // fixarray 1
			0x91, // fixarray 1 - this exceeds depth 3
			0x01, // fixint 1
		}

		d := NewDecoderWithConfig(data, cfg)
		_, err := d.Decode()
		if err != ErrMaxDepthExceeded {
			t.Errorf("expected ErrMaxDepthExceeded, got %v", err)
		}
	})
}
