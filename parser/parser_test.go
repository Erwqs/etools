package parser

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

// Test structures for comprehensive testing
type Address struct {
	Street  string `etf:"street"`
	City    string `etf:"city"`
	ZipCode int    `etf:"zipcode"`
}

type Person struct {
	Name    string         `etf:"name"`
	Age     int            `etf:"age"`
	Active  bool           `etf:"active"`
	Height  float64        `etf:"height"`
	Address *Address       `etf:"address"`
	Tags    []string       `etf:"tags"`
	Scores  map[string]int `etf:"scores"`
}

type Company struct {
	Name      string             `etf:"name"`
	Founded   int                `etf:"founded"`
	Employees []Person           `etf:"employees"`
	Locations map[string]Address `etf:"locations"`
	CEO       *Person            `etf:"ceo"`
}

func TestBasicEncodeDecoder(t *testing.T) {
	parser := New()

	// Test data
	original := Person{
		Name:   "John Doe",
		Age:    30,
		Active: true,
		Height: 5.9,
		Address: &Address{
			Street:  "123 Main St",
			City:    "Anytown",
			ZipCode: 12345,
		},
		Tags:   []string{"developer", "gopher"},
		Scores: map[string]int{"math": 95, "science": 87},
	}

	// Encode
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded Person
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %v, want %v", decoded.Age, original.Age)
	}
	if decoded.Active != original.Active {
		t.Errorf("Active mismatch: got %v, want %v", decoded.Active, original.Active)
	}
	if decoded.Height != original.Height {
		t.Errorf("Height mismatch: got %v, want %v", decoded.Height, original.Height)
	}

	// Check nested struct
	if decoded.Address == nil {
		t.Fatal("Address is nil")
	}
	if decoded.Address.Street != original.Address.Street {
		t.Errorf("Address.Street mismatch: got %v, want %v", decoded.Address.Street, original.Address.Street)
	}
	if decoded.Address.City != original.Address.City {
		t.Errorf("Address.City mismatch: got %v, want %v", decoded.Address.City, original.Address.City)
	}
	if decoded.Address.ZipCode != original.Address.ZipCode {
		t.Errorf("Address.ZipCode mismatch: got %v, want %v", decoded.Address.ZipCode, original.Address.ZipCode)
	}

	// Check slice
	if !reflect.DeepEqual(decoded.Tags, original.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", decoded.Tags, original.Tags)
	}

	// Check map
	if !reflect.DeepEqual(decoded.Scores, original.Scores) {
		t.Errorf("Scores mismatch: got %v, want %v", decoded.Scores, original.Scores)
	}
}

func TestConfigurationOptions(t *testing.T) {
	// Test with different configurations
	configs := []*Config{
		{
			CompressionLevel: CompressionNone,
			IncludeTypeInfo:  false,
			MaxDepth:         50,
			TagName:          "etf",
		},
		{
			CompressionLevel: CompressionFast,
			IncludeTypeInfo:  true,
			MaxDepth:         100,
			TagName:          "etf",
		},
	}

	original := Person{
		Name:   "Test User",
		Age:    25,
		Active: true,
		Address: &Address{
			Street: "Test Street",
			City:   "Test City",
		},
	}

	for i, config := range configs {
		parser := NewWithConfig(config)

		encoded, err := parser.Encode(original)
		if err != nil {
			t.Fatalf("Config %d: Encode failed: %v", i, err)
		}

		var decoded Person
		err = parser.Decode(encoded, &decoded)
		if err != nil {
			t.Fatalf("Config %d: Decode failed: %v", i, err)
		}

		if decoded.Name != original.Name {
			t.Errorf("Config %d: Name mismatch", i)
		}
	}
}

func TestNilPointers(t *testing.T) {
	config := DefaultConfig()
	config.SkipNilPointers = true
	parser := NewWithConfig(config)

	original := Person{
		Name:    "Test",
		Age:     30,
		Address: nil, // nil pointer
		Tags:    []string{"test"},
		Scores:  make(map[string]int),
	}

	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded Person
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Address != nil {
		t.Error("Expected nil address, got non-nil")
	}
}

func TestValidation(t *testing.T) {
	parser := New()

	valid := Person{Name: "Test", Age: 30}
	err := parser.ValidateStruct(valid)
	if err != nil {
		t.Errorf("ValidateStruct failed for valid struct: %v", err)
	}

	// Test with non-struct
	err = parser.ValidateStruct("not a struct")
	if err == nil {
		t.Error("ValidateStruct should fail for non-struct")
	}
}

