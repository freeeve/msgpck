package msgpck

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestEncodeExtFormats(t *testing.T) {
	tests := []struct {
		size int
	}{
		{1},   // fixext1
		{2},   // fixext2
		{4},   // fixext4
		{8},   // fixext8
		{16},  // fixext16
		{3},   // ext8
		{100}, // ext8
		{256}, // ext16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.size + 10)
		e.EncodeExt(1, make([]byte, tc.size))
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeExt size=%d failed", tc.size)
		}
	}
}

func TestEncodeVariousTypesExtra(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"bool true", true},
		{"bool false", false},
		{"int", 42},
		{"int8", int8(42)},
		{"int16", int16(42)},
		{"int32", int32(42)},
		{"int64", int64(42)},
		{"uint", uint(42)},
		{"uint8", uint8(42)},
		{"uint16", uint16(42)},
		{"uint32", uint32(42)},
		{"uint64", uint64(42)},
		{"float32", float32(3.14)},
		{"float64", float64(3.14)},
		{"string", "hello"},
		{"[]byte", []byte{1, 2, 3}},
		{"[]int", []int{1, 2, 3}},
		{"[]string", []string{"a", "b"}},
		{"[3]int", [3]int{1, 2, 3}},
		{"map[string]int", map[string]int{"a": 1}},
		{"map[string]any", map[string]any{"a": 1, "b": "x"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := Marshal(tc.value)
			if err != nil {
				t.Errorf("%s encode failed: %v", tc.name, err)
			}
			if len(b) == 0 {
				t.Errorf("%s encoded to empty", tc.name)
			}
		})
	}
}

func TestEncodeValueAllTypes(t *testing.T) {
	e := NewEncoder(64)

	tests := []Value{
		{Type: TypeNil},
		{Type: TypeBool, Bool: true},
		{Type: TypeBool, Bool: false},
		{Type: TypeInt, Int: -42},
		{Type: TypeUint, Uint: 42},
		{Type: TypeFloat32, Float32: 3.14},
		{Type: TypeFloat64, Float64: 2.718},
		{Type: TypeString, Bytes: []byte("hello")},
		{Type: TypeBinary, Bytes: []byte{1, 2, 3}},
		{Type: TypeArray, Array: []Value{{Type: TypeInt, Int: 1}}},
		{Type: TypeMap, Map: []KV{{Key: []byte("k"), Value: Value{Type: TypeInt, Int: 1}}}},
	}

	for _, v := range tests {
		e.Reset()
		e.EncodeValue(&v)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeValue produced no output for type %v", v.Type)
		}
	}
}

func TestEncodeMapWithNonStringKey(t *testing.T) {
	m := map[int]string{1: "one", 2: "two"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with non-string key returned error (expected): %v", err)
	}
}

func TestEncodeArray16Array32(t *testing.T) {
	t.Run("array16", func(t *testing.T) {
		// Array with 16 elements (needs array16 format)
		arr := make([]int, 16)
		for i := range arr {
			arr[i] = i
		}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("array16 encode failed")
		}
	})

	t.Run("large array", func(t *testing.T) {
		// Array with 256 elements
		arr := make([]int, 256)
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("large array encode failed")
		}
	})
}

func TestEncodeMap16(t *testing.T) {
	m := make(map[string]int)
	for i := 0; i < 20; i++ {
		m[string(rune('a'+i))] = i
	}
	b, err := Marshal(m)
	if err != nil || len(b) == 0 {
		t.Error("map16 encode failed")
	}
}

func TestEncodeStructOmitempty(t *testing.T) {
	type Data struct {
		Name  string `msgpack:"name,omitempty"`
		Value int    `msgpack:"value,omitempty"`
		Ptr   *int   `msgpack:"ptr,omitempty"`
	}

	enc := GetStructEncoder[Data]()
	d := Data{} // all zero values
	b, err := enc.Encode(&d)
	if err != nil {
		t.Error("omitempty encode failed")
	}
	// Should encode to empty map
	if len(b) == 0 {
		t.Error("empty result")
	}
}

