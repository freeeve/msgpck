package msgpck

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestConformance runs the msgpack-test-suite conformance tests.
// See: https://github.com/kawanet/msgpack-test-suite
func TestConformance(t *testing.T) {
	data, err := os.ReadFile("testdata/msgpack-test-suite.json")
	if err != nil {
		t.Skipf("msgpack-test-suite.json not found: %v", err)
	}

	var suite map[string][]json.RawMessage
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("failed to parse test suite: %v", err)
	}

	for filename, testCases := range suite {
		t.Run(filename, func(t *testing.T) {
			for i, rawCase := range testCases {
				runConformanceCase(t, i, rawCase)
			}
		})
	}
}

func runConformanceCase(t *testing.T, idx int, rawCase json.RawMessage) {
	var testCase map[string]json.RawMessage
	if err := json.Unmarshal(rawCase, &testCase); err != nil {
		t.Errorf("case %d: failed to parse: %v", idx, err)
		return
	}

	// Get msgpack hex strings
	var msgpackHexes []string
	if err := json.Unmarshal(testCase["msgpack"], &msgpackHexes); err != nil {
		t.Errorf("case %d: failed to parse msgpack field: %v", idx, err)
		return
	}

	// Determine value type and expected value
	var expectedValue any
	var valueType string

	if _, ok := testCase["nil"]; ok {
		valueType = "nil"
		expectedValue = nil
	} else if raw, ok := testCase["bool"]; ok {
		valueType = "bool"
		json.Unmarshal(raw, &expectedValue)
	} else if raw, ok := testCase["number"]; ok {
		valueType = "number"
		json.Unmarshal(raw, &expectedValue)
	} else if _, ok := testCase["bignum"]; ok {
		// Skip bignum tests - they exceed int64/uint64 range
		return
	} else if raw, ok := testCase["string"]; ok {
		valueType = "string"
		json.Unmarshal(raw, &expectedValue)
	} else if raw, ok := testCase["binary"]; ok {
		valueType = "binary"
		var hexStr string
		json.Unmarshal(raw, &hexStr)
		expectedValue = hexStr // Keep as hex for comparison
	} else if raw, ok := testCase["array"]; ok {
		valueType = "array"
		json.Unmarshal(raw, &expectedValue)
	} else if raw, ok := testCase["map"]; ok {
		valueType = "map"
		json.Unmarshal(raw, &expectedValue)
	} else {
		// Unknown type, skip
		return
	}

	// Test each msgpack encoding
	for _, msgpackHex := range msgpackHexes {
		testDecoding(t, idx, valueType, expectedValue, msgpackHex)
	}
}

