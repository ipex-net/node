//go:build perf

package netintel

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
)

// TestPerformance_ConcurrentLookups tests performance under high concurrent load
func TestPerformance_ConcurrentLookups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create comprehensive test data
	sourcesDir := filepath.Join(tempDir, "sources")

	// Large reputation dataset
	repData := "# Large reputation dataset\n"
	for i := 1; i <= 1000; i++ {
		repData += fmt.Sprintf("192.168.%d.%d\n", (i%254)+1, (i%254)+1)
	}
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte(repData), 0644)

	// Large geolocation dataset
	geoData := "# Large geolocation dataset\n"
	for i := 1; i <= 1000; i++ {
		geoData += fmt.Sprintf("10.%d.%d.0/24\n", (i%254)+1, (i%254)+1)
	}
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	service := NewServiceForTesting(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness
	timeout := time.After(10 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Performance test parameters
	const numGoroutines = 100
	const lookupsPerGoroutine = 1000
	const totalLookups = numGoroutines * lookupsPerGoroutine

	// Test IPs for lookups
	testIPs := []string{
		"192.168.1.1", "192.168.50.50", "192.168.100.100",
		"10.1.1.1", "10.50.50.1", "10.100.100.1",
		"203.0.113.1", "198.51.100.1", "8.8.8.8",
	}

	// Track performance metrics
	var wg sync.WaitGroup
	startTime := time.Now()

	errors := make(chan error, numGoroutines)
	latencies := make(chan time.Duration, totalLookups)

	// Run concurrent reputation lookups
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < lookupsPerGoroutine; j++ {
				lookupStart := time.Now()

				// Test reputation lookup
				testIP := testIPs[j%len(testIPs)]
				result := service.CheckReputation(testIP+"/32", 1)

				// Test country lookup
				service.CheckCountry(testIP, "US")

				// Test service status
				if j%100 == 0 {
					service.IsReady()
				}

				lookupDuration := time.Since(lookupStart)
				latencies <- lookupDuration

				// Validate that operations complete without errors
				if result < 0 {
					errors <- fmt.Errorf("invalid reputation result: %d", result)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)
	close(errors)
	close(latencies)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent lookup error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Found %d errors during concurrent testing", errorCount)
	}

	// Calculate performance metrics
	var totalLatency time.Duration
	var maxLatency time.Duration
	var minLatency time.Duration = time.Hour // Start with high value
	latencyCount := 0

	for latency := range latencies {
		totalLatency += latency
		latencyCount++
		if latency > maxLatency {
			maxLatency = latency
		}
		if latency < minLatency {
			minLatency = latency
		}
	}

	avgLatency := totalLatency / time.Duration(latencyCount)
	throughput := float64(totalLookups) / totalDuration.Seconds()

	// Performance assertions
	t.Logf("Performance Results:")
	t.Logf("  Total operations: %d", totalLookups)
	t.Logf("  Total duration: %v", totalDuration)
	t.Logf("  Throughput: %.2f ops/sec", throughput)
	t.Logf("  Average latency: %v", avgLatency)
	t.Logf("  Min latency: %v", minLatency)
	t.Logf("  Max latency: %v", maxLatency)

	// Performance expectations (adjust based on requirements)
	if throughput < 1000 {
		t.Errorf("Throughput too low: %.2f ops/sec (expected > 1000)", throughput)
	}

	if avgLatency > 10*time.Millisecond {
		t.Errorf("Average latency too high: %v (expected < 10ms)", avgLatency)
	}

	if maxLatency > 100*time.Millisecond {
		t.Errorf("Max latency too high: %v (expected < 100ms)", maxLatency)
	}

	// Service should still be ready
	if !service.IsReady() {
		t.Error("Service should remain ready after performance test")
	}
}

// TestPerformance_MemoryUsage tests memory usage under load
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create large datasets to test memory efficiency
	sourcesDir := filepath.Join(tempDir, "sources")

	// Large reputation data
	repSize := 10000
	repData := "# Large reputation dataset\n"
	for i := 1; i <= repSize; i++ {
		repData += fmt.Sprintf("192.168.%d.%d\n", (i%254)+1, (i%254)+1)
	}

	for level := 1; level <= 4; level++ {
		repFile := filepath.Join(sourcesDir, fmt.Sprintf("firehol_level%d.netset", level))
		os.MkdirAll(filepath.Dir(repFile), 0755)
		os.WriteFile(repFile, []byte(repData), 0644)
	}

	// Large geolocation data
	geoSize := 5000
	geoData := "# Large geolocation dataset\n"
	for i := 1; i <= geoSize; i++ {
		geoData += fmt.Sprintf("10.%d.%d.0/24\n", (i%254)+1, (i%254)+1)
	}

	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	// Measure memory before service start
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	service := NewServiceForTesting(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for full data loading
	timeout := time.After(15 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready")
		case <-time.After(100 * time.Millisecond):
		}
	}

	// Measure memory after service start and data loading
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Perform operations to test memory stability
	const numOperations = 10000
	for i := 0; i < numOperations; i++ {
		testIP := fmt.Sprintf("192.168.%d.%d", (i%254)+1, (i%254)+1)
		service.CheckReputation(testIP+"/32", 1)
		service.CheckCountry(testIP, "US")

		if i%1000 == 0 {
			service.GetLoadedLevels()
			service.GetLoadedCountries()
		}
	}

	// Measure memory after operations
	runtime.GC()
	var memAfterOps runtime.MemStats
	runtime.ReadMemStats(&memAfterOps)

	// Calculate memory usage
	memoryUsed := memAfter.Alloc - memBefore.Alloc

	// Calculate memory growth during operations (handle underflow)
	var memoryGrowth uint64
	if memAfterOps.Alloc > memAfter.Alloc {
		memoryGrowth = memAfterOps.Alloc - memAfter.Alloc
	} else {
		memoryGrowth = 0 // Memory decreased or stayed same
	}

	t.Logf("Memory Usage Results:")
	t.Logf("  Memory before: %d KB", memBefore.Alloc/1024)
	t.Logf("  Memory after loading: %d KB", memAfter.Alloc/1024)
	t.Logf("  Memory after operations: %d KB", memAfterOps.Alloc/1024)
	t.Logf("  Memory used by service: %d KB", memoryUsed/1024)
	t.Logf("  Memory growth during operations: %d KB", memoryGrowth/1024)

	// Memory usage expectations (adjust based on requirements)
	maxExpectedMemory := uint64(100 * 1024 * 1024) // 100MB
	if memoryUsed > maxExpectedMemory {
		t.Errorf("Memory usage too high: %d KB (expected < %d KB)",
			memoryUsed/1024, maxExpectedMemory/1024)
	}

	// Memory should not grow significantly during operations
	maxGrowth := uint64(10 * 1024 * 1024) // 10MB
	if memoryGrowth > maxGrowth {
		t.Errorf("Memory growth during operations too high: %d KB (expected < %d KB)",
			memoryGrowth/1024, maxGrowth/1024)
	}

	// Service should still be functional
	if !service.IsReady() {
		t.Error("Service should remain ready after memory test")
	}
}

