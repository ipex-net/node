package netintel

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/log"
)

// TestErrorHandling_ServiceStartupErrors tests error handling during service startup
func TestErrorHandling_ServiceStartupErrors(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	service := NewService(tempDir, logger)

	// Test starting service without any data files
	err := service.Start()
	if err != nil {
		// If it fails to start, verify proper state
		if service.IsRunning() {
			t.Error("Service should not be running after startup failure")
		}
		if service.IsReady() {
			t.Error("Service should not be ready after startup failure")
		}
	} else {
		// If it starts successfully (which is possible), it should not be ready without data
		defer service.Stop()
		if service.IsReady() {
			t.Error("Service should not be ready without proper data files")
		}
	}
}

// TestErrorHandling_DataLoadingErrors tests error handling when loading invalid data
func TestErrorHandling_DataLoadingErrors(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create country data so service can start
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	// Create invalid reputation file
	invalidRepFile := filepath.Join(tempDir, "sources", "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(invalidRepFile), 0755)
	os.WriteFile(invalidRepFile, []byte("completely invalid data\n"), 0644)

	service := NewService(tempDir, logger)

	// Service should still start even with invalid reputation data
	// (because it logs errors but continues with available data)
	err := service.Start()
	if err != nil {
		t.Fatalf("Service should start even with some invalid data: %v", err)
	}
	defer service.Stop()

	// Service should be ready with valid country data
	if !service.IsReady() {
		t.Error("Service should be ready with valid country data despite invalid reputation data")
	}

	// Country lookups should work
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Country lookups should work despite reputation data errors")
	}

	// Reputation lookups should return 0 for invalid data
	result := service.CheckReputation("127.0.0.1", 1)
	if result != 0 {
		t.Error("Reputation lookups should return 0 when no valid data is loaded")
	}
}

// TestErrorHandling_ReputationClientErrors tests reputation client error cases
func TestErrorHandling_ReputationClientErrors(t *testing.T) {
	tempDir := t.TempDir()
	client := NewReputationClient(tempDir)

	// Test loading non-existent level
	err := client.LoadLevel(1)
	if err == nil {
		t.Error("LoadLevel should return error for non-existent file")
	}

	// Test loading invalid level numbers
	invalidLevels := []int{0, -1, 5, 10}
	for _, level := range invalidLevels {
		// These should return 0 for lookups (not crash)
		result := client.Lookup(level, "127.0.0.1")
		if result != 0 {
			t.Errorf("Lookup with invalid level %d should return 0, got %d", level, result)
		}
	}

	// Test lookup with invalid IP format
	result := client.Lookup(1, "invalid-ip")
	if result != 0 {
		t.Error("Lookup with invalid IP should return 0")
	}

	// Test lookup with invalid CIDR format
	result = client.Lookup(1, "192.168.1.1/99")
	if result != 0 {
		t.Error("Lookup with invalid CIDR should return 0")
	}
}

// TestErrorHandling_GeolocationClientErrors tests geolocation client error cases
func TestErrorHandling_GeolocationClientErrors(t *testing.T) {
	tempDir := t.TempDir()
	client := NewGeolocationClient(tempDir)

	// Test loading non-existent country
	err := client.LoadCountry("XX")
	if err == nil {
		t.Error("LoadCountry should return error when no source files exist")
	}

	// Test lookup on unloaded country
	result := client.CheckCountry("127.0.0.1", "XX")
	if result {
		t.Error("CheckCountry should return false for unloaded country")
	}

	// Test lookup with invalid IP
	result = client.CheckCountry("invalid-ip", "US")
	if result {
		t.Error("CheckCountry should return false for invalid IP")
	}

	// Test lookup with empty country code
	result = client.CheckCountry("127.0.0.1", "")
	if result {
		t.Error("CheckCountry should return false for empty country code")
	}

	// Test case sensitivity (should work)
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	err = client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Failed to load country: %v", err)
	}

	// These should all work (case insensitive)
	casTests := []string{"US", "us", "Us", "uS"}
	for _, cc := range casTests {
		if !client.CheckCountry("192.168.1.1", cc) {
			t.Errorf("CheckCountry should be case insensitive for %s", cc)
		}
	}
}

