package msgpck

import (
	"bytes"
	"math"
	"testing"
)

func TestStringFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{31},   // max fixstr
		{32},   // str8
		{255},  // max str8
		{256},  // str16
		{1000}, // str16
	}

	for _, tc := range tests {
		s := string(make([]byte, tc.len))
		e := NewEncoder(tc.len + 10)
		e.EncodeStringBytes([]byte(s))
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeStringBytes len=%d failed", tc.len)
		}
	}
}

func TestBinaryFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{255},   // max bin8
		{256},   // bin16
		{65535}, // max bin16
	}

	for _, tc := range tests {
		b := make([]byte, tc.len)
		e := NewEncoder(tc.len + 10)
		e.EncodeBinary(b)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeBinary len=%d failed", tc.len)
		}
	}
}

func TestMapFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixmap
		{16},    // map16
		{65535}, // max map16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len*4 + 10)
		e.EncodeMapHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeMapHeader len=%d failed", tc.len)
		}
	}
}

func TestArrayFormats(t *testing.T) {
	tests := []struct {
		len int
	}{
		{0},
		{15},    // max fixarray
		{16},    // array16
		{65535}, // max array16
	}

	for _, tc := range tests {
		e := NewEncoder(tc.len + 10)
		e.EncodeArrayHeader(tc.len)
		if len(e.Bytes()) == 0 {
			t.Errorf("EncodeArrayHeader len=%d failed", tc.len)
		}
	}
}

func TestExtFormats(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"fixext1", []byte{formatFixExt1, 1, 0xff}},
		{"fixext2", []byte{formatFixExt2, 1, 0xff, 0xff}},
		{"fixext4", []byte{formatFixExt4, 1, 0xff, 0xff, 0xff, 0xff}},
		{"fixext8", []byte{formatFixExt8, 1, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{"fixext16", append([]byte{formatFixExt16, 1}, make([]byte, 16)...)},
		{"ext8", []byte{formatExt8, 3, 1, 0xff, 0xff, 0xff}},
		{"ext16", append([]byte{formatExt16, 0, 3, 1}, make([]byte, 3)...)},
		{"ext32", append([]byte{formatExt32, 0, 0, 0, 3, 1}, make([]byte, 3)...)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := NewDecoder(tc.data)
			v, err := d.Decode()
			if err != nil {
				t.Errorf("%s decode failed: %v", tc.name, err)
			}
			if v.Type != TypeExt {
				t.Errorf("%s: expected ext type, got %v", tc.name, v.Type)
			}
		})
	}
}

func TestTypeString(t *testing.T) {
	types := []Type{
		TypeNil,
		TypeBool,
		TypeInt,
		TypeUint,
		TypeFloat32,
		TypeFloat64,
		TypeString,
		TypeBinary,
		TypeArray,
		TypeMap,
		TypeExt,
	}

	for _, typ := range types {
		s := typ.String()
		if s == "" {
			t.Errorf("String() returned empty for type %v", typ)
		}
	}
}

// TestFormatNilAndBool tests nil and boolean formats
func TestFormatNilAndBool(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		data := []byte{formatNil}
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeNil {
			t.Errorf("nil decode failed: %v, %v", v, err)
		}
	})

	t.Run("false", func(t *testing.T) {
		d := NewDecoder([]byte{formatFalse})
		v, _ := d.Decode()
		if v.Type != TypeBool || v.Bool != false {
			t.Error("false decode failed")
		}
	})

	t.Run("true", func(t *testing.T) {
		d := NewDecoder([]byte{formatTrue})
		v, _ := d.Decode()
		if v.Type != TypeBool || v.Bool != true {
			t.Error("true decode failed")
		}
	})
}

