package netintel

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// TestReputationClient_BasicOperations tests basic reputation client functionality
func TestReputationClient_BasicOperations(t *testing.T) {
	tempDir := t.TempDir()
	client := NewReputationClient(tempDir)

	// Initially, no levels should be loaded
	levels := client.GetLoadedLevels()
	if len(levels) != 0 {
		t.Errorf("Expected 0 loaded levels initially, got %d", len(levels))
	}

	// Lookup on unloaded level should return 0
	result := client.Lookup(1, "127.0.0.1")
	if result != 0 {
		t.Errorf("Lookup on unloaded level should return 0, got %d", result)
	}
}

// TestReputationClient_DataLoading tests loading reputation data from files
func TestReputationClient_DataLoading(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test reputation file with various formats
	testData := `# Test reputation data - comments should be ignored
# Another comment line

# Empty lines above and below should be ignored

127.0.0.1
192.168.1.0/24
10.0.0.0/8
172.16.0.0/12

# More comments
203.0.113.0/24
198.51.100.0/24
`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)

	// Load level 1
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Failed to load reputation level 1: %v", err)
	}

	// Check that level 1 is loaded
	levels := client.GetLoadedLevels()
	if len(levels) != 1 || levels[0] != 1 {
		t.Errorf("Expected level 1 to be loaded, got levels: %v", levels)
	}
}

// TestReputationClient_IPLookups tests IP address lookups
func TestReputationClient_IPLookups(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data with specific IPs and subnets
	testData := `127.0.0.1
192.168.1.0/24
10.0.0.0/8
203.0.113.42`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Failed to load reputation level: %v", err)
	}

	testCases := []struct {
		target   string
		expected int
		desc     string
	}{
		{"127.0.0.1/32", 1, "exact IP match"},
		{"192.168.1.1/32", 1, "IP in subnet"},
		{"192.168.1.100/32", 1, "another IP in subnet"},
		{"192.168.1.255/32", 1, "broadcast IP in subnet"},
		{"10.5.10.15/32", 1, "IP in large subnet"},
		{"203.0.113.42/32", 1, "exact IP match 2"},
		{"8.8.8.8/32", 0, "IP not in any range"},
		{"192.168.2.1/32", 0, "IP in similar but different subnet"},
		{"11.0.0.1/32", 0, "IP outside large subnet"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := client.Lookup(1, tc.target)
			if result != tc.expected {
				t.Errorf("Lookup(%s) = %d, expected %d (%s)", tc.target, result, tc.expected, tc.desc)
			}
		})
	}
}

// TestReputationClient_SubnetLookups tests subnet scanning functionality
func TestReputationClient_SubnetLookups(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data
	testData := `192.168.1.50
192.168.1.100
192.168.2.25
10.0.0.0/16`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Failed to load reputation level: %v", err)
	}

	testCases := []struct {
		subnet   string
		expected int
		desc     string
	}{
		{"192.168.1.0/24", 2, "subnet with 2 matching IPs"},
		{"192.168.2.0/24", 1, "subnet with 1 matching IP"},
		{"192.168.3.0/24", 0, "subnet with no matching IPs"},
		{"10.0.0.0/24", 0, "subset of larger range should count as 0 individual IPs"},
		{"172.16.0.0/24", 0, "completely unrelated subnet"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := client.Lookup(1, tc.subnet)
			if result != tc.expected {
				t.Errorf("Lookup(%s) = %d, expected %d (%s)", tc.subnet, result, tc.expected, tc.desc)
			}
		})
	}
}

// TestReputationClient_MultipleLevels tests loading and using multiple reputation levels
func TestReputationClient_MultipleLevels(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create different data for different levels
	level1Data := `127.0.0.1
192.168.1.0/24`

	level2Data := `10.0.0.1
192.168.2.0/24`

	level3Data := `172.16.0.1
203.0.113.0/24`

	files := map[int]string{
		1: level1Data,
		2: level2Data,
		3: level3Data,
	}

	for level, data := range files {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		err = os.WriteFile(repFile, []byte(data), 0644)
		if err != nil {
			t.Fatalf("Failed to create test reputation file for level %d: %v", level, err)
		}
	}

	client := NewReputationClient(tempDir)

	// Load all levels
	for level := 1; level <= 3; level++ {
		err = client.LoadLevel(level)
		if err != nil {
			t.Fatalf("Failed to load reputation level %d: %v", level, err)
		}
	}

	// Check all levels are loaded
	levels := client.GetLoadedLevels()
	if len(levels) != 3 {
		t.Errorf("Expected 3 loaded levels, got %d", len(levels))
	}

	// Test level-specific lookups
	testCases := []struct {
		ip       string
		level    int
		expected int
		desc     string
	}{
		{"127.0.0.1", 1, 1, "IP in level 1"},
		{"127.0.0.1", 2, 0, "IP not in level 2"},
		{"127.0.0.1", 3, 0, "IP not in level 3"},
		{"10.0.0.1", 1, 0, "IP not in level 1"},
		{"10.0.0.1", 2, 1, "IP in level 2"},
		{"172.16.0.1", 3, 1, "IP in level 3"},
		{"192.168.1.1", 1, 1, "IP in level 1 subnet"},
		{"192.168.2.1", 2, 1, "IP in level 2 subnet"},
		{"203.0.113.1", 3, 1, "IP in level 3 subnet"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_level_%d", tc.desc, tc.level), func(t *testing.T) {
			result := client.Lookup(tc.level, tc.ip)
			if result != tc.expected {
				t.Errorf("Lookup(level=%d, ip=%s) = %d, expected %d", tc.level, tc.ip, result, tc.expected)
			}
		})
	}
}

