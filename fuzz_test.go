package msgpck

import (
	"bytes"
	"testing"
)

// FuzzDecoder tests the decoder with random input
func FuzzDecoder(f *testing.F) {
	// Add seed corpus with valid msgpack data
	seeds := [][]byte{
		{0xc0},                         // nil
		{0xc2},                         // false
		{0xc3},                         // true
		{0x00},                         // positive fixint 0
		{0x7f},                         // positive fixint 127
		{0xe0},                         // negative fixint -32
		{0xff},                         // negative fixint -1
		{0xcc, 0xff},                   // uint8
		{0xcd, 0xff, 0xff},             // uint16
		{0xce, 0xff, 0xff, 0xff, 0xff}, // uint32
		{0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // uint64
		{0xd0, 0x80},                   // int8 -128
		{0xd1, 0x80, 0x00},             // int16
		{0xd2, 0x80, 0x00, 0x00, 0x00}, // int32
		{0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // int64
		{0xca, 0x40, 0x48, 0xf5, 0xc3},                         // float32 3.14
		{0xcb, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18}, // float64
		{0xa5, 'h', 'e', 'l', 'l', 'o'},                        // fixstr
		{0xd9, 0x05, 'h', 'e', 'l', 'l', 'o'},                  // str8
		{0xc4, 0x03, 0x01, 0x02, 0x03},                         // bin8
		{0x93, 0x01, 0x02, 0x03},                               // fixarray [1,2,3]
		{0x82, 0xa1, 'a', 0x01, 0xa1, 'b', 0x02},               // fixmap {"a":1,"b":2}
		{0xd4, 0x01, 0xff},                                     // fixext1
		{0xd5, 0x01, 0xff, 0xff},                               // fixext2
		{0xd6, 0x01, 0xff, 0xff, 0xff, 0xff},                   // fixext4
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test that decoder doesn't panic on any input
		d := NewDecoder(data)
		_, _ = d.Decode() // Errors expected for invalid input

		// Test DecodeAny
		d.Reset(data)
		_, _ = d.DecodeAny() // Errors expected for invalid input

		// Test with strict limits
		cfg := Config{
			MaxStringLen: 100,
			MaxBinaryLen: 100,
			MaxArrayLen:  10,
			MaxMapLen:    10,
			MaxExtLen:    100,
			MaxDepth:     5,
		}
		d2 := NewDecoderWithConfig(data, cfg)
		_, _ = d2.Decode() // Errors expected for invalid input
	})
}

// FuzzRoundTrip tests encode/decode round-trip for various types
func FuzzRoundTrip(f *testing.F) {
	// Add seed values
	f.Add([]byte("hello"))
	f.Add([]byte(""))
	f.Add([]byte{0, 1, 2, 255})
	f.Add([]byte("unicode: 日本語"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test string round-trip
		encoded, err := MarshalCopy(string(data))
		if err != nil {
			return
		}
		decoded, err := Unmarshal(encoded)
		if err != nil {
			t.Fatalf("failed to decode string: %v", err)
		}
		if s, ok := decoded.(string); !ok || s != string(data) {
			t.Fatalf("string round-trip failed: got %v, want %v", decoded, string(data))
		}

		// Test []byte round-trip
		encoded2, err := MarshalCopy(data)
		if err != nil {
			return
		}
		decoded2, err := Unmarshal(encoded2)
		if err != nil {
			t.Fatalf("failed to decode bytes: %v", err)
		}
		if b, ok := decoded2.([]byte); !ok || !bytes.Equal(b, data) {
			t.Fatalf("bytes round-trip failed")
		}
	})
}

// FuzzRoundTripInt tests integer round-trip
func FuzzRoundTripInt(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(127))
	f.Add(int64(-128))
	f.Add(int64(32767))
	f.Add(int64(-32768))
	f.Add(int64(2147483647))
	f.Add(int64(-2147483648))
	f.Add(int64(9223372036854775807))
	f.Add(int64(-9223372036854775808))

	f.Fuzz(func(t *testing.T, val int64) {
		encoded, err := MarshalCopy(val)
		if err != nil {
			t.Fatalf("failed to encode int64: %v", err)
		}
		decoded, err := Unmarshal(encoded)
		if err != nil {
			t.Fatalf("failed to decode int64: %v", err)
		}
		if v, ok := decoded.(int64); !ok || v != val {
			t.Fatalf("int64 round-trip failed: got %v (%T), want %v", decoded, decoded, val)
		}
	})
}

