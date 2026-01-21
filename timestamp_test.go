package msgpck

import (
	"testing"
	"time"
)

func TestTimestampRoundTrip(t *testing.T) {
	testCases := []struct {
		name string
		ts   time.Time
	}{
		{"unix_epoch", time.Unix(0, 0).UTC()},
		{"recent", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		{"with_nanos", time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)},
		{"max_timestamp32", time.Unix(0xFFFFFFFF, 0).UTC()},
		{"timestamp64_range", time.Unix(0x100000000, 0).UTC()},
		{"timestamp64_nanos", time.Unix(0x100000000, 500000000).UTC()},
		{"negative_unix", time.Unix(-1, 0).UTC()},
		{"far_past", time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"far_future", time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := MarshalTimestamp(tc.ts)
			decoded, err := UnmarshalTimestamp(data)
			if err != nil {
				t.Fatalf(errMsgDecodeFailed, err)
			}
			if !decoded.Equal(tc.ts) {
				t.Errorf(errMsgGotWant, decoded, tc.ts)
			}
			if decoded.Location() != time.UTC {
				t.Errorf("expected UTC, got %v", decoded.Location())
			}
		})
	}
}

func TestTimestampEncoderDecoder(t *testing.T) {
	ts := time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC)

	enc := NewEncoder(64)
	enc.EncodeTimestamp(ts)
	data := enc.Bytes()

	dec := NewDecoder(data)
	decoded, err := dec.DecodeTimestamp()
	if err != nil {
		t.Fatalf(errMsgDecodeFailed, err)
	}
	if !decoded.Equal(ts) {
		t.Errorf(errMsgGotWant, decoded, ts)
	}
}

func TestTimestampFormats(t *testing.T) {
	t.Run("timestamp32", func(t *testing.T) {
		ts := time.Unix(1000000, 0).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatFixExt4 {
			t.Errorf("expected formatFixExt4, got 0x%x", data[0])
		}
		if data[1] != 0xff {
			t.Errorf("expected type -1, got 0x%x", data[1])
		}
	})

	t.Run("timestamp64", func(t *testing.T) {
		ts := time.Unix(1000000, 500000000).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatFixExt8 {
			t.Errorf("expected formatFixExt8, got 0x%x", data[0])
		}
	})

	t.Run("timestamp96", func(t *testing.T) {
		ts := time.Unix(-1000000, 500000000).UTC()
		data := MarshalTimestamp(ts)
		if data[0] != formatExt8 {
			t.Errorf("expected formatExt8, got 0x%x", data[0])
		}
		if data[1] != 12 {
			t.Errorf("expected length 12, got %d", data[1])
		}
	})
}

func TestTimestampWithTimezone(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("timezone not available")
	}

	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, loc)
	data := MarshalTimestamp(ts)
	decoded, err := UnmarshalTimestamp(data)
	if err != nil {
		t.Fatalf(errMsgDecodeFailed, err)
	}

	if !decoded.Equal(ts) {
		t.Errorf("times not equal: got %v, want %v", decoded, ts)
	}
	if decoded.Location() != time.UTC {
		t.Errorf("expected UTC, got %v", decoded.Location())
	}
}

func TestIsTimestamp(t *testing.T) {
	if !IsTimestamp(Ext{Type: -1, Data: []byte{1, 2, 3, 4}}) {
		t.Error("expected true for type -1")
	}
	if IsTimestamp(Ext{Type: 1, Data: []byte{1, 2, 3, 4}}) {
		t.Error("expected false for type 1")
	}
}

func TestExtToTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	data := MarshalTimestamp(ts)

	dec := NewDecoder(data)
	v, err := dec.Decode()
	if err != nil {
		t.Fatalf(errMsgDecodeFailed, err)
	}
	if v.Type != TypeExt {
		t.Fatalf("expected ext type, got %v", v.Type)
	}

	decoded, err := ExtToTimestamp(v.Ext)
	if err != nil {
		t.Fatalf("ExtToTimestamp failed: %v", err)
	}
	if !decoded.Equal(ts) {
		t.Errorf(errMsgGotWant, decoded, ts)
	}
}

func TestTimestampDecodeErrors(t *testing.T) {
	t.Run("wrong_format", func(t *testing.T) {
		data := []byte{formatNil}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf(errMsgTypeMismatch, err)
		}
	})

	t.Run("wrong_ext_type", func(t *testing.T) {
		data := []byte{formatFixExt4, 0x01, 0, 0, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf(errMsgTypeMismatch, err)
		}
	})

	t.Run("wrong_ext8_length", func(t *testing.T) {
		data := []byte{formatExt8, 5, 0xff, 0, 0, 0, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrTypeMismatch {
			t.Errorf(errMsgTypeMismatch, err)
		}
	})

	t.Run("truncated_data", func(t *testing.T) {
		data := []byte{formatFixExt4, 0xff, 0, 0}
		_, err := UnmarshalTimestamp(data)
		if err != ErrUnexpectedEOF {
			t.Errorf(errMsgUnexpectedEOF, err)
		}
	})
}

func TestExtToTimestampErrors(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		ext := Ext{Type: 1, Data: []byte{0, 0, 0, 0}}
		_, err := ExtToTimestamp(ext)
		if err != ErrTypeMismatch {
			t.Errorf(errMsgTypeMismatch, err)
		}
	})

	t.Run("wrong_length", func(t *testing.T) {
		ext := Ext{Type: -1, Data: []byte{0, 0, 0}}
		_, err := ExtToTimestamp(ext)
		if err != ErrTypeMismatch {
			t.Errorf(errMsgTypeMismatch, err)
		}
	})
}
