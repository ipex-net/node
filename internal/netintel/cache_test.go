package netintel

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestAtomicCache_BasicOperations tests basic cache functionality
func TestAtomicCache_BasicOperations(t *testing.T) {
	cache := NewAtomicCache()

	// Initially, all checkers should be nil
	if cache.GetReputationChecker() != nil {
		t.Error("Initial reputation checker should be nil")
	}
	if cache.GetGeolocationChecker() != nil {
		t.Error("Initial geolocation checker should be nil")
	}
	if cache.GetASNChecker() != nil {
		t.Error("Initial ASN checker should be nil")
	}

	// Cache should not be ready initially
	if cache.IsReady() {
		t.Error("Cache should not be ready initially")
	}
	if cache.IsReputationReady() {
		t.Error("Reputation should not be ready initially")
	}
	if cache.IsGeolocationReady() {
		t.Error("Geolocation should not be ready initially")
	}
	if cache.IsASNReady() {
		t.Error("ASN should not be ready initially")
	}

	// Initial last update should be zero time
	lastUpdate := cache.GetLastUpdate()
	if !lastUpdate.IsZero() {
		t.Error("Initial last update should be zero time")
	}
}

// TestAtomicCache_ReputationUpdates tests reputation checker updates
func TestAtomicCache_ReputationUpdates(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Create test reputation client
	repChecker := NewReputationClient(tempDir)

	// Update reputation checker
	cache.UpdateReputationChecker(repChecker)

	// Should now be available
	retrieved := cache.GetReputationChecker()
	if retrieved == nil {
		t.Fatal("Reputation checker should not be nil after update")
	}
	if retrieved != repChecker {
		t.Error("Retrieved reputation checker should be the same instance")
	}

	// Reputation should be ready, but cache overall should not (missing geolocation)
	if !cache.IsReputationReady() {
		t.Error("Reputation should be ready after update")
	}
	if cache.IsReady() {
		t.Error("Cache should not be ready with only reputation")
	}

	// Last update should be recent
	lastUpdate := cache.GetLastUpdate()
	if lastUpdate.IsZero() {
		t.Error("Last update should not be zero after update")
	}
	if time.Since(lastUpdate) > time.Second {
		t.Error("Last update should be very recent")
	}
}

// TestAtomicCache_GeolocationUpdates tests geolocation checker updates
func TestAtomicCache_GeolocationUpdates(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Create test geolocation client
	geoChecker := NewGeolocationClient(tempDir)

	// Update geolocation checker
	cache.UpdateGeolocationChecker(geoChecker)

	// Should now be available
	retrieved := cache.GetGeolocationChecker()
	if retrieved == nil {
		t.Fatal("Geolocation checker should not be nil after update")
	}
	if retrieved != geoChecker {
		t.Error("Retrieved geolocation checker should be the same instance")
	}

	// Geolocation should be ready
	if !cache.IsGeolocationReady() {
		t.Error("Geolocation should be ready after update")
	}
}

// TestAtomicCache_ASNUpdates tests ASN checker updates
func TestAtomicCache_ASNUpdates(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Create test ASN client
	asnChecker := NewASNClient(tempDir)

	// Update ASN checker
	cache.UpdateASNChecker(asnChecker)

	// Should now be available
	retrieved := cache.GetASNChecker()
	if retrieved == nil {
		t.Fatal("ASN checker should not be nil after update")
	}
	if retrieved != asnChecker {
		t.Error("Retrieved ASN checker should be the same instance")
	}

	// ASN should be ready
	if !cache.IsASNReady() {
		t.Error("ASN should be ready after update")
	}
}

// TestAtomicCache_ReadinessLogic tests the cache readiness logic
func TestAtomicCache_ReadinessLogic(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	repChecker := NewReputationClient(tempDir)
	geoChecker := NewGeolocationClient(tempDir)
	asnChecker := NewASNClient(tempDir)

	// Cache should only be ready when geolocation is set
	// Reputation and ASN are optional and don't affect readiness

	// Add reputation only - should not be ready (no geolocation)
	cache.UpdateReputationChecker(repChecker)
	if cache.IsReady() {
		t.Error("Cache should not be ready with only reputation")
	}

	// Add geolocation - now should be ready
	cache.UpdateGeolocationChecker(geoChecker)
	if !cache.IsReady() {
		t.Error("Cache should be ready with geolocation")
	}

	// Add ASN - should still be ready
	cache.UpdateASNChecker(asnChecker)
	if !cache.IsReady() {
		t.Error("Cache should still be ready with all checkers")
	}

	// Remove reputation - should still be ready (geolocation is what matters)
	cache.UpdateReputationChecker(nil)
	if !cache.IsReady() {
		t.Error("Cache should still be ready with geolocation (reputation is optional)")
	}

	// Remove geolocation - should not be ready
	cache.UpdateGeolocationChecker(nil)
	if cache.IsReady() {
		t.Error("Cache should not be ready without geolocation")
	}
}

