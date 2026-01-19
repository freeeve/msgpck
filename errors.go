package msgpck

import "errors"

var (
	// ErrUnexpectedEOF is returned when the input ends before a complete value
	ErrUnexpectedEOF = errors.New("msgpck: unexpected end of input")

	// ErrInvalidFormat is returned when an invalid format byte is encountered
	ErrInvalidFormat = errors.New("msgpck: invalid format byte")

	// ErrStringTooLong is returned when a string exceeds MaxStringLen
	ErrStringTooLong = errors.New("msgpck: string exceeds maximum length")

	// ErrBinaryTooLong is returned when binary data exceeds MaxBinaryLen
	ErrBinaryTooLong = errors.New("msgpck: binary data exceeds maximum length")

	// ErrArrayTooLong is returned when an array exceeds MaxArrayLen
	ErrArrayTooLong = errors.New("msgpck: array exceeds maximum length")

	// ErrMapTooLong is returned when a map exceeds MaxMapLen
	ErrMapTooLong = errors.New("msgpck: map exceeds maximum length")

	// ErrExtTooLong is returned when ext data exceeds MaxExtLen
	ErrExtTooLong = errors.New("msgpck: ext data exceeds maximum length")

	// ErrMaxDepthExceeded is returned when nesting exceeds MaxDepth
	ErrMaxDepthExceeded = errors.New("msgpck: maximum nesting depth exceeded")

	// ErrTypeMismatch is returned when decoding into an incompatible type
	ErrTypeMismatch = errors.New("msgpck: type mismatch")

	// ErrNotPointer is returned when DecodeStruct is called with a non-pointer
	ErrNotPointer = errors.New("msgpck: decode target must be a pointer")

	// ErrNotStruct is returned when DecodeStruct is called with a non-struct pointer
	ErrNotStruct = errors.New("msgpck: decode target must be a pointer to struct")

	// ErrBufferTooSmall is returned when the encode buffer is too small
	ErrBufferTooSmall = errors.New("msgpck: buffer too small")

	// ErrUnsupportedType is returned when encoding an unsupported Go type
	ErrUnsupportedType = errors.New("msgpck: unsupported type")
)
