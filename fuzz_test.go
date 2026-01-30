package msgpck

import (
	"bytes"
	"math"
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
		encoded, err := Marshal(string(data))
		if err != nil {
			return
		}
		var decoded any
		err = Unmarshal(encoded, &decoded)
		if err != nil {
			t.Fatalf("failed to decode string: %v", err)
		}
		if s, ok := decoded.(string); !ok || s != string(data) {
			t.Fatalf("string round-trip failed: got %v, want %v", decoded, string(data))
		}

		// Test []byte round-trip
		encoded2, err := Marshal(data)
		if err != nil {
			return
		}
		var decoded2 any
		err = Unmarshal(encoded2, &decoded2)
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
		encoded, err := Marshal(val)
		if err != nil {
			t.Fatalf("failed to encode int64: %v", err)
		}
		var decoded any
		err = Unmarshal(encoded, &decoded)
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
		encoded, err := Marshal(val)
		if err != nil {
			t.Fatalf("failed to encode float64: %v", err)
		}
		var decoded any
		err = Unmarshal(encoded, &decoded)
		if err != nil {
			t.Fatalf("failed to decode float64: %v", err)
		}
		if v, ok := decoded.(float64); !ok || v != val {
			// NaN special case
			if math.IsNaN(val) && math.IsNaN(v) {
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
		// Test Unmarshal into map[string]any - shouldn't panic
		var m map[string]any
		_ = Unmarshal(data, &m)

		// Access some fields to exercise the map
		if m != nil {
			_ = m["test"]
			_ = m["name"]
			for k, v := range m {
				_ = k
				_ = v
			}
		}
	})
}

// FuzzMapStringString tests typed map decoding
func FuzzMapStringString(f *testing.F) {
	// Seed with valid string->string map data
	f.Add([]byte{0x80})                                             // empty map
	f.Add([]byte{0x81, 0xa1, 'k', 0xa1, 'v'})                       // {"k":"v"}
	f.Add([]byte{0x82, 0xa1, 'a', 0xa1, 'x', 0xa1, 'b', 0xa1, 'y'}) // {"a":"x","b":"y"}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Test Unmarshal into map[string]string - shouldn't panic
		var m map[string]string
		_ = Unmarshal(data, &m)

		// Iterate if not nil
		if m != nil {
			for k, v := range m {
				_ = k
				_ = v
			}
		}
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
		_ = Unmarshal(data, &s) // Shouldn't panic, errors expected
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

// FuzzStructUnmarshal tests the struct decoding
func FuzzStructUnmarshal(f *testing.F) {
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
		// Test Unmarshal into struct - shouldn't panic
		var d Data
		_ = Unmarshal(data, &d)
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
		data, err := enc.Encode(&p)
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
		var result any
		_ = Unmarshal(data, &result)

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
		var m map[string]any
		_ = Unmarshal(data, &m)
	})
}

// newEncoder is a helper that returns a new encoder (for test seeds)
func newEncoder(size int) *Encoder {
	return NewEncoder(size)
}