// TestReputationClient_LoadAllLevels tests loading all levels at once
func TestReputationClient_LoadAllLevels(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create files for levels 1 and 3 (missing level 2 and 4)
	level1Data := "127.0.0.1\n"
	level3Data := "192.168.1.1\n"

	repFile1 := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile1, []byte(level1Data), 0644)
	if err != nil {
		t.Fatalf("Failed to create level 1 file: %v", err)
	}

	repFile3 := filepath.Join(sourcesDir, "firehol_level3.netset")
	err = os.WriteFile(repFile3, []byte(level3Data), 0644)
	if err != nil {
		t.Fatalf("Failed to create level 3 file: %v", err)
	}

	client := NewReputationClient(tempDir)

	// LoadAllLevels should only load existing files
	err = client.LoadAllLevels()
	if err != nil {
		t.Fatalf("LoadAllLevels failed: %v", err)
	}

	// Should have loaded levels 1 and 3
	levels := client.GetLoadedLevels()
	expectedLevels := []int{1, 3}

	if len(levels) != len(expectedLevels) {
		t.Errorf("Expected %d levels, got %d", len(expectedLevels), len(levels))
	}

	for _, expectedLevel := range expectedLevels {
		found := false
		for _, loadedLevel := range levels {
			if loadedLevel == expectedLevel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected level %d to be loaded", expectedLevel)
		}
	}

	// Test that the loaded levels work
	if client.Lookup(1, "127.0.0.1") != 1 {
		t.Error("Level 1 lookup should work")
	}
	if client.Lookup(3, "192.168.1.1") != 1 {
		t.Error("Level 3 lookup should work")
	}
	if client.Lookup(2, "anything") != 0 {
		t.Error("Level 2 should not be loaded")
	}
}

// TestReputationClient_InvalidData tests handling of invalid data
func TestReputationClient_InvalidData(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create file with mixed valid and invalid data
	testData := `# Valid entries
127.0.0.1
192.168.1.0/24

# Invalid entries that should be skipped
not-an-ip
999.999.999.999
192.168.1.0/99
invalid/cidr
300.0.0.1

# More valid entries
10.0.0.0/8
203.0.113.42
`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)

	// Loading should succeed despite invalid entries
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Loading should succeed despite invalid entries: %v", err)
	}

	// Valid entries should still work
	if client.Lookup(1, "127.0.0.1") != 1 {
		t.Error("Valid IP should be found")
	}
	if client.Lookup(1, "192.168.1.50") != 1 {
		t.Error("IP in valid subnet should be found")
	}
	if client.Lookup(1, "10.5.5.5") != 1 {
		t.Error("IP in valid large subnet should be found")
	}
	if client.Lookup(1, "203.0.113.42") != 1 {
		t.Error("Valid IP should be found")
	}

	// Invalid entries should not affect lookups
	if client.Lookup(1, "999.999.999.999") != 0 {
		t.Error("Invalid IP should not be found")
	}
}