// FuzzRoundTripFloat tests float round-trip
func FuzzRoundTripFloat(f *testing.F) {
	f.Add(float64(0))
	f.Add(float64(3.14159))
	f.Add(float64(-3.14159))
	f.Add(float64(1e308))
	f.Add(float64(-1e308))

	f.Fuzz(func(t *testing.T, val float64) {
		encoded, err := MarshalCopy(val)
		if err != nil {
			t.Fatalf("failed to encode float64: %v", err)
		}
		decoded, err := Unmarshal(encoded)
		if err != nil {
			t.Fatalf("failed to decode float64: %v", err)
		}
		if v, ok := decoded.(float64); !ok || v != val {
			// NaN special case
			if val != val && v != v {
				return // both NaN, ok
			}
			t.Fatalf("float64 round-trip failed: got %v, want %v", decoded, val)
		}
	})
}

// FuzzMapDecode tests map decoding with random input
func FuzzMapDecode(f *testing.F) {
	// Seed with valid map data
	f.Add([]byte{0x80})                                                          // empty fixmap
	f.Add([]byte{0x82, 0xa1, 'a', 0x01, 0xa1, 'b', 0x02})                        // fixmap {"a":1,"b":2}
	f.Add([]byte{0x81, 0xa4, 'n', 'a', 'm', 'e', 0xa5, 'A', 'l', 'i', 'c', 'e'}) // {"name":"Alice"}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test UnmarshalMap - shouldn't panic
		_, _ = UnmarshalMapStringAny(data, false)

		// Test UnmarshalMapZeroCopy - shouldn't panic
		_, _ = UnmarshalMapStringAny(data, true)

		// Test DecodeMapFunc - shouldn't panic
		_ = DecodeMapFunc(data, func(m map[string]any) error {
			// Access some fields to exercise the map
			_ = m["test"]
			_ = m["name"]
			for k, v := range m {
				_ = k
				_ = v
			}
			return nil
		})

		// Test UnmarshalMapStringAny with both modes
		_, _ = UnmarshalMapStringAny(data, true)
		_, _ = UnmarshalMapStringAny(data, false)
	})
}

// FuzzMapStringString tests typed map decoding
func FuzzMapStringString(f *testing.F) {
	// Seed with valid string->string map data
	f.Add([]byte{0x80})                                             // empty map
	f.Add([]byte{0x81, 0xa1, 'k', 0xa1, 'v'})                       // {"k":"v"}
	f.Add([]byte{0x82, 0xa1, 'a', 0xa1, 'x', 0xa1, 'b', 0xa1, 'y'}) // {"a":"x","b":"y"}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test UnmarshalMapStringString - shouldn't panic
		_, _ = UnmarshalMapStringString(data, false)
		_, _ = UnmarshalMapStringString(data, true)

		// Test DecodeStringMapFunc - shouldn't panic
		_ = DecodeStringMapFunc(data, func(m map[string]string) error {
			for k, v := range m {
				_ = k
				_ = v
			}
			return nil
		})
	})
}

// FuzzStructDecode tests reflection-based struct decoding
func FuzzStructDecode(f *testing.F) {
	type TestStruct struct {
		Name  string `msgpack:"name"`
		Value int    `msgpack:"value"`
	}

	// Encode a valid struct as seed
	enc := newEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("name")
	enc.EncodeString("test")
	enc.EncodeString("value")
	enc.EncodeInt(42)
	f.Add(enc.Bytes())

	// Empty map
	f.Add([]byte{0x80})

	f.Fuzz(func(t *testing.T, data []byte) {
		var s TestStruct
		_ = UnmarshalStruct(data, &s) // Shouldn't panic, errors expected
	})
}

// FuzzCachedStructDecoder tests the cached struct decoder
func FuzzCachedStructDecoder(f *testing.F) {
	type Person struct {
		Name string `msgpack:"name"`
		Age  int    `msgpack:"age"`
	}

	// Valid person encoding
	enc := newEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("name")
	enc.EncodeString("Alice")
	enc.EncodeString("age")
	enc.EncodeInt(30)
	f.Add(enc.Bytes())

	// Empty map
	f.Add([]byte{0x80})
	// Nil
	f.Add([]byte{0xc0})

	dec := GetStructDecoder[Person](false)
	decZC := GetStructDecoder[Person](true)

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test regular cached decoder - shouldn't panic
		var p1 Person
		_ = dec.Decode(data, &p1)

		// Test zero-copy cached decoder - shouldn't panic
		var p2 Person
		_ = decZC.Decode(data, &p2)
	})
}

