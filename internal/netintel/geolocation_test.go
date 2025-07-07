package netintel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGeolocationClient_BasicOperations tests basic geolocation client functionality
func TestGeolocationClient_BasicOperations(t *testing.T) {
	tempDir := t.TempDir()
	client := NewGeolocationClient(tempDir)

	// Initially, no countries should be loaded
	countries := client.GetLoadedCountries()
	if len(countries) != 0 {
		t.Errorf("Expected 0 loaded countries initially, got %d", len(countries))
	}

	// Lookup on unloaded country should return false
	result := client.CheckCountry("127.0.0.1", "US")
	if result {
		t.Error("Lookup on unloaded country should return false")
	}
}

// TestGeolocationClient_CountryLoading tests loading country data from files
func TestGeolocationClient_CountryLoading(t *testing.T) {
	tempDir := t.TempDir()

	// Create the expected directory structure for geolocation data
	sources := map[string]string{
		"geolite2_country":    "country_us.netset",
		"ip2location_country": "ip2location_country_us.netset",
		"ipdeny_country":      "id_country_us.netset",
		"ipip_country":        "ipip_country_us.netset",
	}

	for providerDir, fileName := range sources {
		fullProviderDir := filepath.Join(tempDir, "sources", providerDir)
		err := os.MkdirAll(fullProviderDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", providerDir, err)
		}

		// Create test country file for US
		testData := `# Test geolocation data for US
192.168.1.0/24
10.0.0.0/8
203.0.113.0/24
`
		countryFile := filepath.Join(fullProviderDir, fileName)
		err = os.WriteFile(countryFile, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test country file %s: %v", countryFile, err)
		}
	}

	client := NewGeolocationClient(tempDir)

	// Load US country data
	err := client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Failed to load country US: %v", err)
	}

	// Check that US is loaded
	countries := client.GetLoadedCountries()
	if len(countries) != 1 || countries[0] != "us" {
		t.Errorf("Expected 'us' to be loaded, got countries: %v", countries)
	}
}

// TestGeolocationClient_CountryLookups tests country-based IP lookups
func TestGeolocationClient_CountryLookups(t *testing.T) {
	tempDir := t.TempDir()

	// Create test data for multiple providers with different coverage
	testCases := map[string]struct {
		dir  string
		file string
		data string
	}{
		"geolite2": {
			"geolite2_country", "country_us.netset",
			`192.168.1.0/24
203.0.113.0/24`,
		},
		"ip2location": {
			"ip2location_country", "ip2location_country_us.netset",
			`10.0.0.0/8
203.0.113.0/24`,
		},
		"ipdeny": {
			"ipdeny_country", "id_country_us.netset",
			`172.16.0.0/12`,
		},
		"ipip": {
			"ipip_country", "ipip_country_us.netset",
			`198.51.100.0/24`,
		},
	}

	for provider, config := range testCases {
		providerDir := filepath.Join(tempDir, "sources", config.dir)
		err := os.MkdirAll(providerDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", provider, err)
		}

		countryFile := filepath.Join(providerDir, config.file)
		err = os.WriteFile(countryFile, []byte(config.data), 0644)
		if err != nil {
			t.Fatalf("Failed to create country file for %s: %v", provider, err)
		}
	}

	client := NewGeolocationClient(tempDir)
	err := client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Failed to load country data: %v", err)
	}

	// Test lookups - should return true if ANY provider has the IP
	lookupTests := []struct {
		ip       string
		expected bool
		desc     string
	}{
		{"192.168.1.1", true, "IP in geolite2"},
		{"10.5.5.5", true, "IP in ip2location"},
		{"172.16.1.1", true, "IP in ipdeny"},
		{"198.51.100.1", true, "IP in ipip"},
		{"203.0.113.1", true, "IP in both geolite2 and ip2location"},
		{"8.8.8.8", false, "IP not in any provider"},
		{"192.168.2.1", false, "IP in similar but different subnet"},
	}

	for _, test := range lookupTests {
		t.Run(test.desc, func(t *testing.T) {
			result := client.CheckCountry(test.ip, "US")
			if result != test.expected {
				t.Errorf("CheckCountry(%s, US) = %v, expected %v (%s)", test.ip, result, test.expected, test.desc)
			}
		})
	}
}

