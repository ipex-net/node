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

// TestIntegration_FullServiceWorkflow tests the complete service workflow from startup to operation
func TestIntegration_FullServiceWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create comprehensive test data
	sourcesDir := filepath.Join(tempDir, "sources")

	// Create reputation data for all levels
	reputationData := map[int]string{
		1: "127.0.0.1\n192.168.1.0/24\n10.0.0.0/8\n",
		2: "203.0.113.1\n198.51.100.0/24\n",
		3: "172.16.0.1\n172.16.0.0/12\n",
		4: "8.8.8.8\n8.8.4.4\n",
	}

	for level, data := range reputationData {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		os.MkdirAll(filepath.Dir(repFile), 0755)
		os.WriteFile(repFile, []byte(data), 0644)
	}

	// Create geolocation data for multiple countries and providers
	geoData := map[string]map[string]string{
		"geolite2_country": {
			"country_us.netset": "192.168.1.0/24\n10.0.0.0/8\n",
			"country_gb.netset": "203.0.113.0/24\n",
			"country_de.netset": "198.51.100.0/24\n",
		},
		"ip2location_country": {
			"ip2location_country_us.netset": "192.168.2.0/24\n172.16.0.0/12\n",
			"ip2location_country_gb.netset": "203.0.113.128/25\n",
			"ip2location_country_fr.netset": "198.51.100.128/25\n",
		},
		"ipdeny_country": {
			"id_country_us.netset": "192.168.3.0/24\n",
			"id_country_ca.netset": "203.0.113.64/26\n",
		},
		"ipip_country": {
			"ipip_country_us.netset": "192.168.4.0/24\n",
			"ipip_country_jp.netset": "198.51.100.64/26\n",
		},
	}

	for provider, countries := range geoData {
		providerDir := filepath.Join(sourcesDir, provider)
		os.MkdirAll(providerDir, 0755)
		for file, data := range countries {
			countryFile := filepath.Join(providerDir, file)
			os.WriteFile(countryFile, []byte(data), 0644)
		}
	}

	// Create ASN DROP list data
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	asnData := `{"asn": 64512, "rir": "IANA", "domain": "test1.com", "cc": "US", "asname": "TEST-ASN-1", "type": "hosting"}
{"asn": 64513, "rir": "IANA", "domain": "test2.com", "cc": "US", "asname": "TEST-ASN-2", "type": "hosting"}
{"asn": 64514, "rir": "IANA", "domain": "test3.com", "cc": "GB", "asname": "TEST-ASN-3", "type": "hosting"}
`
	os.WriteFile(asnFile, []byte(asnData), 0644)

	// Start the service
	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for service to be ready
	timeout := time.After(10 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready in time")
		case <-time.After(100 * time.Millisecond):
			// continue waiting
		}
	}

	// Test reputation functionality across all levels
	reputationTests := []struct {
		ip       string
		level    int
		expected int
		desc     string
	}{
		{"127.0.0.1/32", 1, 1, "Level 1 exact IP"},
		{"192.168.1.50/32", 1, 1, "Level 1 subnet IP"},
		{"203.0.113.1/32", 2, 1, "Level 2 exact IP"},
		{"172.16.5.5/32", 3, 1, "Level 3 subnet IP"},
		{"8.8.8.8/32", 4, 1, "Level 4 exact IP"},
		{"192.168.100.1/32", 1, 0, "IP not in any level"},
	}

	for _, test := range reputationTests {
		result := service.CheckReputation(test.ip, test.level)
		if result != test.expected {
			t.Errorf("Reputation test '%s': expected %d, got %d", test.desc, test.expected, result)
		}
	}

	// Test geolocation functionality across multiple providers
	countryTests := []struct {
		ip       string
		country  string
		expected bool
		desc     string
	}{
		{"192.168.1.1", "US", true, "US IP in geolite2"},
		{"192.168.2.1", "US", true, "US IP in ip2location"},
		{"192.168.3.1", "US", true, "US IP in ipdeny"},
		{"192.168.4.1", "US", true, "US IP in ipip"},
		{"203.0.113.1", "GB", true, "GB IP in geolite2"},
		{"203.0.113.129", "GB", true, "GB IP in ip2location"},
		{"203.0.113.65", "CA", true, "CA IP in ipdeny"},
		{"198.51.100.65", "JP", true, "JP IP in ipip"},
		{"172.16.1.1", "US", true, "US IP in ip2location large subnet"},
		{"1.2.3.4", "US", false, "IP not in any country"},
		{"192.168.1.1", "FR", false, "US IP checked for wrong country"},
	}

	for _, test := range countryTests {
		result := service.CheckCountry(test.ip, test.country)
		if result != test.expected {
			t.Errorf("Country test '%s': expected %v, got %v", test.desc, test.expected, result)
		}
	}

	// Test ASN DROP list functionality
	asnTests := []struct {
		asn      int32
		expected bool
		desc     string
	}{
		{64512, true, "ASN in drop list"},
		{64513, true, "Another ASN in drop list"},
		{64514, true, "Third ASN in drop list"},
		{64515, false, "ASN not in drop list"},
		{12345, false, "Random ASN not in drop list"},
	}

	for _, test := range asnTests {
		result := service.CheckASNDrop(test.asn)
		if result != test.expected {
			t.Errorf("ASN test '%s': expected %v, got %v", test.desc, test.expected, result)
		}
	}

	// Test service status and metadata
	loadedLevels := service.GetLoadedLevels()
	if len(loadedLevels) != 4 {
		t.Errorf("Expected 4 loaded reputation levels, got %d", len(loadedLevels))
	}

	loadedCountries := service.GetLoadedCountries()
	expectedCountries := []string{"us", "gb", "de", "fr", "ca", "jp"}
	if len(loadedCountries) < len(expectedCountries) {
		t.Errorf("Expected at least %d loaded countries, got %d: %v", len(expectedCountries), len(loadedCountries), loadedCountries)
	}

	lastUpdate := service.GetLastUpdate()
	if lastUpdate.IsZero() {
		t.Error("Last update time should not be zero")
	}

	// Test that service remains ready throughout operations
	if !service.IsReady() {
		t.Error("Service should remain ready after all operations")
	}
}