// TestPerformance_AtomicHotSwap tests performance of atomic hot-swapping operations
func TestPerformance_AtomicHotSwap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create test data
	sourcesDir := filepath.Join(tempDir, "sources")

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte("127.0.0.1\n192.168.1.0/24\n"), 0644)

	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte("192.168.1.0/24\n"), 0644)

	service := NewServiceForTesting(tempDir, logger)

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

	// Test concurrent lookups during atomic swaps
	const numLookupGoroutines = 50
	const numSwapGoroutines = 5
	const testDuration = 5 * time.Second

	var wg sync.WaitGroup
	stopTime := time.Now().Add(testDuration)

	lookupCount := make(chan int, numLookupGoroutines)
	swapCount := make(chan int, numSwapGoroutines)
	errors := make(chan error, numLookupGoroutines+numSwapGoroutines)

	// Start continuous lookup goroutines
	for i := 0; i < numLookupGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for time.Now().Before(stopTime) {
				result := service.CheckReputation("127.0.0.1/32", 1)
				service.CheckCountry("192.168.1.1", "US")
				count++

				// Validate consistency
				if result < 0 {
					errors <- fmt.Errorf("invalid result during hot swap: %d", result)
					return
				}
			}
			lookupCount <- count
		}()
	}

	// Start atomic swap goroutines
	for i := 0; i < numSwapGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			for time.Now().Before(stopTime) {
				// Create new clients and swap them in
				newRepClient := NewReputationClient(tempDir)
				newRepClient.LoadLevel(1)
				service.cache.UpdateReputationChecker(newRepClient)

				newGeoClient := NewGeolocationClient(tempDir)
				newGeoClient.LoadCountry("US")
				service.cache.UpdateGeolocationChecker(newGeoClient)

				count++
				time.Sleep(100 * time.Millisecond) // Throttle swaps
			}
			swapCount <- count
		}()
	}

	wg.Wait()
	close(lookupCount)
	close(swapCount)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Hot swap error: %v", err)
		errorCount++
	}

	// Calculate metrics
	totalLookups := 0
	for count := range lookupCount {
		totalLookups += count
	}

	totalSwaps := 0
	for count := range swapCount {
		totalSwaps += count
	}

	lookupThroughput := float64(totalLookups) / testDuration.Seconds()
	swapThroughput := float64(totalSwaps) / testDuration.Seconds()

	t.Logf("Hot Swap Performance Results:")
	t.Logf("  Test duration: %v", testDuration)
	t.Logf("  Total lookups: %d", totalLookups)
	t.Logf("  Total swaps: %d", totalSwaps)
	t.Logf("  Lookup throughput: %.2f ops/sec", lookupThroughput)
	t.Logf("  Swap throughput: %.2f swaps/sec", swapThroughput)
	t.Logf("  Errors: %d", errorCount)

	// Performance expectations
	if errorCount > 0 {
		t.Errorf("Hot swap should not cause errors: %d errors found", errorCount)
	}

	if lookupThroughput < 1000 {
		t.Errorf("Lookup throughput during hot swaps too low: %.2f ops/sec", lookupThroughput)
	}

	if totalSwaps == 0 {
		t.Error("Should have completed some atomic swaps")
	}

	// Service should still be ready and functional
	if !service.IsReady() {
		t.Error("Service should remain ready after hot swap test")
	}

	if service.CheckReputation("127.0.0.1/32", 1) == 0 {
		t.Error("Service should still work after hot swaps")
	}
}