// TestAtomicCache_ConvenienceMethods tests the convenience methods
func TestAtomicCache_ConvenienceMethods(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Test methods with nil checkers (should return safe defaults)
	repResult := cache.CheckReputation(1, "127.0.0.1")
	if repResult != 0 {
		t.Error("Reputation check should return 0 with nil checker")
	}

	countryResult := cache.CheckCountry("127.0.0.1", "US")
	if countryResult {
		t.Error("Country check should return false with nil checker")
	}

	asnResult := cache.CheckASNDrop(64512)
	if asnResult {
		t.Error("ASN check should return false with nil checker")
	}

	// Create test data for ASN checker
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	asnData := `{"asn":64512,"rir":"iana","domain":"test","cc":"US","asname":"TEST-ASN"}`
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	err = os.WriteFile(asnFile, []byte(asnData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ASN file: %v", err)
	}

	// Set up ASN checker with test data
	asnChecker := NewASNClient(tempDir)
	err = asnChecker.LoadASNDropList()
	if err != nil {
		t.Fatalf("Failed to load ASN data: %v", err)
	}

	cache.UpdateASNChecker(asnChecker)

	// Now ASN check should work
	asnResult = cache.CheckASNDrop(64512)
	if !asnResult {
		t.Error("ASN check should return true for loaded ASN")
	}

	asnResult = cache.CheckASNDrop(99999)
	if asnResult {
		t.Error("ASN check should return false for non-existent ASN")
	}
}

// TestAtomicCache_Status tests the cache status reporting
func TestAtomicCache_Status(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Initial status
	status := cache.GetStatus()
	if status.ReputationReady || status.GeolocationReady || status.ASNReady {
		t.Error("Initial status should show all checkers as not ready")
	}
	if len(status.LoadedLevels) != 0 {
		t.Error("Initial loaded levels should be empty")
	}
	if len(status.LoadedCountries) != 0 {
		t.Error("Initial loaded countries should be empty")
	}
	if status.LoadedASNCount != 0 {
		t.Error("Initial loaded ASN count should be zero")
	}

	// Add checkers and verify status updates
	repChecker := NewReputationClient(tempDir)
	cache.UpdateReputationChecker(repChecker)

	status = cache.GetStatus()
	if !status.ReputationReady {
		t.Error("Status should show reputation as ready")
	}
	if status.GeolocationReady || status.ASNReady {
		t.Error("Status should show geo and ASN as not ready")
	}
}

// TestAtomicCache_Clear tests the cache clear functionality
func TestAtomicCache_Clear(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Set up all checkers
	repChecker := NewReputationClient(tempDir)
	geoChecker := NewGeolocationClient(tempDir)
	asnChecker := NewASNClient(tempDir)

	cache.UpdateReputationChecker(repChecker)
	cache.UpdateGeolocationChecker(geoChecker)
	cache.UpdateASNChecker(asnChecker)

	// Verify all are set
	if !cache.IsReputationReady() || !cache.IsGeolocationReady() || !cache.IsASNReady() {
		t.Error("All checkers should be ready before clear")
	}

	// Clear cache
	cache.Clear()

	// Verify all are cleared
	if cache.IsReputationReady() || cache.IsGeolocationReady() || cache.IsASNReady() {
		t.Error("No checkers should be ready after clear")
	}
	if cache.GetReputationChecker() != nil {
		t.Error("Reputation checker should be nil after clear")
	}
	if cache.GetGeolocationChecker() != nil {
		t.Error("Geolocation checker should be nil after clear")
	}
	if cache.GetASNChecker() != nil {
		t.Error("ASN checker should be nil after clear")
	}
}

// TestAtomicCache_ConcurrentAccess tests thread safety
func TestAtomicCache_ConcurrentAccess(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	var wg sync.WaitGroup
	const numGoroutines = 100
	const numOperations = 10

	// Create test clients
	repChecker1 := NewReputationClient(tempDir)
	repChecker2 := NewReputationClient(tempDir)
	geoChecker1 := NewGeolocationClient(tempDir)
	geoChecker2 := NewGeolocationClient(tempDir)
	asnChecker1 := NewASNClient(tempDir)
	asnChecker2 := NewASNClient(tempDir)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				switch (id + j) % 6 {
				case 0:
					cache.UpdateReputationChecker(repChecker1)
				case 1:
					cache.UpdateReputationChecker(repChecker2)
				case 2:
					cache.UpdateGeolocationChecker(geoChecker1)
				case 3:
					cache.UpdateGeolocationChecker(geoChecker2)
				case 4:
					cache.UpdateASNChecker(asnChecker1)
				case 5:
					cache.UpdateASNChecker(asnChecker2)
				}

				// Concurrent reads
				_ = cache.GetReputationChecker()
				_ = cache.GetGeolocationChecker()
				_ = cache.GetASNChecker()
				_ = cache.IsReady()
				_ = cache.GetStatus()
				_ = cache.CheckReputation(1, "127.0.0.1")
				_ = cache.CheckCountry("127.0.0.1", "US")
				_ = cache.CheckASNDrop(64512)
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional after concurrent access
	if cache.GetReputationChecker() == nil {
		t.Error("Reputation checker should not be nil after concurrent access")
	}
	if cache.GetGeolocationChecker() == nil {
		t.Error("Geolocation checker should not be nil after concurrent access")
	}
	if cache.GetASNChecker() == nil {
		t.Error("ASN checker should not be nil after concurrent access")
	}
}

// TestAtomicCache_HotSwap tests hot-swapping during active use
func TestAtomicCache_HotSwap(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	// Set up initial ASN checker with test data
	sourcesDir := filepath.Join(tempDir, "sources")
	err := os.MkdirAll(sourcesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sources directory: %v", err)
	}

	// Initial data
	asnData1 := `{"asn":64512,"rir":"iana","domain":"test1","cc":"US","asname":"TEST-ASN-1"}`
	asnFile := filepath.Join(sourcesDir, "asndrop.json")
	err = os.WriteFile(asnFile, []byte(asnData1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ASN file: %v", err)
	}

	asnChecker1 := NewASNClient(tempDir)
	err = asnChecker1.LoadASNDropList()
	if err != nil {
		t.Fatalf("Failed to load initial ASN data: %v", err)
	}

	cache.UpdateASNChecker(asnChecker1)

	// Verify initial state
	if !cache.CheckASNDrop(64512) {
		t.Error("Initial ASN should be found")
	}

	// Create updated data
	asnData2 := `{"asn":64513,"rir":"iana","domain":"test2","cc":"GB","asname":"TEST-ASN-2"}`
	err = os.WriteFile(asnFile, []byte(asnData2), 0644)
	if err != nil {
		t.Fatalf("Failed to update test ASN file: %v", err)
	}

	asnChecker2 := NewASNClient(tempDir)
	err = asnChecker2.LoadASNDropList()
	if err != nil {
		t.Fatalf("Failed to load updated ASN data: %v", err)
	}

	// Perform hot swap while doing concurrent reads
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					// These should never panic during hot swap
					_ = cache.CheckASNDrop(64512)
					_ = cache.CheckASNDrop(64513)
					_ = cache.GetASNChecker()
					_ = cache.IsASNReady()
				}
			}
		}()
	}

	// Perform the hot swap
	time.Sleep(10 * time.Millisecond) // Let readers start
	cache.UpdateASNChecker(asnChecker2)
	time.Sleep(10 * time.Millisecond) // Let readers see new data

	// Stop readers
	close(done)
	wg.Wait()

	// Verify the swap worked
	if cache.CheckASNDrop(64512) {
		t.Error("Old ASN should not be found after hot swap")
	}
	if !cache.CheckASNDrop(64513) {
		t.Error("New ASN should be found after hot swap")
	}
}