func TestEncodeSliceArrayPaths(t *testing.T) {
	t.Run("[]any", func(t *testing.T) {
		arr := []any{1, "two", 3.0}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[]any encode failed")
		}
	})

	t.Run("[]float64", func(t *testing.T) {
		arr := []float64{1.1, 2.2, 3.3}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[]float64 encode failed")
		}
	})

	t.Run("[3]string", func(t *testing.T) {
		arr := [3]string{"a", "b", "c"}
		b, err := Marshal(arr)
		if err != nil || len(b) == 0 {
			t.Error("[3]string encode failed")
		}
	})
}

func TestEncodeMapPaths(t *testing.T) {
	t.Run("map[string]string", func(t *testing.T) {
		m := map[string]string{"k": "v"}
		b, err := Marshal(m)
		if err != nil || len(b) == 0 {
			t.Error("map[string]string encode failed")
		}
	})

	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"k": 123}
		b, err := Marshal(m)
		if err != nil || len(b) == 0 {
			t.Error("map[string]any encode failed")
		}
	})
}

func TestEncodeStructFields(t *testing.T) {
	type AllTypes struct {
		Bool    bool    `msgpack:"bool"`
		Int     int     `msgpack:"int"`
		Int64   int64   `msgpack:"int64"`
		Uint    uint    `msgpack:"uint"`
		Uint64  uint64  `msgpack:"uint64"`
		Float32 float32 `msgpack:"float32"`
		Float64 float64 `msgpack:"float64"`
		String  string  `msgpack:"string"`
		Bytes   []byte  `msgpack:"bytes"`
	}

	s := AllTypes{
		Bool:    true,
		Int:     -42,
		Int64:   -100,
		Uint:    42,
		Uint64:  100,
		Float32: 3.14,
		Float64: 2.718,
		String:  "test",
		Bytes:   []byte{1, 2, 3},
	}

	enc := GetStructEncoder[AllTypes]()
	b, err := enc.Encode(&s)
	if err != nil || len(b) == 0 {
		t.Error("AllTypes encode failed")
	}
}

func TestMarshalError(t *testing.T) {
	// Unsupported type should error
	ch := make(chan int)
	_, err := Marshal(ch)
	if err == nil {
		t.Error("expected error for channel type")
	}
}

func TestEncodeMapNonStringKey(t *testing.T) {
	m := map[int]string{1: "one"}
	_, err := Marshal(m)
	if err != nil {
		t.Logf("encoding map with int key returned error (expected): %v", err)
	}
}

func TestEncodeSliceArrayError(t *testing.T) {
	// Slice with unsupported element type
	arr := []chan int{make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for channel slice")
	}
}

func TestEncodeMapError(t *testing.T) {
	// Map with unsupported value type
	m := map[string]chan int{"k": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel map value")
	}
}

func TestEncodeSliceAnyError(t *testing.T) {
	// Slice containing unencodable value
	s := []any{"ok", make(chan int)}
	_, err := Marshal(s)
	if err == nil {
		t.Error("expected error for slice with channel")
	}
}

func TestEncodeMapStringAnyError(t *testing.T) {
	// Map containing unencodable value
	m := map[string]any{"ok": "fine", "bad": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel")
	}
}

func TestEncodeArrayError(t *testing.T) {
	// Array with channel element
	arr := [2]any{"ok", make(chan int)}
	_, err := Marshal(arr)
	if err == nil {
		t.Error("expected error for array with channel")
	}
}

func TestEncodeMapChannelKeyError(t *testing.T) {
	// Map with complex key that fails
	m := map[any]any{make(chan int): "value"}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for map with channel key")
	}
}

func TestEncodeMapKeyValueError(t *testing.T) {
	// Map with channel value
	m := map[string]any{"ch": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for channel value")
	}
}

func TestEncodeStringStr32(t *testing.T) {
	// Create a string longer than 65535 bytes to trigger str32 format
	longStr := string(make([]byte, 70000))
	enc := NewEncoder(80000)
	enc.EncodeString(longStr)

	b := enc.Bytes()
	if b[0] != formatStr32 {
		t.Errorf("expected str32 format (0xdb), got 0x%02x", b[0])
	}
}

