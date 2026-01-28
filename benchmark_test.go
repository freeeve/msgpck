package msgpck

import (
	"testing"
)

// Test structs
type SmallStruct struct {
	Name string
	Age  int
}

type MediumStruct struct {
	ID       int64
	Name     string
	Email    string
	Age      int
	Active   bool
	Score    float64
	Tags     []string
	Metadata map[string]string
}

var (
	smallStruct  = SmallStruct{Name: "Alice", Age: 30}
	mediumStruct = MediumStruct{
		ID: 12345, Name: "Bob Smith", Email: "bob@example.com",
		Age: 42, Active: true, Score: 98.6,
		Tags:     []string{"admin", "user", "premium"},
		Metadata: map[string]string{"role": "manager", "dept": "engineering"},
	}
	smallMap  = map[string]any{"name": "Alice", "age": 30}
	mediumMap = map[string]any{
		"id": int64(12345), "name": "Bob Smith", "email": "bob@example.com",
		"age": 42, "active": true, "score": 98.6,
		"tags":     []any{"admin", "user", "premium"},
		"metadata": map[string]any{"role": "manager", "dept": "engineering"},
	}
	stringMap = map[string]string{
		"name": "Alice", "email": testEmail,
		"role": "admin", "dept": "engineering",
	}
)

// Pre-registered codecs (using cached versions)
var (
	smallStructDec   = GetStructDecoder[SmallStruct](false)
	smallStructDecZC = GetStructDecoder[SmallStruct](true)
	smallStructEnc   = GetStructEncoder[SmallStruct]()

	mediumStructDec   = GetStructDecoder[MediumStruct](false)
	mediumStructDecZC = GetStructDecoder[MediumStruct](true)
	mediumStructEnc   = GetStructEncoder[MediumStruct]()
)

// ============================================================================
// Map Encoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallMapMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(smallMap)
	}
}

func BenchmarkMsgpckMediumMapMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(mediumMap)
	}
}

// ============================================================================
// Map Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallMapUnmarshal(b *testing.B) {
	data, _ := Marshal(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		Unmarshal(data, &m)
	}
}

func BenchmarkMsgpckMediumMapUnmarshal(b *testing.B) {
	data, _ := Marshal(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		Unmarshal(data, &m)
	}
}

func BenchmarkMsgpckStringMapUnmarshal(b *testing.B) {
	data, _ := Marshal(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]string
		Unmarshal(data, &m)
	}
}

// ============================================================================
// Struct Encoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallStructMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(smallStruct)
	}
}

func BenchmarkMsgpckSmallStructMarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallStructEnc.Encode(&smallStruct)
	}
}

func BenchmarkMsgpckMediumStructMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(mediumStruct)
	}
}

func BenchmarkMsgpckMediumStructMarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumStructEnc.Encode(&mediumStruct)
	}
}

// ============================================================================
// Struct Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpckSmallStructUnmarshal(b *testing.B) {
	data, _ := Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		Unmarshal(data, &s)
	}
}

func BenchmarkMsgpckSmallStructUnmarshalPreReg(b *testing.B) {
	data, _ := Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpckSmallStructUnmarshalZeroCopy(b *testing.B) {
	data, _ := Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDecZC.Decode(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshal(b *testing.B) {
	data, _ := Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		Unmarshal(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshalPreReg(b *testing.B) {
	data, _ := Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpckMediumStructUnmarshalZeroCopy(b *testing.B) {
	data, _ := Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDecZC.Decode(data, &s)
	}
}

// ============================================================================
// Simple Benchmarks (from msgpck_test.go)
// ============================================================================

// BenchmarkDecodeMap benchmarks map decoding (the hot path)
func BenchmarkDecodeMap(b *testing.B) {
	data := map[string]any{
		"name":   "Alice",
		"age":    30,
		"email":  testEmail,
		"active": true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		d := NewDecoder(encoded)
		_, _ = d.Decode()
	}
}

// BenchmarkDecodeMapAny benchmarks decoding to map[string]any
func BenchmarkDecodeMapAny(b *testing.B) {
	data := map[string]any{
		"name":   "Alice",
		"age":    30,
		"email":  testEmail,
		"active": true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		d := NewDecoder(encoded)
		_, _ = d.DecodeAny()
	}
}

// BenchmarkEncodeMap benchmarks map encoding
func BenchmarkEncodeMap(b *testing.B) {
	data := map[string]any{
		"name":   "Alice",
		"age":    30,
		"email":  testEmail,
		"active": true,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = Marshal(data)
	}
}

// BenchmarkDecodeStruct benchmarks struct decoding
func BenchmarkDecodeStruct(b *testing.B) {
	type Person struct {
		Name   string `msgpack:"name"`
		Age    int    `msgpack:"age"`
		Email  string `msgpack:"email"`
		Active bool   `msgpack:"active"`
	}

	data := Person{
		Name:   "Alice",
		Age:    30,
		Email:  testEmail,
		Active: true,
	}
	encoded, _ := Marshal(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var p Person
		d := NewDecoder(encoded)
		_ = d.DecodeStruct(&p)
	}
}

// Benchmark batch array encoding vs individual encoding
func BenchmarkEncodeInt64ArrayBatch(b *testing.B) {
	arr := make([]int64, 100)
	for i := range arr {
		arr[i] = int64(i * 7)
	}
	e := NewEncoder(1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Reset()
		e.EncodeInt64Array(arr)
	}
}

func BenchmarkEncodeInt64ArrayManual(b *testing.B) {
	arr := make([]int64, 100)
	for i := range arr {
		arr[i] = int64(i * 7)
	}
	e := NewEncoder(1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Reset()
		e.EncodeArrayHeader(len(arr))
		for _, v := range arr {
			e.EncodeInt(v)
		}
	}
}

func BenchmarkDecodeInt64ArrayBatch(b *testing.B) {
	arr := make([]int64, 100)
	for i := range arr {
		arr[i] = int64(i * 7)
	}
	e := NewEncoder(1024)
	e.EncodeInt64Array(arr)
	data := make([]byte, len(e.Bytes()))
	copy(data, e.Bytes())
	d := NewDecoder(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Reset(data)
		d.DecodeInt64Array()
	}
}

func BenchmarkDecodeInt64ArrayManual(b *testing.B) {
	arr := make([]int64, 100)
	for i := range arr {
		arr[i] = int64(i * 7)
	}
	e := NewEncoder(1024)
	e.EncodeInt64Array(arr)
	data := make([]byte, len(e.Bytes()))
	copy(data, e.Bytes())
	d := NewDecoder(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Reset(data)
		v, _ := d.Decode()
		_ = v.Array
	}
}