func TestSizeEstimation(t *testing.T) {
	parser := New()

	data := Person{
		Name: "Test User",
		Age:  30,
		Tags: []string{"tag1", "tag2"},
	}

	size, err := parser.EstimateSize(data)
	if err != nil {
		t.Fatalf("EstimateSize failed: %v", err)
	}

	if size <= 0 {
		t.Error("EstimateSize should return positive size")
	}

	// Encode and compare
	encoded, err := parser.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Size estimate should be reasonably close (within 2x)
	if size < len(encoded)/2 || size > len(encoded)*2 {
		t.Errorf("Size estimate %d too far from actual size %d", size, len(encoded))
	}
}

func TestNestedStructWithFieldTypeInfo(t *testing.T) {
	// Configure parser with field type information
	config := DefaultConfig()
	config.IncludeFieldTypeInfo = true
	config.IncludeChecksum = true
	config.EnableErrorCorrection = true
	parser := NewWithConfig(config)

	// Create test data with deeply nested structures
	ceo := &Person{
		Name: "Jane Smith",
		Age:  45,
		Address: &Address{
			Street:  "456 Executive Blvd",
			City:    "Corporate City",
			ZipCode: 54321,
		},
		Tags:   []string{"CEO", "Leader"},
		Scores: map[string]int{"leadership": 100, "vision": 95},
	}

	original := Company{
		Name:    "TechCorp",
		Founded: 2010,
		CEO:     ceo,
		Employees: []Person{
			{
				Name: "Alice Johnson",
				Age:  28,
				Address: &Address{
					Street:  "789 Developer Ave",
					City:    "Code City",
					ZipCode: 98765,
				},
				Tags:   []string{"developer", "backend"},
				Scores: map[string]int{"coding": 90, "testing": 85},
			},
			{
				Name: "Bob Wilson",
				Age:  32,
				Address: &Address{
					Street:  "321 Programmer St",
					City:    "Binary Town",
					ZipCode: 13579,
				},
				Tags:   []string{"frontend", "designer"},
				Scores: map[string]int{"ui": 88, "ux": 92},
			},
		},
		Locations: map[string]Address{
			"headquarters": {
				Street:  "100 Corporate Way",
				City:    "Business City",
				ZipCode: 11111,
			},
			"branch": {
				Street:  "200 Branch Rd",
				City:    "Satellite City",
				ZipCode: 22222,
			},
		},
	}

	// Encode
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Decode
	var decoded Company
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify company info
	if decoded.Name != original.Name {
		t.Errorf("Company name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Founded != original.Founded {
		t.Errorf("Founded year mismatch: got %d, want %d", decoded.Founded, original.Founded)
	}

	// Verify CEO pointer and nested data
	if decoded.CEO == nil {
		t.Error("CEO should not be nil")
	} else {
		if decoded.CEO.Name != original.CEO.Name {
			t.Errorf("CEO name mismatch: got %s, want %s", decoded.CEO.Name, original.CEO.Name)
		}
		if decoded.CEO.Age != original.CEO.Age {
			t.Errorf("CEO age mismatch: got %d, want %d", decoded.CEO.Age, original.CEO.Age)
		}

		// Verify CEO's nested address
		if decoded.CEO.Address == nil {
			t.Error("CEO address should not be nil")
		} else {
			if decoded.CEO.Address.Street != original.CEO.Address.Street {
				t.Errorf("CEO address street mismatch: got %s, want %s",
					decoded.CEO.Address.Street, original.CEO.Address.Street)
			}
		}

		// Verify CEO's scores map
		for key, value := range original.CEO.Scores {
			if decodedValue, exists := decoded.CEO.Scores[key]; !exists {
				t.Errorf("CEO score %s missing", key)
			} else if decodedValue != value {
				t.Errorf("CEO score %s mismatch: got %d, want %d", key, decodedValue, value)
			}
		}
	}

	// Verify employees slice with nested data
	if len(decoded.Employees) != len(original.Employees) {
		t.Errorf("Employees count mismatch: got %d, want %d", len(decoded.Employees), len(original.Employees))
	} else {
		for i, emp := range original.Employees {
			if decoded.Employees[i].Name != emp.Name {
				t.Errorf("Employee[%d] name mismatch: got %s, want %s", i, decoded.Employees[i].Name, emp.Name)
			}

			// Verify employee address
			if decoded.Employees[i].Address == nil && emp.Address != nil {
				t.Errorf("Employee[%d] address should not be nil", i)
			} else if decoded.Employees[i].Address != nil && emp.Address != nil {
				if decoded.Employees[i].Address.City != emp.Address.City {
					t.Errorf("Employee[%d] city mismatch: got %s, want %s",
						i, decoded.Employees[i].Address.City, emp.Address.City)
				}
			}

			// Verify employee tags slice
			if len(decoded.Employees[i].Tags) != len(emp.Tags) {
				t.Errorf("Employee[%d] tags count mismatch: got %d, want %d",
					i, len(decoded.Employees[i].Tags), len(emp.Tags))
			}
		}
	}

	// Verify locations map
	if len(decoded.Locations) != len(original.Locations) {
		t.Errorf("Locations count mismatch: got %d, want %d", len(decoded.Locations), len(original.Locations))
	} else {
		for key, loc := range original.Locations {
			if decodedLoc, exists := decoded.Locations[key]; !exists {
				t.Errorf("Location %s missing", key)
			} else {
				if decodedLoc.Street != loc.Street {
					t.Errorf("Location[%s] street mismatch: got %s, want %s", key, decodedLoc.Street, loc.Street)
				}
			}
		}
	}
}

func TestDifferentCompressionLevels(t *testing.T) {
	original := Person{
		Name:   "Compression Test",
		Age:    25,
		Active: true,
		Height: 5.8,
		Address: &Address{
			Street:  "123 Compression St",
			City:    "Data City",
			ZipCode: 12345,
		},
		Tags:   []string{"compression", "test", "lz4"},
		Scores: map[string]int{"efficiency": 95, "speed": 90},
	}

	compressionLevels := []CompressionLevel{
		CompressionNone,
		CompressionFast,
		CompressionHigh,
	}

	for _, level := range compressionLevels {
		t.Run(fmt.Sprintf("Compression_%d", level), func(t *testing.T) {
			config := DefaultConfig()
			config.CompressionLevel = level
			parser := NewWithConfig(config)

			// Encode
			encoded, err := parser.Encode(original)
			if err != nil {
				t.Fatalf("Failed to encode with compression level %d: %v", level, err)
			}

			// Decode
			var decoded Person
			err = parser.Decode(encoded, &decoded)
			if err != nil {
				t.Fatalf("Failed to decode with compression level %d: %v", level, err)
			}

			// Verify data integrity
			if decoded.Name != original.Name {
				t.Errorf("Name mismatch with compression %d: got %s, want %s", level, decoded.Name, original.Name)
			}
			if decoded.Age != original.Age {
				t.Errorf("Age mismatch with compression %d: got %d, want %d", level, decoded.Age, original.Age)
			}

			t.Logf("Compression level %d: encoded size = %d bytes", level, len(encoded))
		})
	}
}

func TestErrorCorrectionFeatures(t *testing.T) {
	config := DefaultConfig()
	config.EnableErrorCorrection = true
	config.ErrorCorrectionRetries = 3
	config.IncludeFieldTypeInfo = true
	config.StrictTypeChecking = false // Allow type conversions
	parser := NewWithConfig(config)

	original := Person{
		Name:   "Error Test",
		Age:    30,
		Active: true,
		Address: &Address{
			Street:  "Error Recovery St",
			City:    "Correction City",
			ZipCode: 99999,
		},
	}

	// Test basic encode/decode works
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	var decoded Person
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify data
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Address.ZipCode != original.Address.ZipCode {
		t.Errorf("ZipCode mismatch: got %d, want %d", decoded.Address.ZipCode, original.Address.ZipCode)
	}
}

func TestValidationAndEstimation(t *testing.T) {
	parser := New()

	// Test validation
	validPerson := Person{
		Name: "Valid User",
		Age:  30,
		Address: &Address{
			Street: "123 Valid St",
			City:   "Valid City",
		},
		Tags: []string{"valid", "test"},
	}

	err := parser.ValidateStruct(validPerson)
	if err != nil {
		t.Errorf("Valid struct should pass validation: %v", err)
	}

	// Test with pointer
	err = parser.ValidateStruct(&validPerson)
	if err != nil {
		t.Errorf("Valid struct pointer should pass validation: %v", err)
	}

	// Test size estimation
	estimatedSize, err := parser.EstimateSize(validPerson)
	if err != nil {
		t.Fatalf("Failed to estimate size: %v", err)
	}

	// Encode to get actual size
	encoded, err := parser.Encode(validPerson)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	actualSize := len(encoded)

	// Estimated size should be in a reasonable range of actual size
	if estimatedSize < actualSize/3 || estimatedSize > actualSize*5 {
		t.Errorf("Size estimation seems off: estimated %d, actual %d", estimatedSize, actualSize)
	}

	t.Logf("Estimated size: %d, Actual size: %d", estimatedSize, actualSize)
}

func TestParserConfiguration(t *testing.T) {
	// Test custom configuration
	config := &Config{
		CompressionLevel:       CompressionHigh,
		IncludeTypeInfo:        false,
		IncludeFieldTypeInfo:   true,
		SortFields:             true,
		MaxDepth:               50,
		SkipNilPointers:        false,
		TagName:                "custom",
		StrictTypeChecking:     true,
		IncludeChecksum:        false,
		EnableErrorCorrection:  false,
		ErrorCorrectionRetries: 1,
		IncludeFieldVersions:   true,
	}

	parser := NewWithConfig(config)

	// Verify configuration was applied
	retrievedConfig := parser.GetConfig()
	if retrievedConfig.CompressionLevel != config.CompressionLevel {
		t.Error("CompressionLevel not set correctly")
	}
	if retrievedConfig.TagName != config.TagName {
		t.Error("TagName not set correctly")
	}
	if retrievedConfig.MaxDepth != config.MaxDepth {
		t.Error("MaxDepth not set correctly")
	}

	// Test stats
	stats := parser.GetStats()
	expectedKeys := []string{
		"max_depth", "current_depth", "compression_level",
		"include_type_info", "include_field_info", "error_correction",
		"checksum_enabled",
	}

	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Stats missing key: %s", key)
		}
	}

	// Test clone
	cloned := parser.Clone()
	if cloned.config.MaxDepth != parser.config.MaxDepth {
		t.Error("Cloned parser should have same configuration")
	}

	// Verify they are separate instances
	cloned.config.MaxDepth = 200
	if parser.config.MaxDepth == 200 {
		t.Error("Modifying cloned config should not affect original")
	}
}