// TestGeolocationClient_CaseInsensitiveCountry tests case insensitive country codes
func TestGeolocationClient_CaseInsensitiveCountry(t *testing.T) {
	tempDir := t.TempDir()

	// Create test data for US
	providerDir := filepath.Join(tempDir, "sources", "geolite2_country")
	err := os.MkdirAll(providerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}

	testData := "192.168.1.0/24\n"
	countryFile := filepath.Join(providerDir, "country_us.netset")
	err = os.WriteFile(countryFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create country file: %v", err)
	}

	client := NewGeolocationClient(tempDir)

	// Load with uppercase
	err = client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Failed to load country US: %v", err)
	}

	// Test lookups with different case variations
	caseTests := []string{"US", "us", "Us", "uS"}
	for _, countryCode := range caseTests {
		t.Run(fmt.Sprintf("case_%s", countryCode), func(t *testing.T) {
			result := client.CheckCountry("192.168.1.1", countryCode)
			if !result {
				t.Errorf("CheckCountry should be case insensitive for country code: %s", countryCode)
			}
		})
	}
}

// TestGeolocationClient_MultipleCountries tests loading multiple countries
func TestGeolocationClient_MultipleCountries(t *testing.T) {
	tempDir := t.TempDir()

	countries := map[string]string{
		"us": "192.168.1.0/24\n",
		"gb": "10.0.0.0/24\n",
		"de": "172.16.0.0/24\n",
	}

	// Create data for each country across all providers
	providers := map[string]string{
		"geolite2_country":    "country_%s.netset",
		"ip2location_country": "ip2location_country_%s.netset",
	}
	for provider, fileFormat := range providers {
		providerDir := filepath.Join(tempDir, "sources", provider)
		err := os.MkdirAll(providerDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", provider, err)
		}

		for country, data := range countries {
			countryFile := filepath.Join(providerDir, fmt.Sprintf(fileFormat, country))
			err = os.WriteFile(countryFile, []byte(data), 0644)
			if err != nil {
				t.Fatalf("Failed to create country file %s for %s: %v", country, provider, err)
			}
		}
	}

	client := NewGeolocationClient(tempDir)

	// Load each country
	for country := range countries {
		err := client.LoadCountry(strings.ToUpper(country))
		if err != nil {
			t.Fatalf("Failed to load country %s: %v", country, err)
		}
	}

	// Check all countries are loaded
	loadedCountries := client.GetLoadedCountries()
	if len(loadedCountries) != 3 {
		t.Errorf("Expected 3 countries loaded, got %d: %v", len(loadedCountries), loadedCountries)
	}

	// Test country-specific lookups
	countryTests := []struct {
		ip       string
		country  string
		expected bool
	}{
		{"192.168.1.1", "US", true},
		{"192.168.1.1", "GB", false},
		{"10.0.0.1", "GB", true},
		{"10.0.0.1", "US", false},
		{"172.16.0.1", "DE", true},
		{"172.16.0.1", "US", false},
	}

	for _, test := range countryTests {
		result := client.CheckCountry(test.ip, test.country)
		if result != test.expected {
			t.Errorf("CheckCountry(%s, %s) = %v, expected %v", test.ip, test.country, result, test.expected)
		}
	}
}

// TestGeolocationClient_LoadAllCountries tests loading all available countries
func TestGeolocationClient_LoadAllCountries(t *testing.T) {
	tempDir := t.TempDir()

	// Create data for a few countries (subset of all possible countries)
	countries := []string{"us", "gb", "de", "fr"}
	providers := map[string]string{
		"geolite2_country":    "country_%s.netset",
		"ip2location_country": "ip2location_country_%s.netset",
	}

	for provider, fileFormat := range providers {
		providerDir := filepath.Join(tempDir, "sources", provider)
		err := os.MkdirAll(providerDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", provider, err)
		}

		for i, country := range countries {
			// Create different data for each country
			data := fmt.Sprintf("192.168.%d.0/24\n", i+1)
			countryFile := filepath.Join(providerDir, fmt.Sprintf(fileFormat, country))
			err = os.WriteFile(countryFile, []byte(data), 0644)
			if err != nil {
				t.Fatalf("Failed to create country file %s for %s: %v", country, provider, err)
			}
		}
	}

	client := NewGeolocationClient(tempDir)

	// LoadAllCountries should load only existing files
	err := client.LoadAllCountries()
	if err != nil {
		t.Fatalf("LoadAllCountries failed: %v", err)
	}

	// Should have loaded all available countries
	loadedCountries := client.GetLoadedCountries()
	if len(loadedCountries) != len(countries) {
		t.Errorf("Expected %d countries, got %d: %v", len(countries), len(loadedCountries), loadedCountries)
	}

	// Verify each country works
	for i, country := range countries {
		testIP := fmt.Sprintf("192.168.%d.1", i+1)
		if !client.CheckCountry(testIP, strings.ToUpper(country)) {
			t.Errorf("Country %s should be loaded and working", country)
		}
	}
}