// TestFormatFixints tests positive and negative fixint formats
func TestFormatFixints(t *testing.T) {
	t.Run("positive fixint", func(t *testing.T) {
		for i := 0; i <= 127; i++ {
			d := NewDecoder([]byte{byte(i)})
			v, _ := d.Decode()
			if v.Type != TypeUint || v.Uint != uint64(i) {
				t.Errorf("positive fixint %d failed", i)
			}
		}
	})

	t.Run("negative fixint", func(t *testing.T) {
		for i := -32; i <= -1; i++ {
			d := NewDecoder([]byte{byte(i)})
			v, _ := d.Decode()
			if v.Type != TypeInt || v.Int != int64(i) {
				t.Errorf("negative fixint %d failed: got %v", i, v.Int)
			}
		}
	})
}

// TestFormatUnsignedInts tests uint8-64 formats
func TestFormatUnsignedInts(t *testing.T) {
	t.Run("uint8", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint8, 200})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 200 {
			t.Error("uint8 failed")
		}
	})

	t.Run("uint16", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint16, 0x12, 0x34})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x1234 {
			t.Error("uint16 failed")
		}
	})

	t.Run("uint32", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint32, 0x12, 0x34, 0x56, 0x78})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x12345678 {
			t.Error("uint32 failed")
		}
	})

	t.Run("uint64", func(t *testing.T) {
		d := NewDecoder([]byte{formatUint64, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0})
		v, _ := d.Decode()
		if v.Type != TypeUint || v.Uint != 0x123456789abcdef0 {
			t.Error("uint64 failed")
		}
	})
}