// TestPerformance_LargeDatasets tests performance with large datasets
func TestPerformance_LargeDatasets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create very large datasets
	sourcesDir := filepath.Join(tempDir, "sources")

	// Very large reputation dataset
	t.Log("Creating large reputation dataset...")
	startTime := time.Now()

	repSize := 50000
	repData := "# Very large reputation dataset\n"
	for i := 1; i <= repSize; i++ {
		// Create diverse IP ranges
		if i%4 == 0 {
			repData += fmt.Sprintf("192.168.%d.%d\n", (i%254)+1, (i%254)+1)
		} else {
			repData += fmt.Sprintf("10.%d.%d.0/24\n", (i%254)+1, (i%254)+1)
		}
	}

	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	os.WriteFile(repFile, []byte(repData), 0644)

	// Create geolocation data to make service ready
	geoData := "# Geolocation test data\n192.168.1.0/24\n"
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	dataCreationTime := time.Since(startTime)
	t.Logf("Dataset creation time: %v", dataCreationTime)

	// Start service and measure loading time
	t.Log("Starting service with large dataset...")
	serviceStartTime := time.Now()

	service := NewServiceForTesting(tempDir, logger)

	err := service.Start()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
	defer service.Stop()

	// Wait for readiness with longer timeout for large datasets
	timeout := time.After(30 * time.Second)
	for !service.IsReady() {
		select {
		case <-timeout:
			t.Fatal("Service didn't become ready with large dataset")
		case <-time.After(500 * time.Millisecond):
		}
	}

	serviceLoadTime := time.Since(serviceStartTime)
	t.Logf("Service load time: %v", serviceLoadTime)

	// Test lookup performance with large dataset
	t.Log("Testing lookup performance...")
	lookupStartTime := time.Now()

	const numLookups = 10000
	hits := 0
	misses := 0

	for i := 0; i < numLookups; i++ {
		testIP := fmt.Sprintf("192.168.%d.%d", (i%254)+1, (i%254)+1)
		result := service.CheckReputation(testIP+"/32", 1)
		if result > 0 {
			hits++
		} else {
			misses++
		}
	}

	lookupTime := time.Since(lookupStartTime)
	lookupThroughput := float64(numLookups) / lookupTime.Seconds()

	t.Logf("Large Dataset Performance Results:")
	t.Logf("  Dataset size: %d entries", repSize)
	t.Logf("  Service load time: %v", serviceLoadTime)
	t.Logf("  Lookup operations: %d", numLookups)
	t.Logf("  Lookup time: %v", lookupTime)
	t.Logf("  Lookup throughput: %.2f ops/sec", lookupThroughput)
	t.Logf("  Hit rate: %.2f%% (%d hits, %d misses)", float64(hits)/float64(numLookups)*100, hits, misses)

	// Performance expectations for large datasets
	maxLoadTime := 20 * time.Second
	if serviceLoadTime > maxLoadTime {
		t.Errorf("Service load time too high: %v (expected < %v)", serviceLoadTime, maxLoadTime)
	}

	minThroughput := 500.0 // ops/sec
	if lookupThroughput < minThroughput {
		t.Errorf("Lookup throughput too low: %.2f ops/sec (expected > %.2f)", lookupThroughput, minThroughput)
	}

	// Test memory usage with large dataset
	runtime.GC()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	memoryUsed := memStats.Alloc / 1024 / 1024 // MB
	t.Logf("Memory usage with large dataset: %d MB", memoryUsed)

	maxMemory := uint64(200) // MB
	if memoryUsed > maxMemory {
		t.Errorf("Memory usage too high: %d MB (expected < %d MB)", memoryUsed, maxMemory)
	}

	// Service should still be functional
	if !service.IsReady() {
		t.Error("Service should remain ready with large dataset")
	}
}

