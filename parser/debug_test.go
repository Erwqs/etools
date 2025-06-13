package parser

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDebugEncodeDecodeFlow(t *testing.T) {
	parser := New()

	original := SimplePerson{
		Name: "John",
		Age:  30,
		Address: &SimpleAddress{
			Street: "Main St",
			City:   "City",
		},
	}

	// Encode step by step
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	t.Logf("Encoded data length: %d bytes", len(encoded))
	t.Logf("First 32 bytes: %x", encoded[:min(32, len(encoded))])

	// Let's manually inspect the structure
	buf := bytes.NewReader(encoded)

	// Read original size
	var origSize uint32
	err = binary.Read(buf, binary.LittleEndian, &origSize)
	if err != nil {
		t.Fatalf("Failed to read original size: %v", err)
	}
	t.Logf("Original size from header: %d", origSize)
	t.Logf("Remaining after size read: %d", buf.Len())

	// Check compression flag
	if buf.Len() > int(origSize) {
		var compFlag uint8
		err = binary.Read(buf, binary.LittleEndian, &compFlag)
		if err != nil {
			t.Fatalf("Failed to read compression flag: %v", err)
		}
		t.Logf("Compression flag: %d", compFlag)
		t.Logf("Remaining after compression flag: %d", buf.Len())
	}

	// Now try to decode normally
	var decoded SimplePerson
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	t.Logf("Successfully decoded: %+v", decoded)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