// FuzzStructRoundtrip tests encode/decode roundtrip for structs with all supported field types.
// This fuzzer uses the fuzz input to populate a struct with various field types including pointers,
// encodes it, decodes it back, and verifies all fields match.
func FuzzStructRoundtrip(f *testing.F) {
	// Seed with various byte patterns
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{255, 255, 255, 255, 255, 255, 255, 255})
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	f.Add([]byte("hello world test string"))
	f.Add([]byte{0x80, 0x00, 0x00, 0x00, 0x7f, 0xff, 0xff, 0xff}) // edge cases

	// AllFieldTypes contains all supported field types for roundtrip testing
	type AllFieldTypes struct {
		// Basic types
		String  string  `msgpack:"string"`
		Int     int     `msgpack:"int"`
		Int64   int64   `msgpack:"int64"`
		Int32   int32   `msgpack:"int32"`
		Int16   int16   `msgpack:"int16"`
		Int8    int8    `msgpack:"int8"`
		Uint    uint    `msgpack:"uint"`
		Uint64  uint64  `msgpack:"uint64"`
		Uint32  uint32  `msgpack:"uint32"`
		Uint16  uint16  `msgpack:"uint16"`
		Uint8   uint8   `msgpack:"uint8"`
		Float64 float64 `msgpack:"float64"`
		Float32 float32 `msgpack:"float32"`
		Bool    bool    `msgpack:"bool"`
		Bytes   []byte  `msgpack:"bytes"`

		// Slices
		StringSlice  []string  `msgpack:"string_slice"`
		Int64Slice   []int64   `msgpack:"int64_slice"`
		Float64Slice []float64 `msgpack:"float64_slice"`

		// Maps
		StringMap  map[string]string  `msgpack:"string_map"`
		IntMap     map[string]int     `msgpack:"int_map"`
		Float64Map map[string]float64 `msgpack:"float64_map"`

		// Pointer types
		PtrString  *string  `msgpack:"ptr_string"`
		PtrInt     *int     `msgpack:"ptr_int"`
		PtrInt64   *int64   `msgpack:"ptr_int64"`
		PtrUint64  *uint64  `msgpack:"ptr_uint64"`
		PtrFloat64 *float64 `msgpack:"ptr_float64"`
		PtrBool    *bool    `msgpack:"ptr_bool"`
	}

	enc := GetStructEncoder[AllFieldTypes]()
	dec := GetStructDecoder[AllFieldTypes](false)

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 8 {
			return // Need at least some bytes to work with
		}

		// Use fuzz data to populate struct fields deterministically
		original := AllFieldTypes{}

		// String from fuzz data
		if len(data) > 0 {
			strLen := int(data[0]) % (len(data) + 1)
			if strLen > len(data) {
				strLen = len(data)
			}
			original.String = string(data[:strLen])
		}

		// Integer types from fuzz bytes
		if len(data) >= 8 {
			original.Int64 = int64(data[0]) | int64(data[1])<<8 | int64(data[2])<<16 | int64(data[3])<<24 |
				int64(data[4])<<32 | int64(data[5])<<40 | int64(data[6])<<48 | int64(data[7])<<56
			original.Int = int(original.Int64)
			original.Int32 = int32(original.Int64)
			original.Int16 = int16(original.Int64)
			original.Int8 = int8(original.Int64)
			original.Uint64 = uint64(original.Int64)
			original.Uint = uint(original.Uint64)
			original.Uint32 = uint32(original.Uint64)
			original.Uint16 = uint16(original.Uint64)
			original.Uint8 = uint8(original.Uint64)
		}

		// Float types
		if len(data) >= 8 {
			bits := uint64(data[0]) | uint64(data[1])<<8 | uint64(data[2])<<16 | uint64(data[3])<<24 |
				uint64(data[4])<<32 | uint64(data[5])<<40 | uint64(data[6])<<48 | uint64(data[7])<<56
			original.Float64 = math.Float64frombits(bits)
			original.Float32 = float32(original.Float64)
			// Avoid NaN for comparison simplicity
			if math.IsNaN(original.Float64) {
				original.Float64 = 0
				original.Float32 = 0
			}
		}

		// Bool
		if len(data) > 0 {
			original.Bool = data[0]%2 == 1
		}

		// Bytes slice
		original.Bytes = make([]byte, len(data))
		copy(original.Bytes, data)

		// String slice
		if len(data) > 2 {
			count := int(data[0]) % 5 // 0-4 strings
			original.StringSlice = make([]string, count)
			for i := 0; i < count && i < len(data); i++ {
				original.StringSlice[i] = string(data[i : i+1])
			}
		}

		// Int64 slice
		if len(data) >= 8 {
			count := int(data[0]) % 4 // 0-3 ints
			original.Int64Slice = make([]int64, count)
			for i := 0; i < count; i++ {
				original.Int64Slice[i] = int64(data[i%len(data)])
			}
		}

		// Float64 slice
		if len(data) >= 8 {
			count := int(data[0]) % 3 // 0-2 floats
			original.Float64Slice = make([]float64, count)
			for i := 0; i < count; i++ {
				original.Float64Slice[i] = float64(data[i%len(data)])
			}
		}

		// String map
		if len(data) > 2 {
			count := int(data[0]) % 3 // 0-2 entries
			if count > 0 {
				original.StringMap = make(map[string]string, count)
				for i := 0; i < count && i < len(data); i++ {
					key := string([]byte{'k', byte('0' + i)})
					original.StringMap[key] = string(data[i : i+1])
				}
			}
		}

		// Int map
		if len(data) > 2 {
			count := int(data[1]) % 3
			if count > 0 {
				original.IntMap = make(map[string]int, count)
				for i := 0; i < count && i < len(data); i++ {
					key := string([]byte{'i', byte('0' + i)})
					original.IntMap[key] = int(data[i])
				}
			}
		}

		// Float64 map
		if len(data) > 2 {
			count := int(data[2%len(data)]) % 2
			if count > 0 {
				original.Float64Map = make(map[string]float64, count)
				for i := 0; i < count && i < len(data); i++ {
					key := string([]byte{'f', byte('0' + i)})
					original.Float64Map[key] = float64(data[i])
				}
			}
		}

		// Pointer fields - set based on fuzz data bit pattern
		if len(data) > 0 && data[0]&0x01 != 0 {
			s := original.String
			original.PtrString = &s
		}
		if len(data) > 0 && data[0]&0x02 != 0 {
			i := original.Int
			original.PtrInt = &i
		}
		if len(data) > 0 && data[0]&0x04 != 0 {
			i64 := original.Int64
			original.PtrInt64 = &i64
		}
		if len(data) > 0 && data[0]&0x08 != 0 {
			u64 := original.Uint64
			original.PtrUint64 = &u64
		}
		if len(data) > 0 && data[0]&0x10 != 0 {
			f64 := original.Float64
			original.PtrFloat64 = &f64
		}
		if len(data) > 0 && data[0]&0x20 != 0 {
			b := original.Bool
			original.PtrBool = &b
		}

		// Encode
		encoded, err := enc.Encode(&original)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}

		// Decode
		var decoded AllFieldTypes
		err = dec.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}

		// Verify all fields match
		if decoded.String != original.String {
			t.Errorf("String mismatch: got %q, want %q", decoded.String, original.String)
		}
		if decoded.Int != original.Int {
			t.Errorf("Int mismatch: got %d, want %d", decoded.Int, original.Int)
		}
		if decoded.Int64 != original.Int64 {
			t.Errorf("Int64 mismatch: got %d, want %d", decoded.Int64, original.Int64)
		}
		if decoded.Int32 != original.Int32 {
			t.Errorf("Int32 mismatch: got %d, want %d", decoded.Int32, original.Int32)
		}
		if decoded.Int16 != original.Int16 {
			t.Errorf("Int16 mismatch: got %d, want %d", decoded.Int16, original.Int16)
		}
		if decoded.Int8 != original.Int8 {
			t.Errorf("Int8 mismatch: got %d, want %d", decoded.Int8, original.Int8)
		}
		if decoded.Uint != original.Uint {
			t.Errorf("Uint mismatch: got %d, want %d", decoded.Uint, original.Uint)
		}
		if decoded.Uint64 != original.Uint64 {
			t.Errorf("Uint64 mismatch: got %d, want %d", decoded.Uint64, original.Uint64)
		}
		if decoded.Uint32 != original.Uint32 {
			t.Errorf("Uint32 mismatch: got %d, want %d", decoded.Uint32, original.Uint32)
		}
		if decoded.Uint16 != original.Uint16 {
			t.Errorf("Uint16 mismatch: got %d, want %d", decoded.Uint16, original.Uint16)
		}
		if decoded.Uint8 != original.Uint8 {
			t.Errorf("Uint8 mismatch: got %d, want %d", decoded.Uint8, original.Uint8)
		}
		if decoded.Float64 != original.Float64 {
			t.Errorf("Float64 mismatch: got %v, want %v", decoded.Float64, original.Float64)
		}
		if decoded.Float32 != original.Float32 {
			t.Errorf("Float32 mismatch: got %v, want %v", decoded.Float32, original.Float32)
		}
		if decoded.Bool != original.Bool {
			t.Errorf("Bool mismatch: got %v, want %v", decoded.Bool, original.Bool)
		}
		if !bytes.Equal(decoded.Bytes, original.Bytes) {
			t.Errorf("Bytes mismatch: got %v, want %v", decoded.Bytes, original.Bytes)
		}

		// Verify slices
		if len(decoded.StringSlice) != len(original.StringSlice) {
			t.Errorf("StringSlice length mismatch: got %d, want %d", len(decoded.StringSlice), len(original.StringSlice))
		} else {
			for i := range original.StringSlice {
				if decoded.StringSlice[i] != original.StringSlice[i] {
					t.Errorf("StringSlice[%d] mismatch: got %q, want %q", i, decoded.StringSlice[i], original.StringSlice[i])
				}
			}
		}

		if len(decoded.Int64Slice) != len(original.Int64Slice) {
			t.Errorf("Int64Slice length mismatch: got %d, want %d", len(decoded.Int64Slice), len(original.Int64Slice))
		} else {
			for i := range original.Int64Slice {
				if decoded.Int64Slice[i] != original.Int64Slice[i] {
					t.Errorf("Int64Slice[%d] mismatch: got %d, want %d", i, decoded.Int64Slice[i], original.Int64Slice[i])
				}
			}
		}

		if len(decoded.Float64Slice) != len(original.Float64Slice) {
			t.Errorf("Float64Slice length mismatch: got %d, want %d", len(decoded.Float64Slice), len(original.Float64Slice))
		} else {
			for i := range original.Float64Slice {
				if decoded.Float64Slice[i] != original.Float64Slice[i] {
					t.Errorf("Float64Slice[%d] mismatch: got %v, want %v", i, decoded.Float64Slice[i], original.Float64Slice[i])
				}
			}
		}

		// Verify maps
		if len(decoded.StringMap) != len(original.StringMap) {
			t.Errorf("StringMap length mismatch: got %d, want %d", len(decoded.StringMap), len(original.StringMap))
		} else {
			for k, v := range original.StringMap {
				if decoded.StringMap[k] != v {
					t.Errorf("StringMap[%q] mismatch: got %q, want %q", k, decoded.StringMap[k], v)
				}
			}
		}

		if len(decoded.IntMap) != len(original.IntMap) {
			t.Errorf("IntMap length mismatch: got %d, want %d", len(decoded.IntMap), len(original.IntMap))
		} else {
			for k, v := range original.IntMap {
				if decoded.IntMap[k] != v {
					t.Errorf("IntMap[%q] mismatch: got %d, want %d", k, decoded.IntMap[k], v)
				}
			}
		}

		if len(decoded.Float64Map) != len(original.Float64Map) {
			t.Errorf("Float64Map length mismatch: got %d, want %d", len(decoded.Float64Map), len(original.Float64Map))
		} else {
			for k, v := range original.Float64Map {
				if decoded.Float64Map[k] != v {
					t.Errorf("Float64Map[%q] mismatch: got %v, want %v", k, decoded.Float64Map[k], v)
				}
			}
		}

		// Verify pointer fields
		if (original.PtrString == nil) != (decoded.PtrString == nil) {
			t.Errorf("PtrString nil mismatch: original=%v, decoded=%v", original.PtrString == nil, decoded.PtrString == nil)
		} else if original.PtrString != nil && *decoded.PtrString != *original.PtrString {
			t.Errorf("PtrString mismatch: got %q, want %q", *decoded.PtrString, *original.PtrString)
		}

		if (original.PtrInt == nil) != (decoded.PtrInt == nil) {
			t.Errorf("PtrInt nil mismatch: original=%v, decoded=%v", original.PtrInt == nil, decoded.PtrInt == nil)
		} else if original.PtrInt != nil && *decoded.PtrInt != *original.PtrInt {
			t.Errorf("PtrInt mismatch: got %d, want %d", *decoded.PtrInt, *original.PtrInt)
		}

		if (original.PtrInt64 == nil) != (decoded.PtrInt64 == nil) {
			t.Errorf("PtrInt64 nil mismatch: original=%v, decoded=%v", original.PtrInt64 == nil, decoded.PtrInt64 == nil)
		} else if original.PtrInt64 != nil && *decoded.PtrInt64 != *original.PtrInt64 {
			t.Errorf("PtrInt64 mismatch: got %d, want %d", *decoded.PtrInt64, *original.PtrInt64)
		}

		if (original.PtrUint64 == nil) != (decoded.PtrUint64 == nil) {
			t.Errorf("PtrUint64 nil mismatch: original=%v, decoded=%v", original.PtrUint64 == nil, decoded.PtrUint64 == nil)
		} else if original.PtrUint64 != nil && *decoded.PtrUint64 != *original.PtrUint64 {
			t.Errorf("PtrUint64 mismatch: got %d, want %d", *decoded.PtrUint64, *original.PtrUint64)
		}

		if (original.PtrFloat64 == nil) != (decoded.PtrFloat64 == nil) {
			t.Errorf("PtrFloat64 nil mismatch: original=%v, decoded=%v", original.PtrFloat64 == nil, decoded.PtrFloat64 == nil)
		} else if original.PtrFloat64 != nil && *decoded.PtrFloat64 != *original.PtrFloat64 {
			t.Errorf("PtrFloat64 mismatch: got %v, want %v", *decoded.PtrFloat64, *original.PtrFloat64)
		}

		if (original.PtrBool == nil) != (decoded.PtrBool == nil) {
			t.Errorf("PtrBool nil mismatch: original=%v, decoded=%v", original.PtrBool == nil, decoded.PtrBool == nil)
		} else if original.PtrBool != nil && *decoded.PtrBool != *original.PtrBool {
			t.Errorf("PtrBool mismatch: got %v, want %v", *decoded.PtrBool, *original.PtrBool)
		}
	})
}
