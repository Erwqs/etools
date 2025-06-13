package parser

import (
	"testing"
)

// Simple test structures
type SimpleAddress struct {
	Street string `etf:"street"`
	City   string `etf:"city"`
}

type SimplePerson struct {
	Name    string         `etf:"name"`
	Age     int            `etf:"age"`
	Address *SimpleAddress `etf:"address"`
}

func TestSimpleEncodeDecodeFirst(t *testing.T) {
	parser := New()
	
	original := SimplePerson{
		Name: "John Doe",
		Age:  30,
		Address: &SimpleAddress{
			Street: "123 Main St",
			City:   "Test City",
		},
	}
	
	// Encode
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	
	t.Logf("Encoded size: %d bytes", len(encoded))
	
	// Decode
	var decoded SimplePerson
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	
	// Verify basic fields
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %d, want %d", decoded.Age, original.Age)
	}
	
	// Verify nested struct
	if decoded.Address == nil {
		t.Fatal("Address is nil")
	}
	if decoded.Address.Street != original.Address.Street {
		t.Errorf("Street mismatch: got %s, want %s", decoded.Address.Street, original.Address.Street)
	}
	if decoded.Address.City != original.Address.City {
		t.Errorf("City mismatch: got %s, want %s", decoded.Address.City, original.Address.City)
	}
}