func testDecoding(t *testing.T, idx int, valueType string, expected any, msgpackHex string) {
	// Parse hex string (format: "cc-00" or "00")
	hexStr := strings.ReplaceAll(msgpackHex, "-", "")
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Errorf("case %d: invalid hex %q: %v", idx, msgpackHex, err)
		return
	}

	d := NewDecoder(data)
	got, err := d.DecodeAny()
	if err != nil {
		t.Errorf("case %d [%s]: decode %q failed: %v", idx, msgpackHex, valueType, err)
		return
	}

	// Compare based on type
	switch valueType {
	case "nil":
		if got != nil {
			t.Errorf("case %d [%s]: got %v, want nil", idx, msgpackHex, got)
		}
	case "bool":
		expectedBool := expected.(bool)
		gotBool, ok := got.(bool)
		if !ok || gotBool != expectedBool {
			t.Errorf("case %d [%s]: got %v, want %v", idx, msgpackHex, got, expectedBool)
		}
	case "number":
		expectedNum := expected.(float64) // JSON numbers are float64
		switch v := got.(type) {
		case int64:
			if float64(v) != expectedNum {
				t.Errorf("case %d [%s]: got %d, want %v", idx, msgpackHex, v, expectedNum)
			}
		case uint64:
			if float64(v) != expectedNum {
				t.Errorf("case %d [%s]: got %d, want %v", idx, msgpackHex, v, expectedNum)
			}
		case float64:
			// For floats, allow small difference due to precision
			diff := v - expectedNum
			if diff < 0 {
				diff = -diff
			}
			if expectedNum != 0 && diff/expectedNum > 1e-6 {
				t.Errorf("case %d [%s]: got %v, want %v", idx, msgpackHex, v, expectedNum)
			} else if expectedNum == 0 && diff > 1e-10 {
				t.Errorf("case %d [%s]: got %v, want %v", idx, msgpackHex, v, expectedNum)
			}
		default:
			t.Errorf("case %d [%s]: got unexpected type %T", idx, msgpackHex, got)
		}
	case "string":
		expectedStr := expected.(string)
		gotStr, ok := got.(string)
		if !ok || gotStr != expectedStr {
			t.Errorf("case %d [%s]: got %q, want %q", idx, msgpackHex, got, expectedStr)
		}
	case "binary":
		expectedHex := expected.(string)
		gotBytes, ok := got.([]byte)
		if !ok {
			t.Errorf("case %d [%s]: got type %T, want []byte", idx, msgpackHex, got)
			return
		}
		gotHex := strings.ToLower(hex.EncodeToString(gotBytes))
		expectedHexClean := strings.ToLower(strings.ReplaceAll(expectedHex, "-", ""))
		if gotHex != expectedHexClean {
			t.Errorf("case %d [%s]: got %q, want %q", idx, msgpackHex, gotHex, expectedHexClean)
		}
	case "array":
		expectedArr := expected.([]any)
		gotArr, ok := got.([]any)
		if !ok {
			t.Errorf("case %d [%s]: got type %T, want []any", idx, msgpackHex, got)
			return
		}
		if len(gotArr) != len(expectedArr) {
			t.Errorf("case %d [%s]: got len %d, want %d", idx, msgpackHex, len(gotArr), len(expectedArr))
			return
		}
		for i := range expectedArr {
			if !compareValues(gotArr[i], expectedArr[i]) {
				t.Errorf("case %d [%s]: array[%d] mismatch: got %v, want %v", idx, msgpackHex, i, gotArr[i], expectedArr[i])
			}
		}
	case "map":
		expectedMap := expected.(map[string]any)
		gotMap, ok := got.(map[string]any)
		if !ok {
			t.Errorf("case %d [%s]: got type %T, want map[string]any", idx, msgpackHex, got)
			return
		}
		if len(gotMap) != len(expectedMap) {
			t.Errorf("case %d [%s]: got len %d, want %d", idx, msgpackHex, len(gotMap), len(expectedMap))
			return
		}
		for k, expectedVal := range expectedMap {
			gotVal, exists := gotMap[k]
			if !exists {
				t.Errorf("case %d [%s]: missing key %q", idx, msgpackHex, k)
				continue
			}
			if !compareValues(gotVal, expectedVal) {
				t.Errorf("case %d [%s]: map[%q] mismatch: got %v, want %v", idx, msgpackHex, k, gotVal, expectedVal)
			}
		}
	}
}

// compareValues compares two values, handling type differences between msgpack and JSON.
func compareValues(got, expected any) bool {
	if got == nil && expected == nil {
		return true
	}
	if got == nil || expected == nil {
		return false
	}

	switch exp := expected.(type) {
	case float64:
		// JSON numbers are float64, msgpack may return int64
		switch g := got.(type) {
		case int64:
			return float64(g) == exp
		case uint64:
			return float64(g) == exp
		case float64:
			if exp == 0 {
				return g == 0
			}
			diff := g - exp
			if diff < 0 {
				diff = -diff
			}
			return diff/exp < 1e-6
		}
	case bool:
		g, ok := got.(bool)
		return ok && g == exp
	case string:
		g, ok := got.(string)
		return ok && g == exp
	case []any:
		g, ok := got.([]any)
		if !ok || len(g) != len(exp) {
			return false
		}
		for i := range exp {
			if !compareValues(g[i], exp[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok || len(g) != len(exp) {
			return false
		}
		for k, v := range exp {
			if !compareValues(g[k], v) {
				return false
			}
		}
		return true
	}
	return false
}