// FuzzDecodeStructFunc tests the callback-based struct decoder
func FuzzDecodeStructFunc(f *testing.F) {
	type Data struct {
		ID    int64  `msgpack:"id"`
		Value string `msgpack:"value"`
	}

	// Valid encoding
	enc := newEncoder(64)
	enc.EncodeMapHeader(2)
	enc.EncodeString("id")
	enc.EncodeInt(123)
	enc.EncodeString("value")
	enc.EncodeString("test")
	f.Add(enc.Bytes())

	f.Add([]byte{0x80}) // empty map
	f.Add([]byte{0xc0}) // nil

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test DecodeStructFunc - shouldn't panic
		_ = DecodeStructFunc(data, func(d *Data) error {
			_ = d.ID
			_ = d.Value
			return nil
		})
	})
}

// FuzzStructEncoder tests the cached struct encoder
func FuzzStructEncoder(f *testing.F) {
	f.Add("Alice", int64(30))
	f.Add("", int64(0))
	f.Add("Bob", int64(-1))
	f.Add("名前", int64(9223372036854775807))

	type Person struct {
		Name string `msgpack:"name"`
		Age  int64  `msgpack:"age"`
	}

	enc := GetStructEncoder[Person]()
	dec := GetStructDecoder[Person](false)

	f.Fuzz(func(t *testing.T, name string, age int64) {
		p := Person{Name: name, Age: age}

		// Encode with cached encoder
		data, err := enc.EncodeCopy(&p)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}

		// Decode back
		var p2 Person
		err = dec.Decode(data, &p2)
		if err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		// Verify round-trip
		if p2.Name != name || p2.Age != age {
			t.Fatalf("round-trip failed: got {%q, %d}, want {%q, %d}", p2.Name, p2.Age, name, age)
		}
	})
}

// FuzzNestedStructures tests deeply nested arrays and maps
func FuzzNestedStructures(f *testing.F) {
	// Nested array: [[1]]
	f.Add([]byte{0x91, 0x91, 0x01})
	// Nested map: {"a":{"b":1}}
	f.Add([]byte{0x81, 0xa1, 'a', 0x81, 0xa1, 'b', 0x01})
	// Deeper nesting
	f.Add([]byte{0x91, 0x91, 0x91, 0x91, 0x01})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test with default config
		_, _ = Unmarshal(data)

		// Test with strict depth limit
		cfg := Config{
			MaxStringLen: 1000,
			MaxBinaryLen: 1000,
			MaxArrayLen:  100,
			MaxMapLen:    100,
			MaxExtLen:    1000,
			MaxDepth:     3, // Very shallow
		}
		d := NewDecoderWithConfig(data, cfg)
		_, _ = d.DecodeAny()
	})
}

// FuzzLargeCollections tests handling of large array/map size claims
func FuzzLargeCollections(f *testing.F) {
	// array16 claiming 1000 elements but with no data
	f.Add([]byte{0xdc, 0x03, 0xe8})
	// map16 claiming 1000 pairs but with no data
	f.Add([]byte{0xde, 0x03, 0xe8})
	// array32 with huge size
	f.Add([]byte{0xdd, 0xff, 0xff, 0xff, 0xff})
	// map32 with huge size
	f.Add([]byte{0xdf, 0xff, 0xff, 0xff, 0xff})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should not panic or allocate excessive memory
		cfg := Config{
			MaxStringLen: 1000,
			MaxBinaryLen: 1000,
			MaxArrayLen:  100,
			MaxMapLen:    100,
			MaxExtLen:    1000,
			MaxDepth:     10,
		}
		d := NewDecoderWithConfig(data, cfg)
		_, _ = d.DecodeAny()

		// Also test map decoding
		_, _ = UnmarshalMapStringAny(data, false)
	})
}

// newEncoder is a helper that returns a new encoder (for test seeds)
func newEncoder(size int) *Encoder {
	return NewEncoder(size)
}