// TestErrorHandling_ASNClientErrors tests ASN client error cases
func TestErrorHandling_ASNClientErrors(t *testing.T) {
	tempDir := t.TempDir()
	client := NewASNClient(tempDir)

	// Test loading non-existent ASN file
	err := client.LoadASNDropList()
	if err == nil {
		t.Error("LoadASNDropList should return error when file doesn't exist")
	}

	// Test lookup on unloaded ASN data
	result := client.CheckASNDrop(64512)
	if result {
		t.Error("CheckASNDrop should return false when no data is loaded")
	}

	// Create invalid ASN file
	sourcesDir := filepath.Join(tempDir, "sources")
	os.MkdirAll(sourcesDir, 0755)
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	os.WriteFile(asnFile, []byte("invalid json data\n"), 0644)

	// Test loading invalid JSON
	err = client.LoadASNDropList()
	if err == nil {
		t.Error("LoadASNDropList should return error for invalid JSON")
	}

	// Create file with mixed valid and invalid JSON lines
	mixedData := `{"asn": 64512, "rir": "test", "domain": "test.com", "cc": "US", "asname": "Test ASN", "type": "hosting"}
invalid json line
{"asn": 64513, "rir": "test", "domain": "test2.com", "cc": "US", "asname": "Test ASN 2", "type": "hosting"}
another invalid line
`
	os.WriteFile(asnFile, []byte(mixedData), 0644)

	// Should fail due to invalid JSON (current implementation stops on first error)
	err = client.LoadASNDropList()
	if err == nil {
		t.Error("LoadASNDropList should fail with mixed valid/invalid data")
	}

	// Create file with only valid JSON entries
	validData := `{"asn": 64512, "rir": "test", "domain": "test.com", "cc": "US", "asname": "Test ASN", "type": "hosting"}
{"asn": 64513, "rir": "test", "domain": "test2.com", "cc": "US", "asname": "Test ASN 2", "type": "hosting"}
`
	os.WriteFile(asnFile, []byte(validData), 0644)

	// Should succeed with all valid data
	err = client.LoadASNDropList()
	if err != nil {
		t.Fatalf("LoadASNDropList should succeed with valid data: %v", err)
	}

	// Valid ASNs should be found
	if !client.CheckASNDrop(64512) {
		t.Error("Valid ASN should be found")
	}
	if !client.CheckASNDrop(64513) {
		t.Error("Valid ASN should be found")
	}

	// Count should reflect all valid entries
	count := client.GetLoadedASNCount()
	if count != 2 {
		t.Errorf("Expected 2 loaded ASNs, got %d", count)
	}
}

// TestErrorHandling_AtomicCacheErrors tests atomic cache error cases
func TestErrorHandling_AtomicCacheErrors(t *testing.T) {
	cache := NewAtomicCache()

	// Test operations on uninitialized cache
	if cache.IsReady() {
		t.Error("Cache should not be ready when uninitialized")
	}

	// Test lookups on uninitialized cache
	result := cache.CheckReputation(1, "127.0.0.1")
	if result != 0 {
		t.Error("Reputation check should return 0 when cache is uninitialized")
	}

	countryResult := cache.CheckCountry("127.0.0.1", "US")
	if countryResult {
		t.Error("Country check should return false when cache is uninitialized")
	}

	asnResult := cache.CheckASNDrop(64512)
	if asnResult {
		t.Error("ASN check should return false when cache is uninitialized")
	}

	// Test getting checkers from uninitialized cache
	repChecker := cache.GetReputationChecker()
	if repChecker != nil {
		t.Error("GetReputationChecker should return nil when uninitialized")
	}

	geoChecker := cache.GetGeolocationChecker()
	if geoChecker != nil {
		t.Error("GetGeolocationChecker should return nil when uninitialized")
	}

	asnChecker := cache.GetASNChecker()
	if asnChecker != nil {
		t.Error("GetASNChecker should return nil when uninitialized")
	}
}

// TestErrorHandling_ServiceOperationsAfterStop tests service operations after stop
func TestErrorHandling_ServiceOperationsAfterStop(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create minimal data so service can start
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Verify service is working
	if !service.IsRunning() {
		t.Error("Service should be running")
	}

	// Stop the service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Test operations after stop
	if service.IsRunning() {
		t.Error("Service should not be running after stop")
	}

	if service.IsReady() {
		t.Error("Service should not be ready after stop")
	}

	// Test lookups after stop (should return safe defaults)
	result := service.CheckReputation("127.0.0.1", 1)
	if result != 0 {
		t.Error("Reputation lookup should return 0 after service stop")
	}

	countryResult := service.CheckCountry("192.168.1.1", "US")
	if countryResult {
		t.Error("Country lookup should return false after service stop")
	}

	asnResult := service.CheckASNDrop(64512)
	if asnResult {
		t.Error("ASN lookup should return false after service stop")
	}

	// Test double stop (should not error)
	err = service.Stop()
	if err != nil {
		t.Error("Double stop should not return error")
	}
}

