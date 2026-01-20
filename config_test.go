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
	err := UnmarshalStruct(b, &d)
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