func TestFieldVersionsAndSorting(t *testing.T) {
	// Test with field versions and sorting enabled
	config := DefaultConfig()
	config.IncludeFieldVersions = true
	config.SortFields = true
	parser := NewWithConfig(config)

	original := Person{
		Name:   "Version Test",
		Age:    25,
		Active: true,
		Address: &Address{
			Street:  "Version St",
			City:    "Sort City",
			ZipCode: 12345,
		},
	}

	// Encode and decode
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Failed to encode with versions: %v", err)
	}

	var decoded Person
	err = parser.Decode(encoded, &decoded)
	if err != nil {
		t.Fatalf("Failed to decode with versions: %v", err)
	}

	// Verify data integrity
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Address.City != original.Address.City {
		t.Errorf("City mismatch: got %s, want %s", decoded.Address.City, original.Address.City)
	}
}

func TestSaveToFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.Create("testfile.etf")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	filename := tempFile.Name()
	tempFile.Close()          // Close it so we can reopen later
	defer os.Remove(filename) // Clean up after test

	// Test data
	original := Person{
		Name:   "File Test User",
		Age:    35,
		Active: true,
		Address: &Address{
			Street:  "456 File St",
			City:    "Storage City",
			ZipCode: 54321,
		},
		Tags:   []string{"file", "storage", "test"},
		Scores: map[string]int{"persistence": 100, "io": 95},
	}

	// Create parser and encode data
	parser := New()
	encoded, err := parser.Encode(original)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	// Write encoded data to file
	if err := os.WriteFile(filename, encoded, 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read encoded data from file
	readData, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Decode data from file
	var decoded Person
	err = parser.Decode(readData, &decoded)
	if err != nil {
		t.Fatalf("Failed to decode data from file: %v", err)
	}

	// Verify decoded data matches original
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}
	if decoded.Age != original.Age {
		t.Errorf("Age mismatch: got %d, want %d", decoded.Age, original.Age)
	}
	if decoded.Address.Street != original.Address.Street {
		t.Errorf("Street mismatch: got %s, want %s", decoded.Address.Street, original.Address.Street)
	}
	if !reflect.DeepEqual(decoded.Tags, original.Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", decoded.Tags, original.Tags)
	}
	if !reflect.DeepEqual(decoded.Scores, original.Scores) {
		t.Errorf("Scores mismatch: got %v, want %v", decoded.Scores, original.Scores)
	}

	t.Logf("Successfully encoded to file, read back and decoded. File size: %d bytes", len(readData))
}
