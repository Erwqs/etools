package parser

import (
	"testing"
)

func TestMagicNumber(t *testing.T) {
	parser := New()

	original := SimplePerson{
		Name: "Test User",
		Age:  25,
		Address: &SimpleAddress{
			Street: "Test Street",
			City:   "Test City",
		},
	}

	// Encode
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify magic number is at the beginning
	if len(encoded) < 4 {
		t.Fatal("Encoded data too short to contain magic number")
	}

	// Check first 4 bytes represent our magic number (in little-endian)
	expectedBytes := []byte{0xC0, 0x0E, 0x9E, 0x0C} // 0x0C9E0EC0 in little-endian
	actualBytes := encoded[:4]

	for i := 0; i < 4; i++ {
		if actualBytes[i] != expectedBytes[i] {
			t.Errorf("Magic number byte %d: expected 0x%02X, got 0x%02X", i, expectedBytes[i], actualBytes[i])
		}
	}

	t.Logf("Magic number verified: %02X %02X %02X %02X", actualBytes[0], actualBytes[1], actualBytes[2], actualBytes[3])

	// Test that decode still works
	var decoded SimplePerson
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name mismatch after decode: got %s, want %s", decoded.Name, original.Name)
	}
}

func TestInvalidMagicNumber(t *testing.T) {
	parser := New()

	// Create some fake data with wrong magic number
	fakeData := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x10, 0x00, 0x00, 0x00} // wrong magic + size

	var result SimplePerson
	err := parser.Decode(fakeData, &result)
	if err == nil {
		t.Fatal("Expected error for invalid magic number, but got none")
	}

	expectedErrMsg := "invalid magic number"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("Expected error message to start with '%s', got: %s", expectedErrMsg, err.Error())
	}

	t.Logf("Correctly rejected invalid magic number: %v", err)
}
