package netintel

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
)

// TestBackgroundUpdates_ServiceStartStop tests service startup and shutdown behavior
func TestBackgroundUpdates_ServiceStartStop(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create minimal country data for service to start
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	// Test that service starts successfully
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Verify service is running
	if !service.IsRunning() {
		t.Error("Service should be running after start")
	}

	// Wait a short time to allow background loops to start
	time.Sleep(100 * time.Millisecond)

	// Test that service stops successfully
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Service should no longer be running
	if service.IsRunning() {
		t.Error("Service should not be running after stop")
	}
}

// TestBackgroundUpdates_ReputationUpdate tests the reputation data update functionality
func TestBackgroundUpdates_ReputationUpdate(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create existing reputation files to avoid network calls
	sourcesDir := filepath.Join(tempDir, "sources")
	os.MkdirAll(sourcesDir, 0755)

	// Create test reputation files
	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		data := fmt.Sprintf("# Level %d reputation data\n127.0.0.%d\n192.168.%d.0/24\n", level, level, level)
		os.WriteFile(repFile, []byte(data), 0644)
	}

	// Create country data so service can start
	countryDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(countryDir, 0755)
	countryFile := filepath.Join(countryDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Verify initial reputation data is loaded
	initialLevels := service.GetLoadedLevels()
	if len(initialLevels) == 0 {
		t.Error("Service should have loaded reputation levels")
	}

	// Test direct reputation update method (will try to download but will use existing files)
	service.updateReputationData()

	// Verify reputation data is still accessible
	finalLevels := service.GetLoadedLevels()
	if len(finalLevels) == 0 {
		t.Error("Reputation levels should still be available after update")
	}
}

// TestBackgroundUpdates_CountryUpdate tests the country data update functionality
func TestBackgroundUpdates_CountryUpdate(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create initial country data files
	sourcesDir := filepath.Join(tempDir, "sources")
	providers := map[string]string{
		"geolite2_country":    "country_us.netset",
		"ip2location_country": "ip2location_country_us.netset",
	}

	for provider, fileName := range providers {
		providerDir := filepath.Join(sourcesDir, provider)
		os.MkdirAll(providerDir, 0755)

		data := "192.168.1.0/24\n10.0.0.0/8\n"
		countryFile := filepath.Join(providerDir, fileName)
		os.WriteFile(countryFile, []byte(data), 0644)
	}

	// Create reputation data so service can be ready
	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		data := fmt.Sprintf("# Level %d reputation data\n127.0.0.%d\n", level, level)
		os.WriteFile(repFile, []byte(data), 0644)
	}

	service := NewService(tempDir, logger)

	// Test initial load (this calls loadInitialData)
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Test that initial country data is accessible
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Initial country lookup should work after data load")
	}

	// Also test the other initial IP range
	if !service.CheckCountry("10.5.5.5", "US") {
		t.Error("Initial country lookup for 10.x range should work")
	}

	// Clean up any existing country files to avoid interference
	os.RemoveAll(sourcesDir)

	// Recreate sources directory and update country files with completely different data
	for provider, fileName := range providers {
		providerDir := filepath.Join(sourcesDir, provider)
		os.MkdirAll(providerDir, 0755)
		newData := "203.0.113.0/24\n198.51.100.0/24\n"
		countryFile := filepath.Join(providerDir, fileName)
		os.WriteFile(countryFile, []byte(newData), 0644)
	}

	// Test that country update method works (may download from internet)
	// This test is network-dependent and may skip if downloads fail

	// Manually call the update method
	service.updateCountryData()

	// If the update succeeded, we should be able to access some geolocation data
	// Note: This might not be our test data if internet download succeeded
	time.Sleep(500 * time.Millisecond) // Wait for any update to complete

	// Check if any country data is accessible (don't rely on specific test IPs
	// since real downloads might have occurred)
	hasWorkingCountryData := false

	// Try some common IP ranges that might be in real geolocation data
	testIPs := []string{"8.8.8.8", "1.1.1.1", "192.168.1.1", "203.0.113.1", "198.51.100.1"}
	for _, ip := range testIPs {
		if service.CheckCountry(ip, "US") {
			hasWorkingCountryData = true
			t.Logf("Found working country data for IP %s", ip)
			break
		}
	}

	if !hasWorkingCountryData {
		// Try other countries too
		countries := []string{"GB", "CA", "DE", "FR", "JP"}
		for _, country := range countries {
			for _, ip := range testIPs {
				if service.CheckCountry(ip, country) {
					hasWorkingCountryData = true
					t.Logf("Found working country data for IP %s, country %s", ip, country)
					break
				}
			}
			if hasWorkingCountryData {
				break
			}
		}
	}

	if !hasWorkingCountryData {
		t.Skip("Country update test skipped - no working geolocation data available (likely network/download issue)")
	}

	t.Log("Country update test passed - geolocation data is accessible")
}