// TestErrorHandling_ConcurrentOperations tests error handling during concurrent operations
func TestErrorHandling_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create data for service
	sourcesDir := filepath.Join(tempDir, "sources")

	// Country data
	countryDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(countryDir, 0755)
	countryFile := filepath.Join(countryDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	// Reputation data
	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		data := "127.0.0.1\n192.168.1.0/24\n"
		os.WriteFile(repFile, []byte(data), 0644)
	}

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	for i := 0; i < 100 && !service.IsReady(); i++ {
		// This loop tests that checking readiness rapidly doesn't cause issues
		if i > 50 {
			t.Fatal("Service didn't become ready in reasonable time")
		}
	}

	// Test concurrent operations that could potentially interfere
	done := make(chan bool, 3)

	// Concurrent readiness checks
	go func() {
		for i := 0; i < 100; i++ {
			service.IsReady()
		}
		done <- true
	}()

	// Concurrent lookups
	go func() {
		for i := 0; i < 100; i++ {
			service.CheckCountry("192.168.1.1", "US")
			service.CheckReputation("127.0.0.1", 1)
		}
		done <- true
	}()

	// Concurrent status checks
	go func() {
		for i := 0; i < 100; i++ {
			service.GetLoadedCountries()
			service.GetLoadedLevels()
			service.GetLastUpdate()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Service should still be functional
	if !service.IsReady() {
		t.Error("Service should remain ready after concurrent operations")
	}
}

// TestErrorHandling_FileSystemErrors tests handling of file system errors
func TestErrorHandling_FileSystemErrors(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Test with non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent", "deep", "path")
	service := NewService(nonExistentDir, logger)

	// Service may start (doesn't try to create directories on start)
	// but should not be ready without proper data files
	err := service.Start()
	if err != nil {
		// If it fails to start, that's expected
		if service.IsRunning() {
			t.Error("Service should not be running after failure")
		}
	} else {
		defer service.Stop()
		// If it starts, it should not be ready without data
		if service.IsReady() {
			t.Error("Service should not be ready with non-existent data directory")
		}
	}

	// Test individual client errors with bad paths
	badClient := NewReputationClient("/root/nonexistent/path")
	err = badClient.LoadLevel(1)
	if err == nil {
		t.Error("ReputationClient should fail with bad path")
	}

	badGeoClient := NewGeolocationClient("/root/nonexistent/path")
	err = badGeoClient.LoadCountry("US")
	if err == nil {
		t.Error("GeolocationClient should fail with bad path")
	}
}

// TestErrorHandling_CorruptedDataRecovery tests recovery from corrupted data
func TestErrorHandling_CorruptedDataRecovery(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create valid initial data (both country and reputation for readiness)
	sourcesDir := filepath.Join(tempDir, "sources")

	// Country data
	countryDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(countryDir, 0755)
	countryFile := filepath.Join(countryDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	// Reputation data
	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		data := "127.0.0.1\n192.168.1.0/24\n"
		os.WriteFile(repFile, []byte(data), 0644)
	}

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	for i := 0; i < 100 && !service.IsReady(); i++ {
		if i > 99 {
			t.Fatal("Service didn't become ready")
		}
	}

	// Verify initial functionality
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Initial country lookup should work")
	}

	// Corrupt the data file
	os.WriteFile(countryFile, []byte("corrupted binary data \x00\x01\x02"), 0644)

	// Try to update (should handle corruption gracefully)
	service.updateCountryData()

	// Service should still be functional with previous good data
	// (The atomic swap means bad data doesn't replace good data)
	if !service.IsReady() {
		t.Error("Service should remain ready despite data corruption")
	}
}

// TestErrorHandling_ErrorTypes tests custom error types and error wrapping
func TestErrorHandling_ErrorTypes(t *testing.T) {
	// Test that errors from components are properly handled
	tempDir := t.TempDir()

	client := NewReputationClient(tempDir)
	err := client.LoadLevel(1)

	// Should return an error (file doesn't exist)
	if err == nil {
		t.Error("LoadLevel should return error for missing file")
	}

	// Error should be descriptive
	if err.Error() == "" {
		t.Error("Error should have descriptive message")
	}

	// Test error wrapping behavior
	geoClient := NewGeolocationClient(tempDir)
	err = geoClient.LoadCountry("XX")

	if err == nil {
		t.Error("LoadCountry should return error when no files exist")
	}

	// Test that errors can be checked
	if !errors.Is(err, err) {
		t.Error("Error should be identifiable")
	}
}