// TestIntegration_ComponentInteractions tests interactions between different components
func TestIntegration_ComponentInteractions(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create overlapping test data to test component independence
	sourcesDir := filepath.Join(tempDir, "sources")

	// Reputation data with some overlapping IPs
	repData := "192.168.1.1\n10.0.0.0/8\n203.0.113.0/24\n"
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte(repData), 0644)

	// Geolocation data with some of the same IPs
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoData := "192.168.1.0/24\n203.0.113.0/24\n172.16.0.0/12\n"
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(50 * time.Millisecond):
		}
	}

	// Test that overlapping IPs work independently in both systems
	overlappingIP := "192.168.1.1"

	// Should be found in reputation
	repResult := service.CheckReputation(overlappingIP+"/32", 1)
	if repResult != 1 {
		t.Error("Overlapping IP should be found in reputation system")
	}

	// Should also be found in geolocation
	geoResult := service.CheckCountry(overlappingIP, "US")
	if !geoResult {
		t.Error("Overlapping IP should be found in geolocation system")
	}

	// Test IP that's only in one system
	repOnlyIP := "10.5.5.5"

	repResult = service.CheckReputation(repOnlyIP+"/32", 1)
	if repResult != 1 {
		t.Error("Reputation-only IP should be found in reputation system")
	}

	geoResult = service.CheckCountry(repOnlyIP, "US")
	if geoResult {
		t.Error("Reputation-only IP should not be found in geolocation system")
	}

	// Test IP that's only in geolocation
	geoOnlyIP := "172.16.1.1"

	repResult = service.CheckReputation(geoOnlyIP+"/32", 1)
	if repResult != 0 {
		t.Error("Geolocation-only IP should not be found in reputation system")
	}

	geoResult = service.CheckCountry(geoOnlyIP, "US")
	if !geoResult {
		t.Error("Geolocation-only IP should be found in geolocation system")
	}
}