// TestGeolocationClient_PartialProviderData tests handling when only some providers have data
func TestGeolocationClient_PartialProviderData(t *testing.T) {
	tempDir := t.TempDir()

	// Only create data for 2 out of 4 providers
	providerData := map[string]string{
		"geolite2_country":    "192.168.1.0/24\n",
		"ip2location_country": "10.0.0.0/24\n",
		// ipdeny and ipip directories will exist but have no files
	}

	allProviders := map[string]string{
		"geolite2_country":    "country_us.netset",
		"ip2location_country": "ip2location_country_us.netset",
		"ipdeny_country":      "id_country_us.netset",
		"ipip_country":        "ipip_country_us.netset",
	}
	for provider, fileName := range allProviders {
		providerDir := filepath.Join(tempDir, "sources", provider)
		err := os.MkdirAll(providerDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", provider, err)
		}

		if data, exists := providerData[provider]; exists {
			countryFile := filepath.Join(providerDir, fileName)
			err = os.WriteFile(countryFile, []byte(data), 0644)
			if err != nil {
				t.Fatalf("Failed to create country file for %s: %v", provider, err)
			}
		}
	}

	client := NewGeolocationClient(tempDir)

	// Should succeed even with partial data
	err := client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Loading should succeed with partial provider data: %v", err)
	}

	// Should work with data from available providers
	if !client.CheckCountry("192.168.1.1", "US") {
		t.Error("Should find IP in geolite2 data")
	}
	if !client.CheckCountry("10.0.0.1", "US") {
		t.Error("Should find IP in ip2location data")
	}
	if client.CheckCountry("172.16.0.1", "US") {
		t.Error("Should not find IP in missing provider data")
	}
}

// TestGeolocationClient_InvalidData tests handling of invalid data
func TestGeolocationClient_InvalidData(t *testing.T) {
	tempDir := t.TempDir()

	providerDir := filepath.Join(tempDir, "sources", "geolite2_country")
	err := os.MkdirAll(providerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}

	// Create file with mixed valid and invalid data
	testData := `# Valid entries
192.168.1.0/24
10.0.0.0/8

# Invalid entries that should be skipped
not-an-ip
999.999.999.999
192.168.1.0/99
invalid/cidr

# More valid entries
203.0.113.0/24
`

	countryFile := filepath.Join(providerDir, "country_us.netset")
	err = os.WriteFile(countryFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test country file: %v", err)
	}

	client := NewGeolocationClient(tempDir)

	// Loading should succeed despite invalid entries
	err = client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Loading should succeed despite invalid entries: %v", err)
	}

	// Valid entries should still work
	if !client.CheckCountry("192.168.1.1", "US") {
		t.Error("Valid IP should be found")
	}
	if !client.CheckCountry("10.5.5.5", "US") {
		t.Error("IP in valid large subnet should be found")
	}
	if !client.CheckCountry("203.0.113.1", "US") {
		t.Error("Valid IP should be found")
	}
}

// TestGeolocationClient_EmptyAndMissingFiles tests handling of empty and missing files
func TestGeolocationClient_EmptyAndMissingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create provider directories
	providers := []string{"geolite2_country", "ip2location_country", "ipdeny_country", "ipip_country"}
	for _, provider := range providers {
		providerDir := filepath.Join(tempDir, "sources", provider)
		err := os.MkdirAll(providerDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create provider directory %s: %v", provider, err)
		}
	}

	client := NewGeolocationClient(tempDir)

	// Test loading missing country files - should fail (no files found)
	err := client.LoadCountry("XX")
	if err == nil {
		t.Error("Loading missing country should return error")
	}

	// Create empty file for one provider
	emptyFile := filepath.Join(tempDir, "sources", "geolite2_country", "country_yy.netset")
	err = os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Loading country with empty file should succeed
	err = client.LoadCountry("YY")
	if err != nil {
		t.Fatalf("Loading country with empty file should succeed: %v", err)
	}

	// No lookups should match
	if client.CheckCountry("192.168.1.1", "YY") {
		t.Error("Lookup in empty country should return false")
	}

	// Create file with only comments
	commentsOnlyData := `# This file only has comments
# No actual IP data
`
	commentsFile := filepath.Join(tempDir, "sources", "geolite2_country", "country_zz.netset")
	err = os.WriteFile(commentsFile, []byte(commentsOnlyData), 0644)
	if err != nil {
		t.Fatalf("Failed to create comments-only file: %v", err)
	}

	err = client.LoadCountry("ZZ")
	if err != nil {
		t.Fatalf("Loading comments-only country should succeed: %v", err)
	}

	if client.CheckCountry("192.168.1.1", "ZZ") {
		t.Error("Lookup in comments-only country should return false")
	}
}