// TestFormatSignedInts tests int8-64 formats
func TestFormatSignedInts(t *testing.T) {
	t.Run("int8", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt8, 0x80}) // -128
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -128 {
			t.Error("int8 failed")
		}
	})

	t.Run("int16", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt16, 0x80, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -32768 {
			t.Error("int16 failed")
		}
	})

	t.Run("int32", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt32, 0x80, 0x00, 0x00, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != -2147483648 {
			t.Error("int32 failed")
		}
	})

	t.Run("int64", func(t *testing.T) {
		d := NewDecoder([]byte{formatInt64, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		v, _ := d.Decode()
		if v.Type != TypeInt || v.Int != math.MinInt64 {
			t.Error("int64 failed")
		}
	})
}

// TestFormatFloats tests float32 and float64 formats
func TestFormatFloats(t *testing.T) {
	t.Run("float32", func(t *testing.T) {
		d := NewDecoder([]byte{formatFloat32, 0x40, 0x48, 0xf5, 0xc3})
		v, _ := d.Decode()
		if v.Type != TypeFloat32 || math.Abs(float64(v.Float32)-3.14) > 0.001 {
			t.Errorf("float32 failed: %v", v.Float32)
		}
	})

	t.Run("float64", func(t *testing.T) {
		d := NewDecoder([]byte{formatFloat64, 0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2d, 0x18})
		v, _ := d.Decode()
		if v.Type != TypeFloat64 || math.Abs(v.Float64-3.14159265359) > 0.0000001 {
			t.Errorf("float64 failed: %v", v.Float64)
		}
	})
}

// TestFormatStrings tests string formats
func TestFormatStrings(t *testing.T) {
	t.Run("fixstr", func(t *testing.T) {
		data := append([]byte{0xa5}, []byte("hello")...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || string(v.Bytes) != "hello" {
			t.Error("fixstr failed")
		}
	})

	t.Run("str8", func(t *testing.T) {
		str := "this is a longer string for str8 format testing"
		data := append([]byte{formatStr8, byte(len(str))}, []byte(str)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || string(v.Bytes) != str {
			t.Error("str8 failed")
		}
	})

	t.Run("str16", func(t *testing.T) {
		str := string(make([]byte, 300))
		data := append([]byte{formatStr16, 0x01, 0x2c}, []byte(str)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeString || len(v.Bytes) != 300 {
			t.Error("str16 failed")
		}
	})
}

// TestFormatBinary tests binary formats
func TestFormatBinary(t *testing.T) {
	t.Run("bin8", func(t *testing.T) {
		bin := []byte{1, 2, 3, 4, 5}
		data := append([]byte{formatBin8, 5}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || !bytes.Equal(v.Bytes, bin) {
			t.Error("bin8 failed")
		}
	})

	t.Run("bin16", func(t *testing.T) {
		bin := make([]byte, 300)
		data := append([]byte{formatBin16, 0x01, 0x2c}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || len(v.Bytes) != 300 {
			t.Error("bin16 failed")
		}
	})

	t.Run("bin32", func(t *testing.T) {
		bin := make([]byte, 70000)
		data := append([]byte{formatBin32, 0x00, 0x01, 0x11, 0x70}, bin...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeBinary || len(v.Bytes) != 70000 {
			t.Error("bin32 failed")
		}
	})
}

// TestFormatArrays tests array formats
func TestFormatArrays(t *testing.T) {
	t.Run("fixarray", func(t *testing.T) {
		data := []byte{0x93, 0x01, 0x02, 0x03}
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 3 {
			t.Error("fixarray failed")
		}
	})

	t.Run("array16", func(t *testing.T) {
		e := NewEncoder(64)
		e.EncodeArrayHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeInt(int64(i))
		}
		d := NewDecoder(e.Bytes())
		v, _ := d.Decode()
		if v.Type != TypeArray || len(v.Array) != 20 {
			t.Error("array16 failed")
		}
	})
}

// TestFormatMaps tests map formats
func TestFormatMaps(t *testing.T) {
	t.Run("fixmap", func(t *testing.T) {
		data := []byte{0x82, 0xa1, 'a', 0x01, 0xa1, 'b', 0x02}
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 2 {
			t.Error("fixmap failed")
		}
	})

	t.Run("map16", func(t *testing.T) {
		e := NewEncoder(256)
		e.EncodeMapHeader(20)
		for i := 0; i < 20; i++ {
			e.EncodeString("key")
			e.EncodeInt(int64(i))
		}
		d := NewDecoder(e.Bytes())
		v, _ := d.Decode()
		if v.Type != TypeMap || len(v.Map) != 20 {
			t.Error("map16 failed")
		}
	})
}

// TestFormatExtensions tests extension formats
func TestFormatExtensions(t *testing.T) {
	t.Run("fixext1", func(t *testing.T) {
		data := append([]byte{formatFixExt1, 0x42}, make([]byte, 1)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 1 {
			t.Error("fixext1 failed")
		}
	})

	t.Run("fixext2", func(t *testing.T) {
		data := append([]byte{formatFixExt2, 0x42}, make([]byte, 2)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 2 {
			t.Error("fixext2 failed")
		}
	})

	t.Run("fixext4", func(t *testing.T) {
		data := append([]byte{formatFixExt4, 0x42}, make([]byte, 4)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 4 {
			t.Error("fixext4 failed")
		}
	})

	t.Run("fixext8", func(t *testing.T) {
		data := append([]byte{formatFixExt8, 0x42}, make([]byte, 8)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 8 {
			t.Error("fixext8 failed")
		}
	})

	t.Run("fixext16", func(t *testing.T) {
		data := append([]byte{formatFixExt16, 0x42}, make([]byte, 16)...)
		d := NewDecoder(data)
		v, err := d.Decode()
		if err != nil || v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 16 {
			t.Error("fixext16 failed")
		}
	})

	t.Run("ext8", func(t *testing.T) {
		data := append([]byte{formatExt8, 5, 0x42}, make([]byte, 5)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 5 {
			t.Error("ext8 failed")
		}
	})

	t.Run("ext16", func(t *testing.T) {
		data := append([]byte{formatExt16, 0x01, 0x00, 0x42}, make([]byte, 256)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 256 {
			t.Error("ext16 failed")
		}
	})

	t.Run("ext32", func(t *testing.T) {
		data := append([]byte{formatExt32, 0x00, 0x01, 0x00, 0x00, 0x42}, make([]byte, 65536)...)
		d := NewDecoder(data)
		v, _ := d.Decode()
		if v.Type != TypeExt || v.Ext.Type != 0x42 || len(v.Ext.Data) != 65536 {
			t.Error("ext32 failed")
		}
	})
}
