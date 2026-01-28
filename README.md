# msgpck

[![CI](https://github.com/freeeve/msgpck/actions/workflows/ci.yml/badge.svg)](https://github.com/freeeve/msgpck/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/freeeve/msgpck/badge.svg?branch=main)](https://coveralls.io/github/freeeve/msgpck?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/freeeve/msgpck)](https://goreportcard.com/report/github.com/freeeve/msgpck)
[![Go Reference](https://pkg.go.dev/badge/github.com/freeeve/msgpck.svg)](https://pkg.go.dev/github.com/freeeve/msgpck)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=freeeve_msgpck&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=freeeve_msgpck)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=freeeve_msgpck&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=freeeve_msgpck)

A high-performance msgpack library for Go, optimized for database and serialization-heavy workloads. Zero external dependencies.

## Why Another Msgpack Library?

Gives how standards proliferate vibes. `s/standards/msgpack libraries/g`

![xkcd: Standards](https://imgs.xkcd.com/comics/standards.png)

I built msgpck for [tinykvs](https://github.com/freeeve/tinykvs), a key-value store where msgpack encoding/decoding is on the hot path. 
I found issues with fuzz testing where vmihailenco would allocate excessively on big maps/arrays, and decided to check out the 
alternatives: shamaton/msgpack, hashicorp/go-msgpack, and tinylib/msgp. I didn't want code generation so tinylib/msgp was out. 
hashicorp/go-msgpack was slow. shamaton/msgpack was better but didn't compete with vmihailenco/msgpack on map[string]any performance. 
So I built msgpck focused on the common case: decoding known struct types and map[string]any with minimal allocations.

## Performance

Benchmarks vs vmihailenco/msgpack (Apple M3 Max):

### Struct Operations (using cached `GetStructEncoder`/`GetStructDecoder`)
| Operation | msgpck | allocs |
|-----------|--------|--------|
| SmallStruct Encode | 61 ns | 1 |
| MediumStruct Encode | 226 ns | 1 |
| SmallStruct Decode | 65 ns | 2 |
| SmallStruct Decode (`zeroCopy: true`) | 50 ns | 1 |
| MediumStruct Decode | 446 ns | 13 |
| MediumStruct Decode (`zeroCopy: true`) | 338 ns | 4 |

### Large Struct Operations
| Operation | msgpck | allocs |
|-----------|--------|--------|
| LargeStruct Encode (cached) | 768 ns | 9 |
| LargeStruct Decode (cached) | 1169 ns | 30 |
| LargeStruct Decode (`zeroCopy: true`) | 839 ns | 6 |

### Generic Slice (1M items - simulates database column storage)
| Operation | msgpck | allocs |
|-----------|--------|--------|
| 1M int64 Decode | 4.5 ms | 3 |
| 1M uint16 Decode | 3.4 ms | 2 |

### Map Operations
| Operation | msgpck | allocs |
|-----------|--------|--------|
| SmallMap Encode (`Marshal`) | 82 ns | 1 |
| MediumMap Encode (`Marshal`) | 251 ns | 1 |
| SmallMap Decode (`Unmarshal`) | 336 ns | 10 |
| MediumMap Decode (`Unmarshal`) | 1235 ns | 40 |
| StringMap Decode (`Unmarshal`) | 532 ns | 19 |

Run benchmarks yourself:
```bash
go test -bench=. -benchmem
```

## Quick Start

```go
import "github.com/freeeve/msgpck"

// Encode any value
data, _ := msgpck.Marshal(map[string]any{"name": "Alice", "age": 30})

// Decode to map
var m map[string]any
msgpck.Unmarshal(data, &m)

// Decode to struct
var user User
msgpck.Unmarshal(data, &user)
```

## Key Features

### Cached Struct Codecs

For hot paths, use cached codecs that avoid reflection on every call:

```go
// Get cached encoder/decoder (created on first use, reused forever)
enc := msgpck.GetStructEncoder[User]()
dec := msgpck.GetStructDecoder[User](false)

// Encode - 0 allocations with pooled buffer
data, _ := enc.Encode(&user)

// Decode - minimal allocations
var user User
dec.Decode(data, &user)
```

### Zero-Copy Mode

When your input buffer outlives the decoded result (common in databases), skip string allocations entirely:

```go
// Get cached zero-copy decoder
dec := msgpck.GetStructDecoder[User](true)

// Strings point directly into 'data' - no copies
dec.Decode(data, &user)
```

**Warning**: Zero-copy strings are only valid while the input buffer exists. Copy strings if you need them to outlive the buffer.

### Generic Type Support

The struct encoder/decoder fully supports Go generics:

```go
type SortColumn[T cmp.Ordered] struct {
    Values   []T    `msgpack:"values"`
    MaxDocID uint32 `msgpack:"max_doc_id"`
}

// Works with any concrete type
enc := msgpck.GetStructEncoder[SortColumn[int64]]()
dec := msgpck.GetStructDecoder[SortColumn[int64]](false)
```

### Supported Field Types

The cached struct codecs support:
- **Primitives**: all int/uint sizes (8/16/32/64), float32/64, bool, string
- **Slices**: `[]T` for all primitive types, `[]string`, `[]byte`
- **Maps**: `map[string]string`, `map[string]any`, nested maps
- **Nested structs**: automatically handled
- **Pointers**: pointer fields are supported

## API Reference

### Encoding

```go
// Encode any Go value to msgpack (safe to retain, concurrent-safe)
msgpck.Marshal(v any) ([]byte, error)

// For hot paths: cached struct encoder
enc := msgpck.GetStructEncoder[MyStruct]()
enc.Encode(&src)         // safe to retain (1 alloc)
enc.EncodeWith(e, &src)  // zero-alloc with your own Encoder
```

### Decoding

The API matches `encoding/json`:

```go
// Decode to any type - structs, maps, slices, primitives
var user User
msgpck.Unmarshal(data, &user)

var m map[string]any
msgpck.Unmarshal(data, &m)

var s map[string]string
msgpck.Unmarshal(data, &s)

// For hot paths: cached struct decoder
dec := msgpck.GetStructDecoder[MyStruct](false)
dec.Decode(data, &dst)

// Zero-copy cached decoder (strings point into input buffer)
dec := msgpck.GetStructDecoder[MyStruct](true)
dec.Decode(data, &dst)
```

### Timestamps

msgpck supports the msgpack timestamp extension type (-1). Times are encoded using the most compact format and decoded to UTC:

```go
// Encode a time.Time
data := msgpck.MarshalTimestamp(time.Now())

// Decode back to time.Time (UTC)
t, _ := msgpck.UnmarshalTimestamp(data)

// Streaming API
enc := msgpck.NewEncoder(nil)
enc.EncodeTimestamp(time.Now())

dec := msgpck.NewDecoder(data)
t, _ := dec.DecodeTimestamp()

// Convert extension values
ext, _ := dec.DecodeExt()
if msgpck.IsTimestamp(ext) {
    t, _ := msgpck.ExtToTimestamp(ext)
}
```

## Concurrency

All public APIs are concurrent-safe:
- `Marshal` and `Unmarshal` functions use internal pools
- `GetStructEncoder[T]()`, `GetStructDecoder[T](zeroCopy)` return cached, thread-safe codecs
- `StructEncoder` and `StructDecoder` instances are safe to use from multiple goroutines

## When to Use msgpck vs vmihailenco/msgpack

**Use msgpck when:**
- Encoding/decoding is on your hot path
- You decode the same struct types repeatedly
- You can benefit from zero-copy (database, network buffers)
- You need minimal allocations

**Use vmihailenco/msgpack when:**
- You need custom encoders/decoders for complex types
- You're decoding unknown/dynamic schemas
- Convenience matters more than raw speed

## Compatibility

msgpck produces standard msgpack bytes. Data encoded with vmihailenco/msgpack decodes correctly with msgpck and vice versa.

## License

MIT
