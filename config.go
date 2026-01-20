package msgpck

import "math"

// Default limits - aligned with msgpack spec (max 2^32-1)
// We use MaxInt32 for cross-platform safety (32-bit systems)
const (
	DefaultMaxStringLen = math.MaxInt32 // ~2GB, spec allows 2^32-1
	DefaultMaxBinaryLen = math.MaxInt32 // ~2GB, spec allows 2^32-1
	DefaultMaxArrayLen  = math.MaxInt32 // ~2B elements, spec allows 2^32-1
	DefaultMaxMapLen    = math.MaxInt32 // ~2B pairs, spec allows 2^32-1
	DefaultMaxExtLen    = math.MaxInt32 // ~2GB, spec allows 2^32-1
	DefaultMaxDepth     = 10000         // practical limit for nested structures
)

// Config controls decoder/encoder behavior and security limits
type Config struct {
	// MaxStringLen is the maximum allowed string length in bytes
	MaxStringLen int

	// MaxBinaryLen is the maximum allowed binary data length in bytes
	MaxBinaryLen int

	// MaxArrayLen is the maximum allowed array length (number of elements)
	MaxArrayLen int

	// MaxMapLen is the maximum allowed map length (number of key-value pairs)
	MaxMapLen int

	// MaxExtLen is the maximum allowed ext data length in bytes
	MaxExtLen int

	// MaxDepth is the maximum allowed nesting depth for arrays and maps
	MaxDepth int
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxStringLen: DefaultMaxStringLen,
		MaxBinaryLen: DefaultMaxBinaryLen,
		MaxArrayLen:  DefaultMaxArrayLen,
		MaxMapLen:    DefaultMaxMapLen,
		MaxExtLen:    DefaultMaxExtLen,
		MaxDepth:     DefaultMaxDepth,
	}
}

// WithMaxStringLen returns a new Config with the specified MaxStringLen
func (c Config) WithMaxStringLen(n int) Config {
	c.MaxStringLen = n
	return c
}

// WithMaxBinaryLen returns a new Config with the specified MaxBinaryLen
func (c Config) WithMaxBinaryLen(n int) Config {
	c.MaxBinaryLen = n
	return c
}

// WithMaxArrayLen returns a new Config with the specified MaxArrayLen
func (c Config) WithMaxArrayLen(n int) Config {
	c.MaxArrayLen = n
	return c
}

// WithMaxMapLen returns a new Config with the specified MaxMapLen
func (c Config) WithMaxMapLen(n int) Config {
	c.MaxMapLen = n
	return c
}

// WithMaxExtLen returns a new Config with the specified MaxExtLen
func (c Config) WithMaxExtLen(n int) Config {
	c.MaxExtLen = n
	return c
}

// WithMaxDepth returns a new Config with the specified MaxDepth
func (c Config) WithMaxDepth(n int) Config {
	c.MaxDepth = n
	return c
}