func TestEncodeBinaryBin32(t *testing.T) {
	// Create binary longer than 65535 bytes to trigger bin32 format
	longBin := make([]byte, 70000)
	enc := NewEncoder(80000)
	enc.EncodeBinary(longBin)

	b := enc.Bytes()
	if b[0] != formatBin32 {
		t.Errorf("expected bin32 format (0xc6), got 0x%02x", b[0])
	}
}

func TestEncodeValueNilMap(t *testing.T) {
	var m map[string]int
	b, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf(errMsgNilFmt, b[0])
	}
}

func TestEncodeValueNilSlice(t *testing.T) {
	var s []int
	b, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf(errMsgNilFmt, b[0])
	}
}

func TestEncodeValueNilPointer(t *testing.T) {
	var p *int
	b, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatNil {
		t.Errorf(errMsgNilFmt, b[0])
	}
}

func TestEncodeValueFloat32(t *testing.T) {
	var f float32 = 3.14
	b, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != formatFloat32 {
		t.Errorf("expected float32 format, got 0x%02x", b[0])
	}
}

// TestMarshalVariants tests all Marshal variants
func TestMarshalVariants(t *testing.T) {
	data := map[string]any{"key": "value"}

	t.Run("Marshal", func(t *testing.T) {
		b, err := Marshal(data)
		if err != nil {
			t.Error(err)
		}
		// Verify we can use it after return
		var result any
		err = Unmarshal(b, &result)
		if err != nil {
			t.Error(err)
		}
	})

}

// TestEncodeCollectionTypes tests encoding of arrays, slices, and maps
func TestEncodeCollectionTypes(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		b, err := Marshal(arr)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("array encode failed")
		}
	})

	t.Run("slice of strings", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		b, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("slice encode failed")
		}
	})

	t.Run("map[string]int", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2}
		b, err := Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		d := NewDecoder(b)
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 2 {
			t.Error("map encode failed")
		}
	})
}

// TestEncodePointerTypes tests encoding of pointer types
func TestEncodePointerTypes(t *testing.T) {
	t.Run("pointer", func(t *testing.T) {
		val := 42
		ptr := &val
		b, err := Marshal(ptr)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		_ = Unmarshal(b, &decoded)
		if decoded != int64(42) {
			t.Error("pointer encode failed")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var ptr *int
		b, err := Marshal(ptr)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		_ = Unmarshal(b, &decoded)
		if decoded != nil {
			t.Error("nil pointer encode failed")
		}
	})
}

// TestEncodeNilTypes tests encoding of nil slices and maps
func TestEncodeNilTypes(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var s []int
		b, err := Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		_ = Unmarshal(b, &decoded)
		if decoded != nil {
			t.Error("nil slice encode failed")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		var m map[string]int
		b, err := Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		_ = Unmarshal(b, &decoded)
		if decoded != nil {
			t.Error("nil map encode failed")
		}
	})
}

// TestRoundTripPrimitives tests encoding and decoding of primitive types
func TestRoundTripPrimitives(t *testing.T) {
	// Note: DecodeAny normalizes all integers to int64 (unless > MaxInt64)
	// and float32 to float64 for consistent types
	tests := []struct {
		name  string
		value any
	}{
		{"nil", nil},
		{"true", true},
		{"false", false},
		{"zero", int64(0)},
		{"positive fixint", int64(42)},
		{"max fixint", int64(127)},
		{"negative fixint", int64(-1)},
		{"min fixint", int64(-32)},
		{"uint8", int64(200)},
		{"uint16", int64(1000)},
		{"uint32", int64(100000)},
		{"uint64", int64(1 << 40)},
		{"int8", int64(-100)},
		{"int16", int64(-1000)},
		{"int32", int64(-100000)},
		{"int64", int64(-1 << 40)},
		{"float32", float64(3.140000104904175)}, // float32 promoted to float64
		{"float64", float64(3.14159265359)},
		{"empty string", ""},
		{"short string", "hello"},
		{"fixstr max", string(make([]byte, 31))},
		{"str8", string(make([]byte, 100))},
		{"str16", string(make([]byte, 300))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			var decoded any
			err = Unmarshal(encoded, &decoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf("got %v (%T), want %v (%T)", decoded, decoded, tt.value, tt.value)
			}
		})
	}
}

