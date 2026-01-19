package msgpck

import (
	"testing"

	vmihailenco "github.com/vmihailenco/msgpack/v5"
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
	smallStruct = SmallStruct{Name: "Alice", Age: 30}
	mediumStruct = MediumStruct{
		ID: 12345, Name: "Bob Smith", Email: "bob@example.com",
		Age: 42, Active: true, Score: 98.6,
		Tags:     []string{"admin", "user", "premium"},
		Metadata: map[string]string{"role": "manager", "dept": "engineering"},
	}
	smallMap = map[string]any{"name": "Alice", "age": 30}
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
	smallStructDec   = GetStructDecoder[SmallStruct]()
	smallStructDecZC = GetStructDecoderZeroCopy[SmallStruct]()
	smallStructEnc   = GetStructEncoder[SmallStruct]()

	mediumStructDec   = GetStructDecoder[MediumStruct]()
	mediumStructDecZC = GetStructDecoderZeroCopy[MediumStruct]()
	mediumStructEnc   = GetStructEncoder[MediumStruct]()
)

// ============================================================================
// Map Encoding Benchmarks
// ============================================================================

func BenchmarkVmihailenco_SmallMap_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vmihailenco.Marshal(smallMap)
	}
}

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

func BenchmarkVmihailenco_MediumMap_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vmihailenco.Marshal(mediumMap)
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

func BenchmarkVmihailenco_SmallMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		vmihailenco.Unmarshal(data, &m)
	}
}

func BenchmarkMsgpck_SmallMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMap(data)
	}
}

func BenchmarkMsgpck_SmallMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapZeroCopy(data)
	}
}

func BenchmarkVmihailenco_MediumMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]any
		vmihailenco.Unmarshal(data, &m)
	}
}

func BenchmarkMsgpck_MediumMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMap(data)
	}
}

func BenchmarkMsgpck_MediumMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapZeroCopy(data)
	}
}

// ============================================================================
// Typed Map Decoding Benchmarks
// ============================================================================

func BenchmarkVmihailenco_StringMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]string
		vmihailenco.Unmarshal(data, &m)
	}
}

func BenchmarkMsgpck_StringMap_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, false)
	}
}

func BenchmarkMsgpck_StringMap_UnmarshalZeroCopy(b *testing.B) {
	data, _ := vmihailenco.Marshal(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UnmarshalMapStringString(data, true)
	}
}

// ============================================================================
// Struct Encoding Benchmarks
// ============================================================================

func BenchmarkVmihailenco_SmallStruct_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vmihailenco.Marshal(smallStruct)
	}
}

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

func BenchmarkVmihailenco_MediumStruct_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vmihailenco.Marshal(mediumStruct)
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

func BenchmarkVmihailenco_SmallStruct_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		vmihailenco.Unmarshal(data, &s)
	}
}

func BenchmarkMsgpck_SmallStruct_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpck_SmallStruct_UnmarshalPreReg(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpck_SmallStruct_UnmarshalZeroCopy(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s SmallStruct
		smallStructDecZC.Decode(data, &s)
	}
}

func BenchmarkVmihailenco_MediumStruct_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		vmihailenco.Unmarshal(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_Unmarshal(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		UnmarshalStruct(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_UnmarshalPreReg(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s MediumStruct
		mediumStructDec.Decode(data, &s)
	}
}

func BenchmarkMsgpck_MediumStruct_UnmarshalZeroCopy(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumStruct)
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
	data, _ := vmihailenco.Marshal(smallStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *SmallStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpck_MediumStruct_Callback(b *testing.B) {
	data, _ := vmihailenco.Marshal(mediumStruct)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStructFunc(data, func(s *MediumStruct) error {
			_ = s.Name
			return nil
		})
	}
}

func BenchmarkMsgpck_SmallMap_Callback(b *testing.B) {
	data, _ := vmihailenco.Marshal(smallMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeMapFunc(data, func(m map[string]any) error {
			_ = m["name"]
			return nil
		})
	}
}

func BenchmarkMsgpck_StringMap_Callback(b *testing.B) {
	data, _ := vmihailenco.Marshal(stringMap)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeStringMapFunc(data, func(m map[string]string) error {
			_ = m["name"]
			return nil
		})
	}
}
