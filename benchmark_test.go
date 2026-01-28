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

// LargeStruct has many fields to benchmark worst-case struct handling
type LargeStruct struct {
	ID           int64             `msgpack:"id"`
	UUID         string            `msgpack:"uuid"`
	Name         string            `msgpack:"name"`
	Email        string            `msgpack:"email"`
	Phone        string            `msgpack:"phone"`
	Address      string            `msgpack:"address"`
	City         string            `msgpack:"city"`
	Country      string            `msgpack:"country"`
	PostalCode   string            `msgpack:"postal_code"`
	Age          int               `msgpack:"age"`
	Score        float64           `msgpack:"score"`
	Rating       float32           `msgpack:"rating"`
	Active       bool              `msgpack:"active"`
	Verified     bool              `msgpack:"verified"`
	Premium      bool              `msgpack:"premium"`
	CreatedAt    int64             `msgpack:"created_at"`
	UpdatedAt    int64             `msgpack:"updated_at"`
	LoginCount   uint32            `msgpack:"login_count"`
	FailedLogins uint16            `msgpack:"failed_logins"`
	Tags         []string          `msgpack:"tags"`
	Roles        []string          `msgpack:"roles"`
	Scores       []int64           `msgpack:"scores"`
	Metadata     map[string]string `msgpack:"metadata"`
}

// GenericSliceStruct tests generic type parameter handling (like roaringsearch sortColumnData)
type GenericSliceStruct[T any] struct {
	Values   []T    `msgpack:"values"`
	MaxDocID uint32 `msgpack:"max_doc_id"`
}

var (
	smallStruct  = SmallStruct{Name: "Alice", Age: 30}
	mediumStruct = MediumStruct{
		ID: 12345, Name: testBobName, Email: testBobEmail,
		Age: 42, Active: true, Score: 98.6,
		Tags:     []string{"admin", "user", "premium"},
		Metadata: map[string]string{"role": "manager", "dept": "engineering"},
	}
	largeStruct = LargeStruct{
		ID: 12345, UUID: "550e8400-e29b-41d4-a716-446655440000",
		Name: testBobName, Email: testBobEmail, Phone: "+1-555-123-4567",
		Address: "123 Main St", City: "San Francisco", Country: "USA", PostalCode: "94102",
		Age: 42, Score: 98.6, Rating: 4.5, Active: true, Verified: true, Premium: true,
		CreatedAt: 1609459200, UpdatedAt: 1704067200, LoginCount: 1500, FailedLogins: 3,
		Tags:     []string{"admin", "user", "premium", "beta", "early-adopter"},
		Roles:    []string{"admin", "moderator", "contributor"},
		Scores:   []int64{100, 95, 88, 92, 97, 100, 85, 90},
		Metadata: map[string]string{"role": "manager", "dept": "engineering", "team": "platform", "level": "senior"},
	}
	smallMap  = map[string]any{"name": "Alice", "age": 30}
	mediumMap = map[string]any{
		"id": int64(12345), "name": testBobName, "email": testBobEmail,
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

	largeStructDec   = GetStructDecoder[LargeStruct](false)
	largeStructDecZC = GetStructDecoder[LargeStruct](true)
	largeStructEnc   = GetStructEncoder[LargeStruct]()

	// Generic slice struct codecs (1M items)
	genericSliceInt64Enc  = GetStructEncoder[GenericSliceStruct[int64]]()
	genericSliceInt64Dec  = GetStructDecoder[GenericSliceStruct[int64]](false)
	genericSliceUint16Enc = GetStructEncoder[GenericSliceStruct[uint16]]()
	genericSliceUint16Dec = GetStructDecoder[GenericSliceStruct[uint16]](false)
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

// ============================================================================
// Large Struct Benchmarks
// ============================================================================

func BenchmarkMsgpckLargeStructMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(largeStruct)
	}
}

func BenchmarkMsgpckLargeStructMarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		largeStructEnc.Encode(&largeStruct)
	}
}

func BenchmarkMsgpckLargeStructUnmarshal(b *testing.B) {
	data, _ := Marshal(largeStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		Unmarshal(data, &s)
	}
}

func BenchmarkMsgpckLargeStructUnmarshalPreReg(b *testing.B) {
	data, _ := Marshal(largeStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		largeStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpckLargeStructUnmarshalZeroCopy(b *testing.B) {
	data, _ := Marshal(largeStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s LargeStruct
		largeStructDecZC.Decode(data, &s)
	}
}

// ============================================================================
// Generic Slice Struct Benchmarks (1M items - simulates roaringsearch sortColumnData)
// ============================================================================

func BenchmarkGenericSlice1MInt64Encode(b *testing.B) {
	values := make([]int64, 1_000_000)
	for i := range values {
		values[i] = int64(i * 7)
	}
	data := GenericSliceStruct[int64]{Values: values, MaxDocID: 999999}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genericSliceInt64Enc.Encode(&data)
	}
}

func BenchmarkGenericSlice1MInt64Decode(b *testing.B) {
	values := make([]int64, 1_000_000)
	for i := range values {
		values[i] = int64(i * 7)
	}
	data := GenericSliceStruct[int64]{Values: values, MaxDocID: 999999}
	encoded, _ := genericSliceInt64Enc.Encode(&data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result GenericSliceStruct[int64]
		genericSliceInt64Dec.Decode(encoded, &result)
	}
}

func BenchmarkGenericSlice1MUint16Encode(b *testing.B) {
	values := make([]uint16, 1_000_000)
	for i := range values {
		values[i] = uint16(i % 65536)
	}
	data := GenericSliceStruct[uint16]{Values: values, MaxDocID: 999999}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genericSliceUint16Enc.Encode(&data)
	}
}

func BenchmarkGenericSlice1MUint16Decode(b *testing.B) {
	values := make([]uint16, 1_000_000)
	for i := range values {
		values[i] = uint16(i % 65536)
	}
	data := GenericSliceStruct[uint16]{Values: values, MaxDocID: 999999}
	encoded, _ := genericSliceUint16Enc.Encode(&data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result GenericSliceStruct[uint16]
		genericSliceUint16Dec.Decode(encoded, &result)
	}
}
