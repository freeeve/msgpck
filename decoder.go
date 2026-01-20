package msgpck

import (
	"encoding/binary"
	"math"
)

// Decoder reads MessagePack data from a byte slice.
// It provides zero-copy decoding where possible.
type Decoder struct {
	data  []byte
	pos   int
	cfg   Config
	depth int
}

// NewDecoder creates a new Decoder for the given data with default config
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data: data,
		cfg:  DefaultConfig(),
	}
}

// NewDecoderWithConfig creates a new Decoder with custom config
func NewDecoderWithConfig(data []byte, cfg Config) *Decoder {
	return &Decoder{
		data: data,
		cfg:  cfg,
	}
}

// Reset resets the decoder to decode new data
func (d *Decoder) Reset(data []byte) {
	d.data = data
	d.pos = 0
	d.depth = 0
}

// Remaining returns the number of unread bytes
func (d *Decoder) Remaining() int {
	return len(d.data) - d.pos
}

// Position returns the current position in the data
func (d *Decoder) Position() int {
	return d.pos
}

// hasBytes returns true if there are at least n bytes remaining
func (d *Decoder) hasBytes(n int) bool {
	return d.pos+n <= len(d.data)
}

// readByte reads a single byte
func (d *Decoder) readByte() (byte, error) {
	if d.pos >= len(d.data) {
		return 0, ErrUnexpectedEOF
	}
	b := d.data[d.pos]
	d.pos++
	return b, nil
}

// peekByte returns the next byte without consuming it
func (d *Decoder) peekByte() (byte, error) {
	if d.pos >= len(d.data) {
		return 0, ErrUnexpectedEOF
	}
	return d.data[d.pos], nil
}

// readBytes reads n bytes and returns a slice into the source
func (d *Decoder) readBytes(n int) ([]byte, error) {
	if !d.hasBytes(n) {
		return nil, ErrUnexpectedEOF
	}
	b := d.data[d.pos : d.pos+n]
	d.pos += n
	return b, nil
}

// readUint8 reads a uint8
func (d *Decoder) readUint8() (uint8, error) {
	b, err := d.readByte()
	return b, err
}

// readUint16 reads a big-endian uint16
func (d *Decoder) readUint16() (uint16, error) {
	if !d.hasBytes(2) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.BigEndian.Uint16(d.data[d.pos:])
	d.pos += 2
	return v, nil
}

// readUint32 reads a big-endian uint32
func (d *Decoder) readUint32() (uint32, error) {
	if !d.hasBytes(4) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.BigEndian.Uint32(d.data[d.pos:])
	d.pos += 4
	return v, nil
}

// readUint64 reads a big-endian uint64
func (d *Decoder) readUint64() (uint64, error) {
	if !d.hasBytes(8) {
		return 0, ErrUnexpectedEOF
	}
	v := binary.BigEndian.Uint64(d.data[d.pos:])
	d.pos += 8
	return v, nil
}

// readInt8 reads an int8
func (d *Decoder) readInt8() (int8, error) {
	b, err := d.readByte()
	return int8(b), err
}

// readInt16 reads a big-endian int16
func (d *Decoder) readInt16() (int16, error) {
	v, err := d.readUint16()
	return int16(v), err
}

// readInt32 reads a big-endian int32
func (d *Decoder) readInt32() (int32, error) {
	v, err := d.readUint32()
	return int32(v), err
}

// readInt64 reads a big-endian int64
func (d *Decoder) readInt64() (int64, error) {
	v, err := d.readUint64()
	return int64(v), err
}

// readFloat32 reads a big-endian float32
func (d *Decoder) readFloat32() (float32, error) {
	v, err := d.readUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(v), nil
}

// readFloat64 reads a big-endian float64
func (d *Decoder) readFloat64() (float64, error) {
	v, err := d.readUint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(v), nil
}

// enterContainer increments depth and checks MaxDepth
func (d *Decoder) enterContainer() error {
	d.depth++
	if d.depth > d.cfg.MaxDepth {
		return ErrMaxDepthExceeded
	}
	return nil
}

// leaveContainer decrements depth
func (d *Decoder) leaveContainer() {
	d.depth--
}

// validateStringLen checks if length is valid for a string
func (d *Decoder) validateStringLen(length int) error {
	if length > d.cfg.MaxStringLen {
		return ErrStringTooLong
	}
	if !d.hasBytes(length) {
		return ErrUnexpectedEOF
	}
	return nil
}

// validateBinaryLen checks if length is valid for binary data
func (d *Decoder) validateBinaryLen(length int) error {
	if length > d.cfg.MaxBinaryLen {
		return ErrBinaryTooLong
	}
	if !d.hasBytes(length) {
		return ErrUnexpectedEOF
	}
	return nil
}

// validateArrayLen checks if length is valid for an array
func (d *Decoder) validateArrayLen(length int) error {
	if length > d.cfg.MaxArrayLen {
		return ErrArrayTooLong
	}
	// Sanity check: each element needs at least 1 byte
	if !d.hasBytes(length) {
		return ErrUnexpectedEOF
	}
	return nil
}

// validateMapLen checks if length is valid for a map
func (d *Decoder) validateMapLen(length int) error {
	if length > d.cfg.MaxMapLen {
		return ErrMapTooLong
	}
	// Sanity check: each key-value pair needs at least 2 bytes
	if !d.hasBytes(length * 2) {
		return ErrUnexpectedEOF
	}
	return nil
}

// validateExtLen checks if length is valid for ext data
func (d *Decoder) validateExtLen(length int) error {
	if length > d.cfg.MaxExtLen {
		return ErrExtTooLong
	}
	if !d.hasBytes(length) {
		return ErrUnexpectedEOF
	}
	return nil
}

// readStringBytes reads a string and returns raw bytes (zero-copy).
// Used by struct decoders and typed decode functions.
func (d *Decoder) readStringBytes() ([]byte, error) {
	format, err := d.readByte()
	if err != nil {
		return nil, err
	}

	var length int
	if isFixstr(format) {
		length = fixstrLen(format)
	} else {
		switch format {
		case formatStr8:
			n, err := d.readUint8()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr16:
			n, err := d.readUint16()
			if err != nil {
				return nil, err
			}
			length = int(n)
		case formatStr32:
			n, err := d.readUint32()
			if err != nil {
				return nil, err
			}
			length = int(n)
		default:
			return nil, ErrTypeMismatch
		}
	}

	if err := d.validateStringLen(length); err != nil {
		return nil, err
	}
	return d.readBytes(length)
}
