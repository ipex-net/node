package netintel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestASNClient_LoadASNDropList(t *testing.T) {
	// Create temporary directory for test data
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test ASN DROP list as newline-delimited JSON
	testEntries := []SpamhausASNEntry{
		{ASN: 12345, RIR: "arin", Domain: "test-domain-1", CC: "US", ASName: "TEST-ASN-1"},
		{ASN: 67890, RIR: "ripe", Domain: "test-domain-2", CC: "GB", ASName: "TEST-ASN-2"},
		{ASN: 11111, RIR: "apnic", Domain: "test-domain-3", CC: "JP", ASName: "TEST-ASN-3"},
	}

	var jsonLines []string
	for _, entry := range testEntries {
		line, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal test entry: %v", err)
		}
		jsonLines = append(jsonLines, string(line))
	}
	jsonData := []byte(strings.Join(jsonLines, "\n"))

	asnDropFile := filepath.Join(sourcesDir, "asndrop.json")
	if err := os.WriteFile(asnDropFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test ASN DROP file: %v", err)
	}

	// Create ASN client and load data
	client := NewASNClient(tempDir)
	if err := client.LoadASNDropList(); err != nil {
		t.Fatalf("Failed to load ASN DROP list: %v", err)
	}

	// Test that the correct number of ASNs were loaded
	expectedCount := 3
	if count := client.GetLoadedASNCount(); count != expectedCount {
		t.Errorf("Expected %d ASNs, got %d", expectedCount, count)
	}

	// Test checking ASNs that should be in the DROP list
	testCases := []struct {
		asn      int32
		expected bool
	}{
		{12345, true},  // Should be in DROP list
		{67890, true},  // Should be in DROP list
		{11111, true},  // Should be in DROP list
		{99999, false}, // Should NOT be in DROP list
		{0, false},     // Should NOT be in DROP list
	}

	for _, tc := range testCases {
		result := client.CheckASNDrop(tc.asn)
		if result != tc.expected {
			t.Errorf("CheckASNDrop(%d) = %v, expected %v", tc.asn, result, tc.expected)
		}
	}
}

func TestASNClient_LoadASNDropList_EmptyFile(t *testing.T) {
	// Create temporary directory for test data
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create empty ASN DROP list JSON file
	jsonData := []byte("") // Empty file

	asnDropFile := filepath.Join(sourcesDir, "asndrop.json")
	if err := os.WriteFile(asnDropFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write empty ASN DROP file: %v", err)
	}

	// Create ASN client and load data
	client := NewASNClient(tempDir)
	if err := client.LoadASNDropList(); err != nil {
		t.Fatalf("Failed to load empty ASN DROP list: %v", err)
	}

	// Test that no ASNs were loaded
	if count := client.GetLoadedASNCount(); count != 0 {
		t.Errorf("Expected 0 ASNs, got %d", count)
	}

	// Test that any ASN check returns false
	if result := client.CheckASNDrop(12345); result {
		t.Errorf("CheckASNDrop(12345) = true, expected false for empty list")
	}
}

func TestASNClient_LoadASNDropList_FileNotFound(t *testing.T) {
	// Create temporary directory WITHOUT the ASN DROP file
	tempDir := t.TempDir()

	// Create ASN client and try to load non-existent data
	client := NewASNClient(tempDir)
	err := client.LoadASNDropList()

	// Should return error for missing file
	if err == nil {
		t.Error("Expected error for missing ASN DROP file, got nil")
	}
}

func TestASNClient_LoadASNDropList_WithMetadata(t *testing.T) {
	// Create temporary directory for test data
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data with ASN entries and metadata entry as newline-delimited JSON
	testEntries := []SpamhausASNEntry{
		{ASN: 12345, RIR: "arin", Domain: "test-domain-1", CC: "US", ASName: "TEST-ASN-1"},
		{ASN: 67890, RIR: "ripe", Domain: "test-domain-2", CC: "GB", ASName: "TEST-ASN-2"},
		{Type: "metadata"}, // This should be filtered out
	}

	var jsonLines []string
	for _, entry := range testEntries {
		line, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("Failed to marshal test entry: %v", err)
		}
		jsonLines = append(jsonLines, string(line))
	}
	jsonData := []byte(strings.Join(jsonLines, "\n"))

	asnDropFile := filepath.Join(sourcesDir, "asndrop.json")
	if err := os.WriteFile(asnDropFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test ASN DROP file: %v", err)
	}

	// Create ASN client and load data
	client := NewASNClient(tempDir)
	if err := client.LoadASNDropList(); err != nil {
		t.Fatalf("Failed to load ASN DROP list: %v", err)
	}

	// Test that only ASN entries were loaded (metadata entry should be filtered out)
	expectedCount := 2 // Only the two ASN entries, not the metadata entry
	if count := client.GetLoadedASNCount(); count != expectedCount {
		t.Errorf("Expected %d ASNs, got %d", expectedCount, count)
	}

	// Test that ASN entries are found and metadata is not
	if !client.CheckASNDrop(12345) {
		t.Error("Expected ASN 12345 to be in DROP list")
	}
	if !client.CheckASNDrop(67890) {
		t.Error("Expected ASN 67890 to be in DROP list")
	}
}

func TestASNClient_LoadASNDropList_MalformedJSON(t *testing.T) {
	// Create temporary directory for test data
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data with malformed JSON
	jsonData := []byte(`{"asn": 12345, "rir": "arin", "domain": "test-domain-1", "cc": "US", "asname": "TEST-ASN-1"}
{"asn": 67890, "rir": "ripe", "domain": "test-domain-2", "cc": "GB", "asname": "TEST-ASN-2"}
malformed json
`)

	asnDropFile := filepath.Join(sourcesDir, "asndrop.json")
	if err := os.WriteFile(asnDropFile, jsonData, 0644); err != nil {
		t.Fatalf("Failed to write test ASN DROP file: %v", err)
	}

	// Create ASN client and load data
	client := NewASNClient(tempDir)
	err := client.LoadASNDropList()
	if err == nil {
		t.Fatal("Expected error when loading malformed JSON")
	}
}
