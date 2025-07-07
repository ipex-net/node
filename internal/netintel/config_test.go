package netintel

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
)

// TestConfiguration_DefaultConstants tests the default configuration constants
func TestConfiguration_DefaultConstants(t *testing.T) {
	// Test that configuration constants are reasonable
	expectedValues := map[string]interface{}{
		"reputationUpdateInterval": reputationUpdateInterval,
		"countryUpdateInterval":    countryUpdateInterval,
		"asnUpdateInterval":        asnUpdateInterval,
		"initialLoadTimeout":       initialLoadTimeout,
	}

	// Reputation update interval should be reasonable (1 minute)
	if reputationUpdateInterval != 1*time.Minute {
		t.Errorf("Reputation update interval should be 1 minute, got %v", reputationUpdateInterval)
	}

	// Country update interval should be reasonable (5 minutes in dev)
	if countryUpdateInterval != 5*time.Minute {
		t.Errorf("Country update interval should be 5 minutes, got %v", countryUpdateInterval)
	}

	// ASN update interval should be reasonable (1 hour)
	if asnUpdateInterval != 1*time.Hour {
		t.Errorf("ASN update interval should be 1 hour, got %v", asnUpdateInterval)
	}

	// Initial load timeout should be reasonable (30 seconds)
	if initialLoadTimeout != 30*time.Second {
		t.Errorf("Initial load timeout should be 30 seconds, got %v", initialLoadTimeout)
	}

	t.Logf("Configuration Constants:")
	for name, value := range expectedValues {
		t.Logf("  %s: %v", name, value)
	}
}

// TestConfiguration_ServiceWithDifferentDataDirectories tests service behavior with various data directory configurations
func TestConfiguration_ServiceWithDifferentDataDirectories(t *testing.T) {
	logger := log.NewNopLogger()

	testCases := []struct {
		name        string
		setupDir    func(string) string
		expectError bool
		description string
	}{
		{
			name: "valid_directory",
			setupDir: func(baseDir string) string {
				dataDir := filepath.Join(baseDir, "valid")
				os.MkdirAll(dataDir, 0755)

				// Create minimal data
				sourcesDir := filepath.Join(dataDir, "sources", "geolite2_country")
				os.MkdirAll(sourcesDir, 0755)
				countryFile := filepath.Join(sourcesDir, "country_us.netset")
				os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

				return dataDir
			},
			expectError: false,
			description: "Service should work with valid data directory",
		},
		{
			name: "empty_directory",
			setupDir: func(baseDir string) string {
				dataDir := filepath.Join(baseDir, "empty")
				os.MkdirAll(dataDir, 0755)
				return dataDir
			},
			expectError: false,
			description: "Service should handle empty directory gracefully",
		},
		{
			name: "nonexistent_directory",
			setupDir: func(baseDir string) string {
				return filepath.Join(baseDir, "nonexistent")
			},
			expectError: false,
			description: "Service should handle nonexistent directory gracefully",
		},
		{
			name: "nested_directory",
			setupDir: func(baseDir string) string {
				dataDir := filepath.Join(baseDir, "deep", "nested", "structure")
				os.MkdirAll(dataDir, 0755)

				// Create minimal data
				sourcesDir := filepath.Join(dataDir, "sources", "geolite2_country")
				os.MkdirAll(sourcesDir, 0755)
				countryFile := filepath.Join(sourcesDir, "country_us.netset")
				os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

				return dataDir
			},
			expectError: false,
			description: "Service should work with nested directory structure",
		},
		{
			name: "relative_path",
			setupDir: func(baseDir string) string {
				// Create directory with relative path
				absDir := filepath.Join(baseDir, "relative")
				os.MkdirAll(absDir, 0755)

				// Create minimal data
				sourcesDir := filepath.Join(absDir, "sources", "geolite2_country")
				os.MkdirAll(sourcesDir, 0755)
				countryFile := filepath.Join(sourcesDir, "country_us.netset")
				os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

				// Change to temp dir and return relative path
				originalDir, _ := os.Getwd()
				os.Chdir(baseDir)
				defer os.Chdir(originalDir)

				return "./relative"
			},
			expectError: false,
			description: "Service should work with relative paths",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()
			dataDir := tc.setupDir(baseDir)

			service := NewServiceForTesting(dataDir, logger)

			err := service.Start()

			if tc.expectError {
				if err == nil {
					service.Stop()
					t.Errorf("%s: expected error but got none", tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tc.description, err)
				} else {
					defer service.Stop()

					// Service should be functional for valid configurations
					if service.IsRunning() {
						t.Logf("%s: service started successfully", tc.description)
					}
				}
			}
		})
	}
}