// TestGeolocationClient_IPv6Support tests IPv6 address handling
func TestGeolocationClient_IPv6Support(t *testing.T) {
	tempDir := t.TempDir()

	providerDir := filepath.Join(tempDir, "sources", "geolite2_country")
	err := os.MkdirAll(providerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}

	// Create test data with IPv6 addresses
	testData := `2001:db8::1/128
2001:db8::/32
::1/128
fe80::/64`

	countryFile := filepath.Join(providerDir, "country_us.netset")
	err = os.WriteFile(countryFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test country file: %v", err)
	}

	client := NewGeolocationClient(tempDir)
	err = client.LoadCountry("US")
	if err != nil {
		t.Fatalf("Failed to load country: %v", err)
	}

	ipv6Tests := []struct {
		ip       string
		expected bool
		desc     string
	}{
		{"2001:db8::1", true, "exact IPv6 match"},
		{"2001:db8::2", true, "IPv6 in subnet"},
		{"2001:db8:1::1", true, "IPv6 in larger subnet"},
		{"::1", true, "localhost IPv6"},
		{"fe80::1", true, "link-local IPv6"},
		{"2001:db9::1", false, "IPv6 not in range"},
		{"192.168.1.1", false, "IPv4 should not match IPv6"},
	}

	for _, test := range ipv6Tests {
		t.Run(test.desc, func(t *testing.T) {
			result := client.CheckCountry(test.ip, "US")
			if result != test.expected {
				t.Errorf("CheckCountry(%s, US) = %v, expected %v (%s)", test.ip, result, test.expected, test.desc)
			}
		})
	}
}

// TestGeolocationClient_ConcurrentAccess tests thread safety
func TestGeolocationClient_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()

	// Create test data for multiple countries
	countries := map[string]string{
		"us": "192.168.1.0/24\n",
		"gb": "10.0.0.0/24\n",
		"de": "172.16.0.0/24\n",
	}

	providerDir := filepath.Join(tempDir, "sources", "geolite2_country")
	err := os.MkdirAll(providerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}

	for country, data := range countries {
		countryFile := filepath.Join(providerDir, fmt.Sprintf("country_%s.netset", country))
		err = os.WriteFile(countryFile, []byte(data), 0644)
		if err != nil {
			t.Fatalf("Failed to create country file %s: %v", country, err)
		}
	}

	client := NewGeolocationClient(tempDir)

	// Load all countries
	for country := range countries {
		err = client.LoadCountry(strings.ToUpper(country))
		if err != nil {
			t.Fatalf("Failed to load country %s: %v", country, err)
		}
	}

	// Test concurrent lookups
	const numGoroutines = 50
	const numLookups = 10

	testCases := []struct {
		ip       string
		country  string
		expected bool
	}{
		{"192.168.1.1", "US", true},
		{"10.0.0.1", "GB", true},
		{"172.16.0.1", "DE", true},
		{"8.8.8.8", "US", false},
		{"192.168.1.1", "GB", false},
	}

	results := make(chan bool, numGoroutines*numLookups*len(testCases))

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numLookups; j++ {
				for _, test := range testCases {
					result := client.CheckCountry(test.ip, test.country)
					results <- (result == test.expected)
				}
			}
		}()
	}

	// Collect all results
	totalTests := numGoroutines * numLookups * len(testCases)
	successCount := 0
	for i := 0; i < totalTests; i++ {
		if <-results {
			successCount++
		}
	}

	if successCount != totalTests {
		t.Errorf("Concurrent access failed: %d/%d tests passed", successCount, totalTests)
	}
}
