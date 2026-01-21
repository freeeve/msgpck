package msgpck

// Test error message constants to avoid duplication
const (
	errMsgUnexpectedEOF = "expected ErrUnexpectedEOF, got %v"
	errMsgExpectedEOF   = "expected ErrUnexpectedEOF"
	errMsgEOFError      = "expected EOF error"
	errMsgStringTooLong = "expected ErrStringTooLong, got %v"
	errMsgTypeMismatch  = "expected ErrTypeMismatch, got %v"
	errMsgUnexpectedErr = "unexpected error: %v"
	errMsgExpectedErr   = "expected error"
	errMsgGot42         = "got %d, want 42"
	errMsgBinaryTooLong = "expected ErrBinaryTooLong, got %v"
	errMsgArrayTooLong  = "expected ErrArrayTooLong, got %v"
	errMsgMapTooLong    = "expected ErrMapTooLong, got %v"

	// Common test data
	testEmail = "alice@example.com"

	// Common error format strings
	errMsgMarshalFailed   = "Marshal failed: %v"
	errMsgUnmarshalFailed = "Unmarshal failed: %v"
	errMsgStructFailed    = "UnmarshalStruct failed: %v"
	errMsgDecodeFailed    = "decode failed: %v"
	errMsgGotWant         = "got %v, want %v"
	errMsgGotWantStruct   = "got %+v, want %+v"
	errMsgGotWantStr      = "got %s, want %s"
	errMsgFmtSV           = "%s: %v"
	errMsgExpectedAlice   = "expected 'Alice', got %q"
	errMsgNilFmt          = "expected nil format, got 0x%02x"
	errMsgGotLen          = "got len=%d, want %d"
)
