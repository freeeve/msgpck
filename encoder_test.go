package msgpck

import (
	"testing"
)

func TestEncoderBufferGrowth(t *testing.T) {
	e := NewEncoder(1) // Start with tiny buffer

	// Write something larger than initial buffer
	bigString := string(make([]byte, 100))
	e.EncodeString(bigString)

	if len(e.Bytes()) == 0 {
		t.Error("encoder should have grown buffer")
	}
}

func TestEncoderStringFormats(t *testing.T) {
	e := NewEncoder(256)

	// fixstr (0-31 bytes)
	e.Reset()
	e.EncodeString("hi")
	if e.Bytes()[0]&0xe0 != 0xa0 {
		t.Error("short string should use fixstr")
	}

	// str8 (32-255 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 50)))
	if e.Bytes()[0] != formatStr8 {
		t.Error("medium string should use str8")
	}

	// str16 (256-65535 bytes)
	e.Reset()
	e.EncodeString(string(make([]byte, 300)))
	if e.Bytes()[0] != formatStr16 {
		t.Error("large string should use str16")
	}
}

func TestEncoderBinaryFormats(t *testing.T) {
	e := NewEncoder(256)

	// bin8 (0-255 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 50))
	if e.Bytes()[0] != formatBin8 {
		t.Error("small binary should use bin8")
	}

	// bin16 (256-65535 bytes)
	e.Reset()
	e.EncodeBinary(make([]byte, 300))
	if e.Bytes()[0] != formatBin16 {
		t.Error("medium binary should use bin16")
	}
}

func TestEncoderMapHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixmap (0-15)
	e.Reset()
	e.EncodeMapHeader(5)
	if e.Bytes()[0]&0xf0 != 0x80 {
		t.Error("small map should use fixmap")
	}

	// map16 (16-65535)
	e.Reset()
	e.EncodeMapHeader(20)
	if e.Bytes()[0] != formatMap16 {
		t.Error("medium map should use map16")
	}
}

func TestEncoderArrayHeader(t *testing.T) {
	e := NewEncoder(32)

	// fixarray (0-15)
	e.Reset()
	e.EncodeArrayHeader(5)
	if e.Bytes()[0]&0xf0 != 0x90 {
		t.Error("small array should use fixarray")
	}

	// array16 (16-65535)
	e.Reset()
	e.EncodeArrayHeader(20)
	if e.Bytes()[0] != formatArray16 {
		t.Error("medium array should use array16")
	}
}