// TestConfiguration_ClientInitialization tests client initialization with different configurations
func TestConfiguration_ClientInitialization(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	testCases := []struct {
		clientType string
		testFunc   func(string) error
	}{
		{
			"ReputationClient",
			func(dataDir string) error {
				client := NewReputationClient(dataDir)
				// Test that client is properly initialized
				if client == nil {
					return nil
				}
				// Test client methods don't panic
				client.GetLoadedLevels()
				client.IsLevelLoaded(1)
				client.Clear()
				return nil
			},
		},
		{
			"GeolocationClient",
			func(dataDir string) error {
				client := NewGeolocationClient(dataDir)
				if client == nil {
					return nil
				}
				// Test client methods don't panic
				client.GetLoadedCountries()
				client.IsCountryLoaded("US")
				client.Clear()
				return nil
			},
		},
		{
			"ASNClient",
			func(dataDir string) error {
				client := NewASNClient(dataDir)
				if client == nil {
					return nil
				}
				// Test client methods don't panic
				client.GetLoadedASNCount()
				client.CheckASNDrop(64512)
				return nil
			},
		},
		{
			"HTTPDownloader",
			func(dataDir string) error {
				downloader := NewHTTPDownloader(dataDir, logger)
				if downloader == nil {
					return nil
				}
				// Test downloader methods don't panic
				downloader.GetLastDownload()
				return nil
			},
		},
		{
			"AtomicCache",
			func(dataDir string) error {
				cache := NewAtomicCache()
				if cache == nil {
					return nil
				}
				// Test cache methods don't panic
				cache.IsReady()
				cache.IsReputationReady()
				cache.IsGeolocationReady()
				cache.GetLastUpdate()
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.clientType, func(t *testing.T) {
			err := tc.testFunc(tempDir)
			if err != nil {
				t.Errorf("%s initialization failed: %v", tc.clientType, err)
			}
		})
	}
}

// TestConfiguration_ServiceParameters tests service behavior with different parameter combinations
func TestConfiguration_ServiceParameters(t *testing.T) {
	baseDir := t.TempDir()
	logger := log.NewNopLogger()

	testCases := []struct {
		name     string
		setup    func() (string, log.Logger)
		validate func(*Service, error) error
	}{
		{
			name: "nil_logger",
			setup: func() (string, log.Logger) {
				dataDir := filepath.Join(baseDir, "nil_logger")
				os.MkdirAll(dataDir, 0755)
				return dataDir, nil
			},
			validate: func(service *Service, err error) error {
				// Service should handle nil logger gracefully
				if service == nil {
					return fmt.Errorf("service should not be nil with nil logger")
				}
				return nil
			},
		},
		{
			name: "empty_data_dir",
			setup: func() (string, log.Logger) {
				return "", logger
			},
			validate: func(service *Service, err error) error {
				// Service should handle empty data dir
				if service == nil {
					return fmt.Errorf("service should not be nil with empty data dir")
				}
				return nil
			},
		},
		{
			name: "valid_parameters",
			setup: func() (string, log.Logger) {
				dataDir := filepath.Join(baseDir, "valid")
				os.MkdirAll(dataDir, 0755)
				return dataDir, logger
			},
			validate: func(service *Service, err error) error {
				if service == nil {
					return fmt.Errorf("service should not be nil with valid parameters")
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, logger := tc.setup()

			service := NewServiceForTesting(dataDir, logger)

			if err := tc.validate(service, nil); err != nil {
				t.Error(err)
			}
		})
	}
}

// TestConfiguration_DataPathStructure tests expected data directory structure
func TestConfiguration_DataPathStructure(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Test expected directory structure
	expectedPaths := map[string]string{
		"reputation_level1": "sources/firehol_level1.netset",
		"reputation_level2": "sources/firehol_level2.netset",
		"reputation_level3": "sources/firehol_level3.netset",
		"reputation_level4": "sources/firehol_level4.netset",
		"geolite2_us":       "sources/geolite2_country/country_us.netset",
		"ip2location_us":    "sources/ip2location_country/ip2location_country_us.netset",
		"ipdeny_us":         "sources/ipdeny_country/id_country_us.netset",
		"ipip_us":           "sources/ipip_country/ipip_country_us.netset",
		"asn_drop":          "sources/asndrop.json",
	}

	// Create all expected paths
	for name, path := range expectedPaths {
		fullPath := filepath.Join(tempDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)

		// Create appropriate test data
		var data string
		if filepath.Ext(path) == ".json" {
			data = `{"asn": 64512, "rir": "IANA", "domain": "test.com", "cc": "US", "asname": "TEST-ASN", "type": "hosting"}`
		} else {
			data = "192.168.1.0/24\n127.0.0.1\n"
		}

		os.WriteFile(fullPath, []byte(data), 0644)
		t.Logf("Created %s: %s", name, path)
	}

	// Test service with complete data structure
	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Service should start with complete data structure: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(10 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service should become ready with complete data")
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Verify all data types are loaded
	levels := service.GetLoadedLevels()
	if len(levels) != 4 {
		t.Errorf("Expected 4 reputation levels, got %d", len(levels))
	}

	countries := service.GetLoadedCountries()
	if len(countries) == 0 {
		t.Error("Expected countries to be loaded")
	}

	// Test functionality
	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Reputation lookup should work with complete data")
	}

	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Country lookup should work with complete data")
	}

	t.Logf("Service fully functional with complete data structure")
}

// TestConfiguration_UpdateIntervals tests that update intervals are properly configured
func TestConfiguration_UpdateIntervals(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create minimal data
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Test that the service starts background loops with configured intervals
	// We can't easily test the actual timing without mocking, but we can verify
	// that the service doesn't crash and the intervals are reasonable

	// Verify intervals are in reasonable ranges
	if reputationUpdateInterval < 30*time.Second {
		t.Error("Reputation update interval seems too short for production")
	}
	if reputationUpdateInterval > 10*time.Minute {
		t.Error("Reputation update interval seems too long")
	}

	if countryUpdateInterval < 1*time.Minute {
		t.Error("Country update interval seems too short")
	}
	if countryUpdateInterval > 48*time.Hour {
		t.Error("Country update interval seems too long")
	}

	if asnUpdateInterval < 10*time.Minute {
		t.Error("ASN update interval seems too short")
	}
	if asnUpdateInterval > 24*time.Hour {
		t.Error("ASN update interval seems too long")
	}

	if initialLoadTimeout < 5*time.Second {
		t.Error("Initial load timeout seems too short")
	}
	if initialLoadTimeout > 2*time.Minute {
		t.Error("Initial load timeout seems too long")
	}

	t.Logf("All update intervals are within reasonable ranges")
}

// TestConfiguration_DefaultBehavior tests default service behavior without any configuration
func TestConfiguration_DefaultBehavior(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Test service with minimal setup (just empty directories)
	service := NewService(tempDir, logger)

	// Service should be created without error
	if service == nil {
		t.Fatal("Service should be created with default configuration")
	}

	// Test that service methods don't panic even without data
	if service.IsRunning() {
		t.Error("Service should not be running before start")
	}

	if service.IsReady() {
		t.Error("Service should not be ready before start")
	}

	// Test safe defaults for lookups
	repResult := service.CheckReputation("127.0.0.1/32", 1)
	if repResult != 0 {
		t.Error("Reputation check should return 0 when not ready")
	}

	countryResult := service.CheckCountry("127.0.0.1", "US")
	if countryResult {
		t.Error("Country check should return false when not ready")
	}

	asnResult := service.CheckASNDrop(64512)
	if asnResult {
		t.Error("ASN check should return false when not ready")
	}

	// Test metadata methods
	levels := service.GetLoadedLevels()
	countries := service.GetLoadedCountries()
	lastUpdate := service.GetLastUpdate()

	t.Logf("Default behavior results:")
	t.Logf("  Loaded levels: %v", levels)
	t.Logf("  Loaded countries: %v", countries)
	t.Logf("  Last update: %v", lastUpdate)

	// These should be safe to call
	if levels == nil {
		t.Error("GetLoadedLevels should not return nil")
	}
	if countries == nil {
		t.Error("GetLoadedCountries should not return nil")
	}
}

// TestConfiguration_EdgeCases tests edge cases in configuration
func TestConfiguration_EdgeCases(t *testing.T) {
	logger := log.NewNopLogger()

	testCases := []struct {
		name        string
		dataDir     string
		description string
	}{
		{
			name:        "very_long_path",
			dataDir:     filepath.Join(os.TempDir(), "very", "very", "very", "long", "path", "structure", "that", "goes", "deep", "into", "filesystem", "hierarchy"),
			description: "Service should handle very long paths",
		},
		{
			name:        "special_characters",
			dataDir:     filepath.Join(os.TempDir(), "special-chars_123", "test@data", "config#test"),
			description: "Service should handle special characters in paths",
		},
		{
			name:        "unicode_path",
			dataDir:     filepath.Join(os.TempDir(), "测试", "тест", "テスト"),
			description: "Service should handle unicode characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up first
			os.RemoveAll(tc.dataDir)

			// Create directory structure
			os.MkdirAll(tc.dataDir, 0755)
			defer os.RemoveAll(tc.dataDir)

			service := NewService(tc.dataDir, logger)

			// Service should be created without panic
			if service == nil {
				t.Errorf("%s: service should be created", tc.description)
			}

			// Starting might fail, but shouldn't panic
			err := service.Start()
			if err == nil {
				defer service.Stop()
				t.Logf("%s: service started successfully", tc.description)
			} else {
				t.Logf("%s: service failed to start as expected: %v", tc.description, err)
			}
		})
	}
}

