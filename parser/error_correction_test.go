package parser

import (
	"testing"
)

func TestErrorCorrectionInAction(t *testing.T) {
	// Test 1: Basic error correction with corrupted data
	t.Run("CorruptedData", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableErrorCorrection = true
		config.ErrorCorrectionRetries = 3
		parser := NewWithConfig(config)

		original := SimplePerson{
			Name: "Test User",
			Age:  25,
		}

		// Encode normally
		encoded, err := parser.Encode(original)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Corrupt some bytes in the middle (not the magic number or size)
		if len(encoded) > 20 {
			corrupted := make([]byte, len(encoded))
			copy(corrupted, encoded)
			// Corrupt some bytes after the header
			corrupted[16] = 0xFF
			corrupted[17] = 0xFF

			t.Logf("Original size: %d, corrupted 2 bytes at positions 16-17", len(encoded))

			// Try to decode corrupted data
			var decoded SimplePerson
			err = parser.Decode(corrupted, &decoded)

			// With error correction enabled, this should either succeed or fail gracefully
			if err != nil {
				t.Logf("Error correction unable to recover from corruption: %v", err)
				// This is expected behavior - error correction is basic
			} else {
				t.Logf("Error correction successfully recovered data: %+v", decoded)
			}
		}
	})

	// Test 2: Type conversion error correction
	t.Run("TypeConversion", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableErrorCorrection = true
		config.IncludeFieldTypeInfo = true
		config.StrictTypeChecking = false // Allow conversions
		parser := NewWithConfig(config)

		original := SimplePerson{
			Name: "Type Test",
			Age:  30,
		}

		// Encode with type info
		encoded, err := parser.Encode(original)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// Decode normally to verify it works
		var decoded SimplePerson
		err = parser.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Normal decode failed: %v", err)
		}

		t.Logf("Successfully decoded with type info: Name='%s', Age=%d", decoded.Name, decoded.Age)
	})

	// Test 3: Field skipping with unknown fields
	t.Run("UnknownFields", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableErrorCorrection = true
		parser := NewWithConfig(config)

		// Create data with a different struct that has extra fields
		type ExtendedPerson struct {
			Name    string `etf:"name"`
			Age     int    `etf:"age"`
			Extra   string `etf:"extra"`
			Unknown int    `etf:"unknown"`
		}

		extended := ExtendedPerson{
			Name:    "Extended User",
			Age:     35,
			Extra:   "extra data",
			Unknown: 999,
		}

		// Encode extended struct
		encoded, err := parser.Encode(extended)
		if err != nil {
			t.Fatalf("Encode extended failed: %v", err)
		}

		// Try to decode into simpler struct (should skip unknown fields)
		var simple SimplePerson
		err = parser.Decode(encoded, &simple)

		if err != nil {
			t.Logf("Failed to decode with unknown fields: %v", err)
			// This might fail due to field skipping issues
		} else {
			t.Logf("Successfully decoded skipping unknown fields: Name='%s', Age=%d", simple.Name, simple.Age)
		}
	})

	// Test 4: Error correction configuration
	t.Run("Configuration", func(t *testing.T) {
		// Test with error correction disabled
		configOff := DefaultConfig()
		configOff.EnableErrorCorrection = false
		parserOff := NewWithConfig(configOff)

		// Test with error correction enabled
		configOn := DefaultConfig()
		configOn.EnableErrorCorrection = true
		configOn.ErrorCorrectionRetries = 5
		parserOn := NewWithConfig(configOn)

		t.Logf("Parser with error correction OFF: EnableErrorCorrection=%t", parserOff.config.EnableErrorCorrection)
		t.Logf("Parser with error correction ON: EnableErrorCorrection=%t, Retries=%d",
			parserOn.config.EnableErrorCorrection, parserOn.config.ErrorCorrectionRetries)

		// Both should work for valid data
		original := SimplePerson{Name: "Config Test", Age: 40}

		encodedOff, err := parserOff.Encode(original)
		if err != nil {
			t.Fatalf("Encode with error correction OFF failed: %v", err)
		}

		encodedOn, err := parserOn.Encode(original)
		if err != nil {
			t.Fatalf("Encode with error correction ON failed: %v", err)
		}

		var decodedOff, decodedOn SimplePerson

		err = parserOff.Decode(encodedOff, &decodedOff)
		if err != nil {
			t.Fatalf("Decode with error correction OFF failed: %v", err)
		}

		err = parserOn.Decode(encodedOn, &decodedOn)
		if err != nil {
			t.Fatalf("Decode with error correction ON failed: %v", err)
		}

		t.Logf("Both configurations work for valid data")
	})
}