// TestIntegration_ConcurrentOperations tests concurrent access to all service functions
func TestIntegration_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create comprehensive test data
	sourcesDir := filepath.Join(tempDir, "sources")

	// Reputation data
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte("127.0.0.1\n192.168.1.0/24\n"), 0644)

	// Geolocation data
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte("192.168.1.0/24\n10.0.0.0/8\n"), 0644)

	// ASN data
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	asnData := `{"asn": 64512, "rir": "IANA", "domain": "test.com", "cc": "US", "asname": "TEST-ASN", "type": "hosting"}
`
	os.WriteFile(asnFile, []byte(asnData), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(50 * time.Millisecond):
		}
	}

	// Run concurrent operations
	const numGoroutines = 20
	const numOperations = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Concurrent reputation checks
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				result := service.CheckReputation("127.0.0.1/32", 1)
				if result != 1 {
					errors <- fmt.Errorf("concurrent reputation check failed: expected 1, got %d", result)
					return
				}
			}
		}()
	}

	// Concurrent country checks
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				result := service.CheckCountry("192.168.1.1", "US")
				if !result {
					errors <- fmt.Errorf("concurrent country check failed")
					return
				}
			}
		}()
	}

	// Concurrent ASN checks
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				result := service.CheckASNDrop(64512)
				if !result {
					errors <- fmt.Errorf("concurrent ASN check failed")
					return
				}
			}
		}()
	}

	// Concurrent status checks
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				if !service.IsReady() {
					errors <- fmt.Errorf("service should remain ready during concurrent operations")
					return
				}
				service.GetLoadedLevels()
				service.GetLoadedCountries()
				service.GetLastUpdate()
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}

	// Service should still be functional
	if !service.IsReady() {
		t.Error("Service should remain ready after concurrent operations")
	}
}

// TestIntegration_DataConsistency tests data consistency across service operations
func TestIntegration_DataConsistency(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create comprehensive initial data
	sourcesDir := filepath.Join(tempDir, "sources")

	// Reputation data for all levels
	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		os.MkdirAll(filepath.Dir(repFile), 0755)
		data := fmt.Sprintf("127.0.0.%d\n192.168.%d.0/24\n", level, level)
		os.WriteFile(repFile, []byte(data), 0644)
	}

	// Geolocation data
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte("192.168.1.0/24\n10.0.0.0/8\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(50 * time.Millisecond):
		}
	}

	// Test initial state consistency
	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Initial reputation check should work")
	}

	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Initial country check should work")
	}

	// Test that update calls don't crash the service and maintain consistency
	// Note: These may succeed and download real data, or fail - both are valid
	service.updateReputationData()
	service.updateCountryData()
	service.updateASNData()

	// Wait for updates to complete
	time.Sleep(500 * time.Millisecond)

	// Test that service maintains readiness regardless of update success/failure
	if !service.IsReady() {
		t.Error("Service should remain ready after update attempts")
	}

	// Test that service remains functional (some data should be accessible)
	hasReputationData := false
	for level := 1; level <= 4; level++ {
		if service.CheckReputation("127.0.0.1/32", level) > 0 {
			hasReputationData = true
			break
		}
	}
	if !hasReputationData {
		t.Error("Service should have some reputation data available")
	}

	hasCountryData := false
	testIPs := []string{"192.168.1.1", "10.1.1.1", "8.8.8.8", "1.1.1.1"}
	for _, ip := range testIPs {
		if service.CheckCountry(ip, "US") {
			hasCountryData = true
			break
		}
	}
	if !hasCountryData {
		t.Error("Service should have some country data available")
	}

	// Test that essential service functions work
	levels := service.GetLoadedLevels()
	countries := service.GetLoadedCountries()
	finalUpdate := service.GetLastUpdate()

	if len(levels) == 0 {
		t.Error("Service should have some reputation levels loaded")
	}

	if len(countries) == 0 {
		t.Error("Service should have some countries loaded")
	}

	// Last update time should be reasonable
	if finalUpdate.IsZero() {
		t.Error("Last update time should not be zero")
	}
}