// TestConfiguration_ResourceLimits tests behavior under resource constraints
func TestConfiguration_ResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource limits test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Test with large number of small files (simulate many countries)
	sourcesDir := filepath.Join(tempDir, "sources")

	// Create many small geolocation files
	for i := 0; i < 100; i++ {
		providerDir := filepath.Join(sourcesDir, "geolite2_country")
		os.MkdirAll(providerDir, 0755)

		countryCode := fmt.Sprintf("t%02d", i) // Test country codes
		countryFile := filepath.Join(providerDir, fmt.Sprintf("country_%s.netset", countryCode))
		data := fmt.Sprintf("192.168.%d.0/24\n", i)
		os.WriteFile(countryFile, []byte(data), 0644)
	}

	service := NewService(tempDir, logger)

	// Service should handle many files gracefully
	err := service.Start()
	if err != nil {
		t.Fatalf("Service should start with many files: %v", err)
	}
	defer service.Stop()

	// Wait for readiness (might take longer with many files)
	timeout := time.After(30 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service should become ready even with many files")
		case <-time.After(500 * time.Millisecond):
		}
	}

	// Test that service is functional
	countries := service.GetLoadedCountries()
	t.Logf("Loaded %d countries with many files", len(countries))

	if len(countries) == 0 {
		t.Error("Should have loaded some countries")
	}

	// Service should remain stable
	if !service.IsReady() {
		t.Error("Service should remain ready with many files")
	}
}

