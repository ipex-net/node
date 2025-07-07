package netintel

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
)

// TestService_Lifecycle tests the complete service lifecycle
func TestService_Lifecycle(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Test initial state
	if service.IsRunning() {
		t.Error("Service should not be running initially")
	}
	if service.IsReady() {
		t.Error("Service should not be ready initially")
	}

	// Test successful startup
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Service should be running immediately after start
	if !service.IsRunning() {
		t.Error("Service should be running after successful start")
	}

	// Wait a short time for initial load to potentially complete
	time.Sleep(100 * time.Millisecond)

	// Test graceful shutdown
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Service should not be running after stop
	if service.IsRunning() {
		t.Error("Service should not be running after stop")
	}
}

func TestService_DoubleStart(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Start service first time
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service first time: %v", err)
	}
	defer service.Stop()

	// Try to start again - should fail
	err = service.Start()
	if err == nil {
		t.Error("Expected error when starting service twice, got nil")
	}
}

func TestService_MultipleStop(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Start service
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Stop first time
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service first time: %v", err)
	}

	// Stop second time - should not error
	err = service.Stop()
	if err != nil {
		t.Errorf("Multiple stop calls should not error: %v", err)
	}
}

func TestService_StopWithoutStart(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Stop without starting - should not error
	err := service.Stop()
	if err != nil {
		t.Errorf("Stop without start should not error: %v", err)
	}
}

func TestService_ConcurrentStartStop(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	var wg sync.WaitGroup
	const numGoroutines = 10

	// Start multiple goroutines trying to start/stop
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%2 == 0 {
				service.Start()
			} else {
				service.Stop()
			}
		}(i)
	}

	wg.Wait()

	// Clean up - ensure service is stopped
	service.Stop()
}

func TestService_StateTransitions(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Initial state: not running, not ready
	if service.IsRunning() || service.IsReady() {
		t.Error("Service should not be running or ready initially")
	}

	// Start service
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Should be running, but readiness depends on data loading
	if !service.IsRunning() {
		t.Error("Service should be running after start")
	}

	// Test status information
	status := service.GetStatus()
	if !status.Running {
		t.Error("Status should show service as running")
	}

	// Test individual methods work without panic
	_ = service.CheckReputation("127.0.0.1", 1)
	_ = service.ScanSubnetReputation("127.0.0.0/24", 1)
	_ = service.CheckCountry("127.0.0.1", "US")
	_ = service.CheckASNDrop(64512)
	_ = service.GetLastUpdate()
	_ = service.GetLoadedLevels()
	_ = service.GetLoadedCountries()

	// Stop service
	err = service.Stop()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Should not be running or ready after stop
	if service.IsRunning() || service.IsReady() {
		t.Error("Service should not be running or ready after stop")
	}
}