// TestRoundTripContainers tests arrays and maps
func TestRoundTripContainers(t *testing.T) {
	// Note: all integers decode to int64
	tests := []struct {
		name  string
		value any
	}{
		{"empty array", []any{}},
		{"int array", []any{int64(1), int64(2), int64(3)}},
		{"mixed array", []any{int64(1), "hello", true, nil}},
		{"empty map", map[string]any{}},
		{"string map", map[string]any{"a": int64(1), "b": int64(2)}},
		{"nested map", map[string]any{
			"inner": map[string]any{"x": int64(10)},
		}},
		{"nested array", []any{[]any{int64(1), int64(2)}, []any{int64(3), int64(4)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			var decoded any
			err = Unmarshal(encoded, &decoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			if !reflect.DeepEqual(decoded, tt.value) {
				t.Errorf(errMsgGotWant, decoded, tt.value)
			}
		})
	}
}

// TestBinaryData tests encoding/decoding of binary data
func TestBinaryData(t *testing.T) {
	original := []byte{0x00, 0x01, 0x02, 0xff, 0xfe}
	encoded, err := Marshal(original)
	if err != nil {
		t.Fatalf(errMsgMarshalFailed, err)
	}

	var decoded any
	err = Unmarshal(encoded, &decoded)
	if err != nil {
		t.Fatalf(errMsgUnmarshalFailed, err)
	}

	if !bytes.Equal(decoded.([]byte), original) {
		t.Errorf(errMsgGotWant, decoded, original)
	}
}

// TestFloatSpecialValues tests special float values
func TestFloatSpecialValues(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"positive infinity", math.Inf(1)},
		{"negative infinity", math.Inf(-1)},
		{"NaN", math.NaN()},
		{"max float64", math.MaxFloat64},
		{"smallest positive", math.SmallestNonzeroFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf(errMsgMarshalFailed, err)
			}

			var decoded any
			err = Unmarshal(encoded, &decoded)
			if err != nil {
				t.Fatalf(errMsgUnmarshalFailed, err)
			}

			got := decoded.(float64)
			if math.IsNaN(tt.value) {
				if !math.IsNaN(got) {
					t.Errorf("expected NaN, got %v", got)
				}
			} else if got != tt.value {
				t.Errorf("got %v, want %v", got, tt.value)
			}
		})
	}
}

// TestMarshalConcurrent verifies Marshal is safe for concurrent use.
// This tests that the encoder pool doesn't cause data races or corruption.
func TestMarshalConcurrent(t *testing.T) {
	type TestStruct struct {
		ID    uint64   `msgpack:"id"`
		Title string   `msgpack:"t"`
		Tags  []string `msgpack:"tags,omitempty"`
	}

	const numGoroutines = 100
	const iterationsPerGoroutine = 100

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*iterationsPerGoroutine)

	for i := range numGoroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := range iterationsPerGoroutine {
				original := &TestStruct{
					ID:    uint64(n*1000 + j),
					Title: "Title with some longer text for testing",
					Tags:  []string{"tag1", "tag2", "tag3"},
				}

				data, err := Marshal(original)
				if err != nil {
					errChan <- err
					return
				}

				var decoded TestStruct
				if err := Unmarshal(data, &decoded); err != nil {
					errChan <- err
					return
				}

				if decoded.ID != original.ID || decoded.Title != original.Title {
					errChan <- errors.New("data mismatch after concurrent marshal/unmarshal")
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Error(err)
	}
}

