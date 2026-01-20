package msgpck

// Test error message constants to avoid duplication
const (
	errMsgUnexpectedEOF = "expected ErrUnexpectedEOF, got %v"
	errMsgEOFError      = "expected EOF error"
	errMsgStringTooLong = "expected ErrStringTooLong, got %v"
	errMsgTypeMismatch  = "expected ErrTypeMismatch, got %v"
	errMsgUnexpectedErr = "unexpected error: %v"
	errMsgExpectedErr   = "expected error"
	errMsgGot42         = "got %d, want 42"
	errMsgBinaryTooLong = "expected ErrBinaryTooLong, got %v"
	errMsgArrayTooLong  = "expected ErrArrayTooLong, got %v"
	errMsgMapTooLong    = "expected ErrMapTooLong, got %v"
)
