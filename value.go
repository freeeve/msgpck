package msgpck

// Type represents the type of a decoded MessagePack value
type Type uint8

const (
	TypeNil Type = iota
	TypeBool
	TypeInt
	TypeUint
	TypeFloat32
	TypeFloat64
	TypeString
	TypeBinary
	TypeArray
	TypeMap
	TypeExt
)

// String returns the string representation of the type
func (t Type) String() string {
	switch t {
	case TypeNil:
		return "nil"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeUint:
		return "uint"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeString:
		return "string"
	case TypeBinary:
		return "binary"
	case TypeArray:
		return "array"
	case TypeMap:
		return "map"
	case TypeExt:
		return "ext"
	default:
		return "unknown"
	}
}

// Value represents a decoded MessagePack value.
// For strings and binary data, Bytes points directly into the source buffer (zero-copy).
// The caller must not modify the source buffer while Value is in use.
type Value struct {
	Type    Type
	Bool    bool
	Int     int64
	Uint    uint64
	Float32 float32
	Float64 float64
	Bytes   []byte  // string/binary - points into source
	Array   []Value // array elements
	Map     []KV    // map key-value pairs
	Ext     Ext     // extension data
}

// KV represents a key-value pair in a map.
// Key is a []byte pointing into the source buffer for zero-copy.
type KV struct {
	Key   []byte // points into source
	Value Value
}

// Ext represents MessagePack extension data
type Ext struct {
	Type int8
	Data []byte // points into source
}

// IsNil returns true if the value is nil
func (v *Value) IsNil() bool {
	return v.Type == TypeNil
}

// AsBool returns the value as bool. Panics if Type != TypeBool.
func (v *Value) AsBool() bool {
	return v.Bool
}

// AsInt returns the value as int64. Works for TypeInt and TypeUint.
func (v *Value) AsInt() int64 {
	if v.Type == TypeUint {
		return int64(v.Uint)
	}
	return v.Int
}

// AsUint returns the value as uint64. Works for TypeInt and TypeUint.
func (v *Value) AsUint() uint64 {
	if v.Type == TypeInt {
		return uint64(v.Int)
	}
	return v.Uint
}

// AsFloat64 returns the value as float64. Works for float32/64 and int/uint.
func (v *Value) AsFloat64() float64 {
	switch v.Type {
	case TypeFloat64:
		return v.Float64
	case TypeFloat32:
		return float64(v.Float32)
	case TypeInt:
		return float64(v.Int)
	case TypeUint:
		return float64(v.Uint)
	default:
		return 0
	}
}

// AsString returns Bytes as a string. This allocates.
func (v *Value) AsString() string {
	return string(v.Bytes)
}

// AsBytes returns Bytes directly (zero-copy reference to source).
func (v *Value) AsBytes() []byte {
	return v.Bytes
}

// Len returns the length of array/map/string/binary.
func (v *Value) Len() int {
	switch v.Type {
	case TypeArray:
		return len(v.Array)
	case TypeMap:
		return len(v.Map)
	case TypeString, TypeBinary:
		return len(v.Bytes)
	default:
		return 0
	}
}

// Index returns the i-th element of an array. Panics if out of bounds or not an array.
func (v *Value) Index(i int) *Value {
	return &v.Array[i]
}

// Get returns the value for a key in a map (linear search).
// Returns nil if not found or not a map.
func (v *Value) Get(key []byte) *Value {
	if v.Type != TypeMap {
		return nil
	}
	for i := range v.Map {
		if bytesEqual(v.Map[i].Key, key) {
			return &v.Map[i].Value
		}
	}
	return nil
}

// GetString is a convenience for Get with a string key.
func (v *Value) GetString(key string) *Value {
	return v.Get([]byte(key))
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