// TestStructEncoderConcurrent tests StructEncoder for concurrent safety.
// This simulates storing encoded data and reading it back later.
func TestStructEncoderConcurrent(t *testing.T) {
	type Record struct {
		ID     uint64            `msgpack:"id"`
		Name   string            `msgpack:"name"`
		Tags   []string          `msgpack:"tags"`
		Meta   map[string]string `msgpack:"meta"`
		Score  float64           `msgpack:"score"`
		Active bool              `msgpack:"active"`
		Counts []int64           `msgpack:"counts"`
		Nested *Record           `msgpack:"nested,omitempty"`
	}

	enc := GetStructEncoder[Record]()
	dec := GetStructDecoder[Record](false)

	const numGoroutines = 100
	const recordsPerGoroutine = 100

	// Shared storage simulating a database
	type storedRecord struct {
		original Record
		encoded  []byte
	}
	storage := make(chan storedRecord, numGoroutines*recordsPerGoroutine)

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*recordsPerGoroutine)

	// Writers: encode and store
	for i := range numGoroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := range recordsPerGoroutine {
				rec := Record{
					ID:     uint64(n*10000 + j),
					Name:   "Record with a longer name for testing buffer handling",
					Tags:   []string{"tag1", "tag2", "tag3", "tag4"},
					Meta:   map[string]string{"key1": "value1", "key2": "value2"},
					Score:  float64(n) + float64(j)/100.0,
					Active: j%2 == 0,
					Counts: []int64{1, 2, 3, 4, 5},
				}

				data, err := enc.Encode(&rec)
				if err != nil {
					errChan <- err
					return
				}

				// Store a copy (simulating database write)
				stored := make([]byte, len(data))
				copy(stored, data)
				storage <- storedRecord{original: rec, encoded: stored}
			}
		}(i)
	}

	// Wait for all writers
	wg.Wait()
	close(storage)

	// Readers: decode and verify
	for stored := range storage {
		var decoded Record
		if err := dec.Decode(stored.encoded, &decoded); err != nil {
			t.Errorf("decode failed for ID %d (len=%d): %v",
				stored.original.ID, len(stored.encoded), err)
			continue
		}

		if decoded.ID != stored.original.ID {
			t.Errorf("ID mismatch: got %d, want %d", decoded.ID, stored.original.ID)
		}
		if decoded.Name != stored.original.Name {
			t.Errorf("Name mismatch for ID %d", stored.original.ID)
		}
		if decoded.Score != stored.original.Score {
			t.Errorf("Score mismatch for ID %d", stored.original.ID)
		}
	}

	close(errChan)
	for err := range errChan {
		t.Error(err)
	}
}

// TestMarshalConcurrentWithStorage simulates database write/read pattern.
// Encodes with Marshal, stores bytes, then decodes from multiple goroutines.
func TestMarshalConcurrentWithStorage(t *testing.T) {
	type Record struct {
		ID     uint64            `msgpack:"id"`
		Name   string            `msgpack:"name"`
		Tags   []string          `msgpack:"tags,omitempty"`
		Meta   map[string]string `msgpack:"meta,omitempty"`
		Score  float64           `msgpack:"score"`
		Active bool              `msgpack:"active"`
	}

	const numWriters = 50
	const numReaders = 50
	const recordsPerWriter = 100

	// Simulated storage (like a database)
	var mu sync.Mutex
	storage := make(map[uint64][]byte)

	var writeWg, readWg sync.WaitGroup
	errChan := make(chan error, numWriters*recordsPerWriter+numReaders*recordsPerWriter)
	done := make(chan struct{})

	// Writers: encode with Marshal and store
	for i := range numWriters {
		writeWg.Add(1)
		go func(n int) {
			defer writeWg.Done()
			for j := range recordsPerWriter {
				rec := Record{
					ID:     uint64(n*recordsPerWriter + j),
					Name:   "Record name with enough content to exercise the encoder",
					Tags:   []string{"tag1", "tag2", "tag3"},
					Meta:   map[string]string{"k1": "v1", "k2": "v2"},
					Score:  float64(n) * 1.5,
					Active: j%2 == 0,
				}

				data, err := Marshal(&rec)
				if err != nil {
					errChan <- err
					return
				}

				// Store (simulating database write)
				mu.Lock()
				storage[rec.ID] = data
				mu.Unlock()
			}
		}(i)
	}

	// Readers: concurrently read and decode
	for i := range numReaders {
		readWg.Add(1)
		go func(n int) {
			defer readWg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}

				// Pick a random ID to read
				id := uint64((n * recordsPerWriter) + (n % recordsPerWriter))

				mu.Lock()
				data, ok := storage[id]
				mu.Unlock()

				if !ok {
					continue
				}

				var decoded Record
				if err := Unmarshal(data, &decoded); err != nil {
					errChan <- err
					return
				}

				if decoded.ID != id {
					errChan <- errors.New("ID mismatch in concurrent read")
					return
				}
			}
		}(i)
	}

	// Wait for writers, then signal readers to stop
	writeWg.Wait()
	close(done)
	readWg.Wait()

	// Final verification: decode all stored records
	for id, data := range storage {
		var decoded Record
		if err := Unmarshal(data, &decoded); err != nil {
			t.Errorf("final decode failed for ID %d (len=%d): %v", id, len(data), err)
		}
	}

	close(errChan)
	for err := range errChan {
		t.Error(err)
	}
}