func TestService_InitialDataLoading(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create some test data files to simulate existing data
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create a simple reputation file
	repData := `# Test reputation data
127.0.0.1/32
192.168.1.0/24
`
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(repData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	// Create ASN drop list file
	asnData := `{"asn":64512,"rir":"iana","domain":"test","cc":"US","asname":"TEST-ASN"}
{"asn":64513,"rir":"iana","domain":"test2","cc":"GB","asname":"TEST-ASN-2"}`
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	err = os.WriteFile(asnFile, []byte(asnData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ASN file: %v", err)
	}

	// Create geolocation data to make service ready
	geoData := "1.1.1.1/32\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewServiceForTesting(tempDir, logger)

	// Start service
	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Give service time to load data
	time.Sleep(200 * time.Millisecond)

	// Check that data was loaded
	levels := service.GetLoadedLevels()
	if len(levels) == 0 {
		t.Error("Expected some reputation levels to be loaded")
	}

	// Test reputation lookup
	result := service.CheckReputation("127.0.0.1", 1)
	if result == 0 {
		t.Error("Expected reputation check to find test data")
	}

	// Test ASN lookup
	if !service.CheckASNDrop(64512) {
		t.Error("Expected ASN 64512 to be in drop list")
	}
	if service.CheckASNDrop(99999) {
		t.Error("Expected ASN 99999 to not be in drop list")
	}
}

func TestService_ShutdownTimeout(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Start service
	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	// Stop service and measure time
	start := time.Now()
	err = service.Stop()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	// Shutdown should be quick (well under the 30-second timeout)
	if duration > 5*time.Second {
		t.Errorf("Service shutdown took too long: %v", duration)
	}
}

func TestService_DataDirectoryHandling(t *testing.T) {
	// Test with non-existent directory
	nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
	logger := log.NewNopLogger()
	service := NewServiceForTesting(nonExistentDir, logger)

	// Should still be able to create service
	if service == nil {
		t.Error("Service creation should succeed even with non-existent directory")
	}

	// Starting might fail due to missing data, but shouldn't panic
	service.Start()
	service.Stop()
}

func TestService_MethodsWithoutData(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	service := NewServiceForTesting(tempDir, logger)

	// Test methods before starting service (should not panic)
	result := service.CheckReputation("127.0.0.1", 1)
	if result != 0 {
		t.Error("Expected 0 result when service not ready")
	}

	scanResult := service.ScanSubnetReputation("127.0.0.0/24", 1)
	if scanResult != 0 {
		t.Error("Expected 0 result when service not ready")
	}

	countryResult := service.CheckCountry("127.0.0.1", "US")
	if countryResult {
		t.Error("Expected false result when service not ready")
	}

	asnResult := service.CheckASNDrop(64512)
	if asnResult {
		t.Error("Expected false result when service not ready")
	}

	// These should return empty/zero values
	lastUpdate := service.GetLastUpdate()
	if !lastUpdate.IsZero() {
		t.Error("Expected zero time when no updates have occurred")
	}

	levels := service.GetLoadedLevels()
	if len(levels) != 0 {
		t.Error("Expected empty levels when no data loaded")
	}

	countries := service.GetLoadedCountries()
	if len(countries) != 0 {
		t.Error("Expected empty countries when no data loaded")
	}
}

func TestService_ReadinessLogic(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Test case 1: No data - should not be ready
	service := NewServiceForTesting(tempDir, logger)
	service.Start()
	time.Sleep(100 * time.Millisecond)
	if service.IsReady() {
		t.Error("Service should not be ready without any data")
	}
	service.Stop()

	// Test case 2: Only reputation data - should not be ready
	repData := "127.0.0.1/32\n"
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.WriteFile(repFile, []byte(repData), 0644)

	service = NewServiceForTesting(tempDir, logger)
	service.Start()

	// Wait for initial load to complete (but should not be ready)
	timeout := time.After(5 * time.Second)
	for service.IsRunning() && len(service.GetLoadedLevels()) == 0 {
		select {
		case <-timeout:
			t.Fatal("Service didn't load reputation data in time")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// Service should not be ready with only reputation data (no background updates in test mode)
	if service.IsReady() {
		t.Error("Service should not be ready with only reputation data")
	}
	service.Stop()

	// Test case 3: Only geolocation data - should be ready
	os.Remove(repFile) // Remove reputation data
	geoData := "192.168.1.0/24\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service = NewServiceForTesting(tempDir, logger)
	service.Start()

	// Wait for service to be ready with geolocation data
	timeout = time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready with geolocation data in time")
		case <-time.After(10 * time.Millisecond):
		}
	}

	if !service.IsReady() {
		t.Error("Service should be ready with only geolocation data")
	}
	service.Stop()

	// Test case 4: Both reputation and geolocation - should be ready
	os.WriteFile(repFile, []byte(repData), 0644) // Add reputation data back
	service = NewServiceForTesting(tempDir, logger)
	service.Start()
	time.Sleep(100 * time.Millisecond)
	if !service.IsReady() {
		t.Error("Service should be ready with both reputation and geolocation data")
	}
	service.Stop()
}

func TestService_ScanSubnetReputation(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create a simple reputation file
	repData := `# Test reputation data
192.168.1.10
192.168.1.20
192.168.1.30
`
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(repData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	// Create geolocation data to ensure service becomes ready
	geoData := "1.1.1.1/32\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewServiceForTesting(tempDir, logger)
	err = service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready with extended timeout and debug info
	timeout := time.After(10 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			// Add debug info on timeout
			levels := service.GetLoadedLevels()
			countries := service.GetLoadedCountries()
			t.Fatalf("Service didn't become ready in time. Loaded levels: %v, countries: %v", levels, countries)
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Test scanning a subnet
	count := service.ScanSubnetReputation("192.168.1.0/24", 1)
	if count != 3 {
		t.Errorf("Expected to find 3 reputable IPs in subnet, got %d", count)
	}

	// Test scanning a different subnet
	count = service.ScanSubnetReputation("192.168.2.0/24", 1)
	if count != 0 {
		t.Errorf("Expected to find 0 reputable IPs in subnet, got %d", count)
	}
}

func TestService_CheckASNDrop(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create ASN drop list file
	asnData := `{"asn":64512,"rir":"iana","domain":"test","cc":"US","asname":"TEST-ASN"}`
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	err = os.WriteFile(asnFile, []byte(asnData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ASN file: %v", err)
	}

	// Create geolocation data to ensure service becomes ready
	geoData := "1.1.1.1/32\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewServiceForTesting(tempDir, logger)
	err = service.Start()
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
		}
	}

	// Test checking an ASN that is in the drop list
	if !service.CheckASNDrop(64512) {
		t.Error("Expected ASN 64512 to be in the drop list")
	}

	// Test checking an ASN that is not in the drop list
	if service.CheckASNDrop(12345) {
		t.Error("Expected ASN 12345 to not be in the drop list")
	}
}

func TestService_CheckCountry(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create geolocation data to ensure service becomes ready
	geoData := "192.168.1.0/24\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewServiceForTesting(tempDir, logger)
	err = service.Start()
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
		}
	}

	// Test checking an IP that is in the country
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Expected IP to be in country US")
	}

	// Test checking an IP that is not in the country
	if service.CheckCountry("1.1.1.1", "US") {
		t.Error("Expected IP to not be in country US")
	}

	// Test checking an IP in a different country
	if service.CheckCountry("192.168.1.1", "GB") {
		t.Error("Expected IP to not be in country GB")
	}
}