// TestIntegration_ServiceResilience tests service resilience under various conditions
func TestIntegration_ServiceResilience(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create initial valid data
	sourcesDir := filepath.Join(tempDir, "sources")

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte("127.0.0.1\n"), 0644)

	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(5 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(50 * time.Millisecond):
		}
	}

	// Test initial functionality
	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Initial reputation check should work")
	}

	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Initial country check should work")
	}

	// Test resilience to file corruption
	os.WriteFile(repFile, []byte("corrupted data \x00\x01\x02"), 0644)
	service.updateReputationData()

	// Service should remain functional with previous good data
	if !service.IsReady() {
		t.Error("Service should remain ready despite corrupted reputation data")
	}

	// Country functionality should still work
	if !service.CheckCountry("192.168.1.1", "US") {
		t.Error("Country check should still work despite reputation data corruption")
	}

	// Test resilience to file deletion
	os.Remove(geoFile)
	service.updateCountryData()

	// Service readiness might change (depends on implementation)
	// but it should not crash or panic
	service.CheckCountry("192.168.1.1", "US") // Should not panic

	// Test resilience to invalid operations
	service.CheckReputation("invalid-ip", 1)     // Should not panic
	service.CheckCountry("invalid-ip", "US")     // Should not panic
	service.CheckASNDrop(-1)                     // Should not panic
	service.CheckReputation("127.0.0.1/32", 999) // Should not panic

	// Service should still respond to status queries
	service.GetLoadedLevels()    // Should not panic
	service.GetLoadedCountries() // Should not panic
	service.GetLastUpdate()      // Should not panic
	service.IsReady()            // Should not panic
	service.IsRunning()          // Should not panic
}