// TestPerformance_CacheOperations tests cache operation performance
func TestPerformance_CacheOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cache := NewAtomicCache()

	// Create test clients
	tempDir := t.TempDir()
	sourcesDir := filepath.Join(tempDir, "sources")

	// Reputation client
	repFile := filepath.Join(sourcesDir, "firehol_level1.netset")
	os.MkdirAll(filepath.Dir(repFile), 0755)
	repData := ""
	for i := 1; i <= 1000; i++ {
		repData += fmt.Sprintf("192.168.%d.%d\n", (i%254)+1, (i%254)+1)
	}
	os.WriteFile(repFile, []byte(repData), 0644)

	repClient := NewReputationClient(tempDir)
	repClient.LoadLevel(1)

	// Geolocation client
	geoDir := filepath.Join(sourcesDir, "geolite2_country")
	os.MkdirAll(geoDir, 0755)
	geoData := ""
	for i := 1; i <= 1000; i++ {
		geoData += fmt.Sprintf("10.%d.%d.0/24\n", (i%254)+1, (i%254)+1)
	}
	geoFile := filepath.Join(geoDir, "country_us.netset")
	os.WriteFile(geoFile, []byte(geoData), 0644)

	geoClient := NewGeolocationClient(tempDir)
	geoClient.LoadCountry("US")

	// Test atomic update performance
	const numUpdates = 1000
	const numConcurrentReaders = 50

	var wg sync.WaitGroup
	updateTimes := make(chan time.Duration, numUpdates)
	errors := make(chan error, numConcurrentReaders*numUpdates)

	// Start concurrent readers
	stopReading := make(chan bool)
	for i := 0; i < numConcurrentReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopReading:
					return
				default:
					// Continuous reading operations
					cache.CheckReputation(1, "192.168.1.1/32")
					cache.CheckCountry("10.1.1.1", "US")
					cache.IsReady()
				}
			}
		}()
	}

	// Perform atomic updates while readers are running
	start := time.Now()
	for i := 0; i < numUpdates; i++ {
		updateStart := time.Now()

		// Alternate between reputation and geolocation updates
		if i%2 == 0 {
			cache.UpdateReputationChecker(repClient)
		} else {
			cache.UpdateGeolocationChecker(geoClient)
		}

		updateDuration := time.Since(updateStart)
		updateTimes <- updateDuration
	}
	totalUpdateTime := time.Since(start)

	// Stop readers
	close(stopReading)
	wg.Wait()
	close(updateTimes)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Cache operation error: %v", err)
		errorCount++
	}

	// Calculate metrics
	var totalUpdateDuration time.Duration
	var maxUpdateTime time.Duration
	var minUpdateTime time.Duration = time.Hour

	for updateTime := range updateTimes {
		totalUpdateDuration += updateTime
		if updateTime > maxUpdateTime {
			maxUpdateTime = updateTime
		}
		if updateTime < minUpdateTime {
			minUpdateTime = updateTime
		}
	}

	avgUpdateTime := totalUpdateDuration / numUpdates
	updateThroughput := float64(numUpdates) / totalUpdateTime.Seconds()

	t.Logf("Cache Performance Results:")
	t.Logf("  Total updates: %d", numUpdates)
	t.Logf("  Concurrent readers: %d", numConcurrentReaders)
	t.Logf("  Total update time: %v", totalUpdateTime)
	t.Logf("  Average update time: %v", avgUpdateTime)
	t.Logf("  Min update time: %v", minUpdateTime)
	t.Logf("  Max update time: %v", maxUpdateTime)
	t.Logf("  Update throughput: %.2f updates/sec", updateThroughput)
	t.Logf("  Errors: %d", errorCount)

	// Performance expectations
	if errorCount > 0 {
		t.Errorf("Cache operations should not cause errors: %d errors", errorCount)
	}

	maxAvgUpdateTime := 1 * time.Millisecond
	if avgUpdateTime > maxAvgUpdateTime {
		t.Errorf("Average update time too high: %v (expected < %v)", avgUpdateTime, maxAvgUpdateTime)
	}

	maxUpdateTimeThreshold := 10 * time.Millisecond
	if maxUpdateTime > maxUpdateTimeThreshold {
		t.Errorf("Max update time too high: %v (expected < %v)", maxUpdateTime, maxUpdateTimeThreshold)
	}

	// Cache should still be functional
	if !cache.IsReady() {
		t.Error("Cache should be ready after performance test")
	}
}