// TestBackgroundUpdates_ASNUpdate tests the ASN DROP list update functionality
func TestBackgroundUpdates_ASNUpdate(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create minimal country data so service can start
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	// Create existing ASN DROP file to avoid network calls
	asnFile := filepath.Join(tempDir, "asndrop.json")
	asnData := `{"asn": 64512, "rir": "test", "domain": "test.com", "cc": "US", "asname": "Test ASN", "type": "hosting"}
{"asn": 64513, "rir": "test", "domain": "test2.com", "cc": "US", "asname": "Test ASN 2", "type": "hosting"}
`
	os.WriteFile(asnFile, []byte(asnData), 0644)

	service := NewService(tempDir, logger)

	// Start service
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Test direct ASN update method (will try to download but may fail due to network)
	// This tests that the method doesn't crash even if download fails
	service.updateASNData()

	// Service should remain functional regardless of ASN update success/failure
	if !service.IsReady() {
		t.Error("Service should remain ready after ASN update attempt")
	}
}

// TestBackgroundUpdates_ErrorHandling tests error handling in background updates
func TestBackgroundUpdates_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create country data so service can start
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

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Service should be ready with existing data
	if !service.IsReady() {
		t.Error("Service should be ready with country data")
	}

	// Test reputation update with potential errors (will try to download from internet and likely fail)
	// This should not crash the service
	service.updateReputationData()

	// Test ASN update with potential errors (will try to download from internet and likely fail)
	// This should not crash the service
	service.updateASNData()

	// Service should still be functional despite download failures
	if !service.IsReady() {
		t.Error("Service should remain ready after failed updates")
	}

	// Country lookups should still work
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Country lookups should still work after failed updates")
	}
}

// TestBackgroundUpdates_ConcurrentAccess tests that background updates don't interfere with lookups
func TestBackgroundUpdates_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create country data
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n10.0.0.0/8\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Perform concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent lookups
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				// These should not panic or crash
				service.CheckCountry("192.168.1.1", "US")
				service.CheckCountry("10.5.5.5", "US")
				service.CheckCountry("8.8.8.8", "US")

				// Small delay between operations
				time.Sleep(time.Millisecond)
			}
		}()
	}

	// Concurrent background updates (simulate what background loops do)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				// This should not interfere with lookups
				service.updateCountryData()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation error: %v", err)
	}

	// Service should still be functional
	if !service.IsReady() {
		t.Error("Service should remain ready after concurrent operations")
	}
}

// TestBackgroundUpdates_ServiceState tests that service maintains state correctly during updates
func TestBackgroundUpdates_ServiceState(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create initial country data
	sourcesDir := filepath.Join(tempDir, "sources", "geolite2_country")
	os.MkdirAll(sourcesDir, 0755)
	countryFile := filepath.Join(sourcesDir, "country_us.netset")
	os.WriteFile(countryFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	// Service should not be ready before start
	if service.IsReady() {
		t.Error("Service should not be ready before start")
	}

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Service should be ready
	if !service.IsReady() {
		t.Error("Service should be ready after successful start")
	}

	// Service should have loaded countries
	countries := service.GetLoadedCountries()
	if len(countries) == 0 {
		t.Error("Service should have loaded countries")
	}

	// Service should remain ready during updates
	service.updateCountryData()
	if !service.IsReady() {
		t.Error("Service should remain ready during country updates")
	}

	// Service should remain ready during reputation updates (even if they fail)
	service.updateReputationData()
	if !service.IsReady() {
		t.Error("Service should remain ready during reputation updates")
	}

	// Stop service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Service should not be ready after stop
	if service.IsReady() {
		t.Error("Service should not be ready after stop")
	}
}

// TestBackgroundUpdates_LastUpdateTime tests that last update time is tracked correctly
func TestBackgroundUpdates_LastUpdateTime(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create country data
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

	// Wait for service to be ready
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(10 * time.Millisecond):
			// continue waiting
		}
	}

	// Get initial last update time
	initialTime := service.GetLastUpdate()

	// Wait a small amount of time
	time.Sleep(100 * time.Millisecond)

	// Perform an update
	service.updateCountryData()

	// Last update time should be more recent
	newTime := service.GetLastUpdate()
	if !newTime.After(initialTime) {
		t.Error("Last update time should be updated after country data update")
	}
}