// TestIntegration_EndToEndScenarios tests realistic end-to-end usage scenarios
func TestIntegration_EndToEndScenarios(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create realistic data scenario
	sourcesDir := filepath.Join(tempDir, "sources")

	// Reputation data for multiple levels (simulating real firehol data)
	reputationData := map[int]string{
		1: "# Firehol Level 1 - Most dangerous\n1.2.3.4\n5.6.7.8\n192.168.100.0/24\n",
		2: "# Firehol Level 2 - Dangerous\n10.20.30.40\n172.16.100.0/24\n",
		3: "# Firehol Level 3 - Suspicious\n203.0.113.100\n198.51.100.100\n",
		4: "# Firehol Level 4 - Monitoring\n8.8.8.8\n1.1.1.1\n",
	}

	for level, data := range reputationData {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		os.MkdirAll(filepath.Dir(repFile), 0755)
		os.WriteFile(repFile, []byte(data), 0644)
	}

	// Geolocation data (simulating real geo databases)
	providers := map[string]map[string]string{
		"geolite2_country": {
			"country_us.netset": "192.168.0.0/16\n10.0.0.0/8\n",
			"country_gb.netset": "203.0.113.0/24\n",
			"country_de.netset": "198.51.100.0/24\n",
		},
		"ip2location_country": {
			"ip2location_country_us.netset": "172.16.0.0/12\n192.168.0.0/16\n",
			"ip2location_country_cn.netset": "1.2.3.0/24\n",
			"ip2location_country_ru.netset": "5.6.7.0/24\n",
		},
	}

	for provider, countries := range providers {
		providerDir := filepath.Join(sourcesDir, provider)
		os.MkdirAll(providerDir, 0755)
		for file, data := range countries {
			countryFile := filepath.Join(providerDir, file)
			os.WriteFile(countryFile, []byte(data), 0644)
		}
	}

	// ASN DROP list (simulating real Spamhaus data)
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	asnData := `{"asn": 64496, "rir": "IANA", "domain": "example.com", "cc": "US", "asname": "RESERVED-1", "type": "hosting"}
{"asn": 64497, "rir": "IANA", "domain": "example.org", "cc": "GB", "asname": "RESERVED-2", "type": "hosting"}
{"asn": 64500, "rir": "IANA", "domain": "badactor.net", "cc": "XX", "asname": "BAD-ASN", "type": "hosting"}
`
	os.WriteFile(asnFile, []byte(asnData), 0644)

	service := NewService(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for full initialization
	timeout := time.After(10 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Scenario 1: Threat Analysis - Check if an IP is suspicious across multiple dimensions
	suspiciousIP := "1.2.3.4"

	// Check reputation across all levels
	threatLevel := 0
	for level := 1; level <= 4; level++ {
		if service.CheckReputation(suspiciousIP+"/32", level) > 0 {
			threatLevel = level
			break
		}
	}

	if threatLevel != 1 {
		t.Errorf("Suspicious IP should be found in threat level 1, found in level %d", threatLevel)
	}

	// Check country of origin
	isUSIP := service.CheckCountry(suspiciousIP, "US")
	isCNIP := service.CheckCountry(suspiciousIP, "CN")

	if !isCNIP {
		t.Error("Suspicious IP should be identified as from China")
	}
	if isUSIP {
		t.Error("Suspicious IP should not be identified as from US")
	}

	// Scenario 2: Network Analysis - Analyze a subnet for threats
	subnetAnalysis := []string{
		"192.168.100.1", "192.168.100.50", "192.168.100.100", "192.168.100.200",
	}

	threatsFound := 0
	for _, ip := range subnetAnalysis {
		if service.CheckReputation(ip+"/32", 1) > 0 {
			threatsFound++
		}
	}

	// Should find threats in the subnet (192.168.100.0/24 is in level 1)
	if threatsFound != len(subnetAnalysis) {
		t.Errorf("Expected all IPs in subnet to be flagged as threats, found %d/%d", threatsFound, len(subnetAnalysis))
	}

	// Scenario 3: ASN Analysis - Check if hosting providers are suspicious
	suspiciousASNs := []int32{64496, 64497, 64500}
	flaggedASNs := 0

	for _, asn := range suspiciousASNs {
		if service.CheckASNDrop(asn) {
			flaggedASNs++
		}
	}

	if flaggedASNs != len(suspiciousASNs) {
		t.Errorf("Expected all test ASNs to be flagged, found %d/%d", flaggedASNs, len(suspiciousASNs))
	}

	// Scenario 4: Comprehensive IP Intelligence
	testIP := "192.168.1.1"

	// Get all available intelligence
	repLevel1 := service.CheckReputation(testIP+"/32", 1)
	repLevel2 := service.CheckReputation(testIP+"/32", 2)
	isUS := service.CheckCountry(testIP, "US")
	isGB := service.CheckCountry(testIP, "GB")

	// This IP should be clean (not in reputation) but geolocated to US
	if repLevel1 > 0 || repLevel2 > 0 {
		t.Error("Clean IP should not appear in reputation lists")
	}

	if !isUS {
		t.Error("Test IP should be geolocated to US")
	}

	if isGB {
		t.Error("Test IP should not be geolocated to GB")
	}

	// Verify service statistics
	levels := service.GetLoadedLevels()
	countries := service.GetLoadedCountries()

	if len(levels) != 4 {
		t.Errorf("Expected 4 reputation levels loaded, got %d", len(levels))
	}

	if len(countries) < 5 {
		t.Errorf("Expected at least 5 countries loaded, got %d", len(countries))
	}

	lastUpdate := service.GetLastUpdate()
	if time.Since(lastUpdate) > time.Hour {
		t.Error("Last update should be recent")
	}
}