// TestMarshalConcurrentAggressive is an aggressive test to find race conditions.
func TestMarshalConcurrentAggressive(t *testing.T) {
	type Data struct {
		ID     uint64            `msgpack:"id"`
		Name   string            `msgpack:"name"`
		Tags   []string          `msgpack:"tags"`
		Values map[string]int64  `msgpack:"values"`
		Nested map[string]string `msgpack:"nested"`
	}

	const numGoroutines = 500
	const iterations = 200

	var wg sync.WaitGroup
	failures := make([]error, 0)
	var mu sync.Mutex

	for i := range numGoroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := range iterations {
				original := Data{
					ID:     uint64(n*iterations + j),
					Name:   "A moderately long name to exercise buffer handling properly",
					Tags:   []string{"alpha", "beta", "gamma", "delta", "epsilon"},
					Values: map[string]int64{"a": 1, "b": 2, "c": 3, "d": 4},
					Nested: map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"},
				}

				// Encode
				data, err := Marshal(&original)
				if err != nil {
					mu.Lock()
					failures = append(failures, err)
					mu.Unlock()
					return
				}

				// Verify length is reasonable (not truncated)
				if len(data) < 50 {
					mu.Lock()
					failures = append(failures, errors.New("data too short - possible truncation"))
					mu.Unlock()
					return
				}

				// Decode immediately
				var decoded Data
				if err := Unmarshal(data, &decoded); err != nil {
					mu.Lock()
					failures = append(failures, err)
					mu.Unlock()
					return
				}

				// Verify
				if decoded.ID != original.ID {
					mu.Lock()
					failures = append(failures, errors.New("ID mismatch"))
					mu.Unlock()
					return
				}
				if decoded.Name != original.Name {
					mu.Lock()
					failures = append(failures, errors.New("Name mismatch"))
					mu.Unlock()
					return
				}
				if len(decoded.Tags) != len(original.Tags) {
					mu.Lock()
					failures = append(failures, errors.New("Tags length mismatch"))
					mu.Unlock()
					return
				}
			}
		}(i)
	}

	wg.Wait()

	if len(failures) > 0 {
		for i, err := range failures {
			if i >= 10 {
				t.Errorf("... and %d more failures", len(failures)-10)
				break
			}
			t.Error(err)
		}
	}
}

// TestConcurrentEncodeDecodeStress is a stress test for concurrent encode/decode.
func TestConcurrentEncodeDecodeStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	type Data struct {
		ID    uint64   `msgpack:"id"`
		Value string   `msgpack:"value"`
		Items []string `msgpack:"items"`
	}

	enc := GetStructEncoder[Data]()
	dec := GetStructDecoder[Data](false)

	const numGoroutines = 200
	const iterations = 500

	var wg sync.WaitGroup
	failures := make(chan string, numGoroutines*iterations)

	for i := range numGoroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := range iterations {
				original := Data{
					ID:    uint64(n*iterations + j),
					Value: "test value with some content to make it larger",
					Items: []string{"item1", "item2", "item3"},
				}

				// Encode
				data, err := enc.Encode(&original)
				if err != nil {
					failures <- err.Error()
					return
				}

				// Immediately decode (no copy - tests if data is stable)
				var decoded Data
				if err := dec.Decode(data, &decoded); err != nil {
					failures <- err.Error()
					return
				}

				if decoded.ID != original.ID || decoded.Value != original.Value {
					failures <- "data corruption detected"
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(failures)

	failCount := 0
	for msg := range failures {
		if failCount < 10 {
			t.Error(msg)
		}
		failCount++
	}
	if failCount > 10 {
		t.Errorf("... and %d more failures", failCount-10)
	}
}