func TestConfiguration_ReadOnlyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create a directory structure with data first
	dataDir := filepath.Join(tempDir, "readonly")
	sourcesDir := filepath.Join(dataDir, "sources")
	os.MkdirAll(sourcesDir, 0755)

	// Create reputation data (required for service to be ready)
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.WriteFile(repFile, []byte("127.0.0.1\n192.168.100.0/24\n"), 0644)

	// Create geolocation data (also required for service to be ready)
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte("192.168.1.0/24\n"), 0644)

	// Make the directories read-only AFTER creating all the structure and files
	// This simulates a scenario where data exists but background updates can't write new files
	defer func() {
		// Restore permissions for cleanup
		os.Chmod(dataDir, 0755)
		os.Chmod(sourcesDir, 0755)
		os.Chmod(geoDir, 0755)
	}()

	os.Chmod(geoDir, 0555)     // Make geolocation dir read-only
	os.Chmod(sourcesDir, 0555) // Make sources dir read-only
	os.Chmod(dataDir, 0555)    // Make data dir read-only

	service := NewService(dataDir, logger)
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready (should succeed since data already exists)
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// The service should be able to read the existing data and become ready
	if !service.IsReady() {
		t.Error("Service should be ready with a read-only directory containing existing data")
	}

	// Verify the service can perform lookups with existing data
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Service should be able to perform country lookups with existing data")
	}

	// Verify reputation lookups also work
	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Service should be able to perform reputation lookups with existing data")
	}

	// Background updates will fail silently when trying to write to read-only directories,
	// but the service should not crash and should remain ready
	// We can't easily test the actual background update failure without waiting for the timer,
	// but we can verify the service remains stable
	time.Sleep(100 * time.Millisecond)

	// The service should still be ready after attempting to operate on read-only directories
	if !service.IsReady() {
		t.Error("Service should remain ready even with read-only directories")
	}

	// Verify service remains functional with both types of lookups
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Service should remain functional with read-only directories")
	}

	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Service reputation lookups should remain functional with read-only directories")
	}
}