// TestReputationClient_EmptyAndMissingFiles tests handling of empty and missing files
func TestReputationClient_EmptyAndMissingFiles(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	client := NewReputationClient(tempDir)

	// Test loading missing file
	err = client.LoadLevel(1)
	if err == nil {
		t.Error("Loading missing file should return error")
	}

	// Create empty file
	emptyFile := filepath.Join(sourcesDir, "firehol_level2.netset")
	err = os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Loading empty file should succeed
	err = client.LoadLevel(2)
	if err != nil {
		t.Fatalf("Loading empty file should succeed: %v", err)
	}

	// Level should be loaded but no lookups should match
	levels := client.GetLoadedLevels()
	found := false
	for _, level := range levels {
		if level == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Level 2 should be loaded")
	}

	if client.Lookup(2, "127.0.0.1") != 0 {
		t.Error("Lookup in empty level should return 0")
	}

	// Create file with only comments and empty lines
	commentsOnlyData := `# This file only has comments
# And empty lines

# No actual IP data
`
	commentsFile := filepath.Join(sourcesDir, "firehol_level3.netset")
	err = os.WriteFile(commentsFile, []byte(commentsOnlyData), 0644)
	if err != nil {
		t.Fatalf("Failed to create comments-only file: %v", err)
	}

	// Loading comments-only file should succeed
	err = client.LoadLevel(3)
	if err != nil {
		t.Fatalf("Loading comments-only file should succeed: %v", err)
	}

	if client.Lookup(3, "127.0.0.1") != 0 {
		t.Error("Lookup in comments-only level should return 0")
	}
}

// TestReputationClient_IPv6Support tests IPv6 address handling
func TestReputationClient_IPv6Support(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data with IPv6 addresses
	testData := `2001:db8::1
2001:db8::/32
::1
fe80::/64`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Failed to load reputation level: %v", err)
	}

	testCases := []struct {
		ip       string
		expected int
		desc     string
	}{
		{"2001:db8::1", 1, "exact IPv6 match"},
		{"2001:db8::2", 1, "IPv6 in subnet"},
		{"2001:db8:1::1", 1, "IPv6 in larger subnet"},
		{"::1", 1, "localhost IPv6"},
		{"fe80::1", 1, "link-local IPv6"},
		{"2001:db9::1", 0, "IPv6 not in range"},
		{"127.0.0.1", 0, "IPv4 should not match IPv6"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := client.Lookup(1, tc.ip)
			if result != tc.expected {
				t.Errorf("Lookup(%s) = %d, expected %d (%s)", tc.ip, result, tc.expected, tc.desc)
			}
		})
	}
}

// TestReputationClient_LevelBoundaries tests level boundary conditions
func TestReputationClient_LevelBoundaries(t *testing.T) {
	tempDir := t.TempDir()
	client := NewReputationClient(tempDir)

	// Test invalid levels
	invalidLevels := []int{0, -1, 5, 10, 100}
	for _, level := range invalidLevels {
		result := client.Lookup(level, "127.0.0.1")
		if result != 0 {
			t.Errorf("Lookup with invalid level %d should return 0, got %d", level, result)
		}
	}

	// Test valid levels (even if not loaded)
	validLevels := []int{1, 2, 3, 4}
	for _, level := range validLevels {
		result := client.Lookup(level, "127.0.0.1")
		if result != 0 {
			t.Errorf("Lookup with unloaded valid level %d should return 0, got %d", level, result)
		}
	}
}

// TestReputationClient_ConcurrentAccess tests thread safety
func TestReputationClient_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Create test data
	testData := `127.0.0.1
192.168.1.0/24
10.0.0.0/8`

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	err = os.WriteFile(repFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test reputation file: %v", err)
	}

	client := NewReputationClient(tempDir)
	err = client.LoadLevel(1)
	if err != nil {
		t.Fatalf("Failed to load reputation level: %v", err)
	}

	// First, verify expected results in single-threaded mode
	expectedResults := map[string]int{
		"127.0.0.1":   1, // Individual IP
		"192.168.1.1": 1, // IP in subnet
		"10.5.5.5":    1, // IP in large subnet
		"8.8.8.8":     0, // Not in any range
		"192.168.2.1": 0, // Not in any range
	}

	for ip, expected := range expectedResults {
		result := client.Lookup(1, ip)
		if result != expected {
			t.Fatalf("Single-threaded lookup for %s returned %d, expected %d", ip, result, expected)
		}
	}

	// Test concurrent lookups - just verify no panics or deadlocks
	const numGoroutines = 50
	const numLookups = 20

	testIPs := []string{
		"127.0.0.1",
		"192.168.1.1",
		"10.5.5.5",
		"8.8.8.8",
		"192.168.2.1",
	}

	// Use sync.WaitGroup for better coordination
	var wg sync.WaitGroup
	var errorCount int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numLookups; j++ {
				for _, ip := range testIPs {
					result := client.Lookup(1, ip)
					expected := expectedResults[ip]
					if result != expected {
						atomic.AddInt32(&errorCount, 1)
						// Don't spam errors, just count them
						if atomic.LoadInt32(&errorCount) == 1 {
							t.Errorf("Concurrent lookup for %s returned %d, expected %d", ip, result, expected)
						}
					}
				}
			}
		}()
	}

	wg.Wait()

	// Check for errors
	if errorCount > 0 {
		t.Errorf("Found %d inconsistent results during concurrent testing", errorCount)
	}
}
