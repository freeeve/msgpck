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
		"name": "Alice", "email": "alice@example.com",
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

func BenchmarkMsgpck_SmallMap_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(smallMap)
	}
}

func BenchmarkMsgpck_SmallMap_MarshalCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(smallMap)
	}
}

func BenchmarkMsgpck_MediumMap_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal(mediumMap)
	}
}

func BenchmarkMsgpck_MediumMap_MarshalCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(mediumMap)
	}
}

// ============================================================================
// Map Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpck_SmallMap_Unmarshal(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, false)
	}
}

func BenchmarkMsgpck_SmallMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, true)
	}
}

func BenchmarkMsgpck_MediumMap_Unmarshal(b *testing.B) {
	data, _ := MarshalCopy(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, false)
	}
}

func BenchmarkMsgpck_MediumMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringAny(data, true)
	}
}

// ============================================================================
// Typed Map Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpck_StringMap_Unmarshal(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, false)
	}
}

func BenchmarkMsgpck_StringMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, true)
	}
}

// ============================================================================
// Struct Encoding Benchmarks
// ============================================================================

func BenchmarkMsgpck_SmallStruct_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(smallStruct)
	}
}

func BenchmarkMsgpck_SmallStruct_MarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallStructEnc.Encode(&smallStruct)
	}
}

func BenchmarkMsgpck_SmallStruct_MarshalPreRegCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallStructEnc.EncodeCopy(&smallStruct)
	}
}

func BenchmarkMsgpck_MediumStruct_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalCopy(mediumStruct)
	}
}

func BenchmarkMsgpck_MediumStruct_MarshalPreReg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumStructEnc.Encode(&mediumStruct)
	}
}

func BenchmarkMsgpck_MediumStruct_MarshalPreRegCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumStructEnc.EncodeCopy(&mediumStruct)
	}
}

// ============================================================================
// Struct Decoding Benchmarks
// ============================================================================

func BenchmarkMsgpck_SmallStruct_Unmarshal(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpck_SmallStruct_UnmarshalPreReg(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpck_SmallStruct_UnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDecZC.Decode(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_Unmarshal(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_UnmarshalPreReg(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_UnmarshalZeroCopy(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDecZC.Decode(data, &s)
	}
}

// ============================================================================
// Callback API Benchmarks (Safe Zero-Copy)
// ============================================================================

func BenchmarkMsgpck_SmallStruct_Callback(b *testing.B) {
	data, _ := MarshalCopy(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *SmallStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpck_MediumStruct_Callback(b *testing.B) {
	data, _ := MarshalCopy(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *MediumStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpck_SmallMap_Callback(b *testing.B) {
	data, _ := MarshalCopy(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeMapFunc(data, func(m map[string]any) error {
			_ = m["name"]
			return nil
		})
	}
}

func BenchmarkMsgpck_StringMap_Callback(b *testing.B) {
	data, _ := MarshalCopy(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStringMapFunc(data, func(m map[string]string) error {
			_ = m["name"]
			return nil
		})
	}
}