// TestAtomicCache_TimestampUpdates tests that timestamps update correctly
func TestAtomicCache_TimestampUpdates(t *testing.T) {
	cache := NewAtomicCache()
	tempDir := t.TempDir()

	initialTime := cache.GetLastUpdate()
	if !initialTime.IsZero() {
		t.Error("Initial timestamp should be zero")
	}

	// First update
	repChecker := NewReputationClient(tempDir)
	cache.UpdateReputationChecker(repChecker)

	firstUpdate := cache.GetLastUpdate()
	if firstUpdate.IsZero() {
		t.Error("Timestamp should be set after first update")
	}

	// Wait for next second to ensure timestamp difference (cache uses Unix timestamps)
	time.Sleep(1100 * time.Millisecond)

	// Second update
	geoChecker := NewGeolocationClient(tempDir)
	cache.UpdateGeolocationChecker(geoChecker)

	secondUpdate := cache.GetLastUpdate()
	if !secondUpdate.After(firstUpdate) {
		t.Errorf("Second update timestamp (%v) should be after first (%v)", secondUpdate, firstUpdate)
	}

	// Clear should also update timestamp
	time.Sleep(1100 * time.Millisecond)
	cache.Clear()

	clearUpdate := cache.GetLastUpdate()
	if !clearUpdate.After(secondUpdate) {
		t.Errorf("Clear timestamp (%v) should be after second update (%v)", clearUpdate, secondUpdate)
	}
}
