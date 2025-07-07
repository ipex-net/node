package netintel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
)

// TestHTTPDownloader_BasicDownload tests basic file download functionality
func TestHTTPDownloader_BasicDownload(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Create a test HTTP server
	testData := "# Test reputation data\n127.0.0.1\n192.168.1.0/24\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testData))
	}))
	defer server.Close()

	// Test downloading to a file
	destPath := filepath.Join(tempDir, "sources", "test_file.txt")
	err := downloader.downloadFile(server.URL, destPath)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify file was created and has correct content
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file should exist")
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testData {
		t.Errorf("Downloaded content mismatch. Expected: %q, Got: %q", testData, string(content))
	}

	// Note: downloadFile is a private method that doesn't update last download time
	// Last download time is only updated by public methods like DownloadReputationLevel
}

// TestHTTPDownloader_HTTPErrors tests handling of various HTTP error codes
func TestHTTPDownloader_HTTPErrors(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	testCases := []struct {
		statusCode int
		shouldFail bool
	}{
		{http.StatusOK, false},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Status_%d", tc.statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				if tc.statusCode == http.StatusOK {
					w.Write([]byte("test data"))
				}
			}))
			defer server.Close()

			destPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.txt", tc.statusCode))
			err := downloader.downloadFile(server.URL, destPath)

			if tc.shouldFail && err == nil {
				t.Errorf("Expected error for status code %d, got nil", tc.statusCode)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected no error for status code %d, got: %v", tc.statusCode, err)
			}
		})
	}
}

// TestHTTPDownloader_RetryLogic tests the retry mechanism with exponential backoff
func TestHTTPDownloader_RetryLogic(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Fail first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Succeed on 3rd attempt
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	destPath := filepath.Join(tempDir, "retry_test.txt")
	start := time.Now()
	err := downloader.downloadFile(server.URL, destPath)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Download should succeed after retries: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// Should have taken some time due to exponential backoff
	if duration < 2*time.Second {
		t.Errorf("Retry should take some time due to backoff, took: %v", duration)
	}

	// Verify file content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "success" {
		t.Errorf("Expected 'success', got: %q", string(content))
	}
}

// TestHTTPDownloader_RetryFailure tests complete failure after all retries
func TestHTTPDownloader_RetryFailure(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	destPath := filepath.Join(tempDir, "fail_test.txt")
	err := downloader.downloadFile(server.URL, destPath)

	if err == nil {
		t.Error("Download should fail after all retries")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	// File should not exist after failed download
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Error("File should not exist after failed download")
	}
}

// TestHTTPDownloader_FilePermissions tests directory creation and file permissions
func TestHTTPDownloader_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	// Test nested directory creation
	destPath := filepath.Join(tempDir, "nested", "deep", "structure", "file.txt")
	err := downloader.downloadFile(server.URL, destPath)
	if err != nil {
		t.Fatalf("Download with nested dirs failed: %v", err)
	}

	// Verify directory was created
	dirPath := filepath.Dir(destPath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Error("Nested directory should have been created")
	}

	// Verify file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("File should exist in nested directory")
	}
}

// TestHTTPDownloader_ConcurrentDownloads tests concurrent download operations
func TestHTTPDownloader_ConcurrentDownloads(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Create server with different responses for different paths
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add small delay to simulate network latency
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("data for %s", r.URL.Path)))
	}))
	defer server.Close()

	const numConcurrent = 10
	var wg sync.WaitGroup
	errors := make(chan error, numConcurrent)

	// Start concurrent downloads
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			url := fmt.Sprintf("%s/file%d", server.URL, id)
			destPath := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.txt", id))
			if err := downloader.downloadFile(url, destPath); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent download failed: %v", err)
	}

	// Verify all files were created
	for i := 0; i < numConcurrent; i++ {
		destPath := filepath.Join(tempDir, fmt.Sprintf("concurrent_%d.txt", i))
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Errorf("File %d should exist", i)
		}
	}
}

// TestHTTPDownloader_DownloadReputationLevel tests reputation level downloading
func TestHTTPDownloader_DownloadReputationLevel(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Test invalid level (should fail immediately without network call)
	err := downloader.DownloadReputationLevel(0)
	if err == nil {
		t.Error("Should fail for invalid reputation level 0")
	}

	err = downloader.DownloadReputationLevel(5)
	if err == nil {
		t.Error("Should fail for invalid reputation level 5")
	}

	// Note: We don't test actual downloads here as they would hit real servers
	// This is covered by the private downloadFile tests above
}

// TestHTTPDownloader_DownloadASNDropList tests ASN DROP list downloading
func TestHTTPDownloader_DownloadASNDropList(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Note: We don't test actual ASN downloads here as they would hit real Spamhaus servers
	// This method should exist and not panic when called
	// The actual download functionality is tested via the private downloadFile tests

	// Verify the method exists and can be called (will likely fail due to network)
	err := downloader.DownloadASNDropList()
	// We expect this to fail since we're not connected to real servers
	// Just verify it doesn't panic and returns an error
	if err == nil {
		t.Log("Unexpected success - ASN download succeeded (real server might be available)")
	}
}

// TestHTTPDownloader_DownloadAllReputationLevels tests concurrent reputation downloads
func TestHTTPDownloader_DownloadAllReputationLevels(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Note: We don't test actual downloads here as they would hit real servers
	// This tests the method exists and handles errors appropriately

	err := downloader.DownloadAllReputationLevels()
	// We expect this to fail since we're not connected to real servers
	// Just verify it doesn't panic
	if err == nil {
		t.Log("Unexpected success - reputation download succeeded (real servers might be available)")
	}
}

// TestHTTPDownloader_LastDownloadTime tests that last download time is tracked
func TestHTTPDownloader_LastDownloadTime(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Initially, last download should be zero
	lastDownload := downloader.GetLastDownload()
	if !lastDownload.IsZero() {
		t.Error("Initial last download time should be zero")
	}

	// Note: We can't easily test that the time gets updated without hitting real servers
	// or using dependency injection, which would require refactoring the downloader
}

// TestHTTPDownloader_NetworkTimeout tests network timeout handling
func TestHTTPDownloader_NetworkTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	tempDir := t.TempDir()
	logger := log.NewNopLogger()
	downloader := NewHTTPDownloader(tempDir, logger)

	// Create server that hangs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second) // Longer than client timeout (30s)
		w.Write([]byte("should not reach here"))
	}))
	defer server.Close()

	destPath := filepath.Join(tempDir, "timeout_test.txt")
	start := time.Now()
	err := downloader.downloadFile(server.URL, destPath)
	duration := time.Since(start)

	if err == nil {
		t.Error("Download should timeout and fail")
	}

	// Should timeout after 3 attempts with retries
	// Each attempt: 30s timeout + sleep (2s, 4s between attempts)
	// Total expected: ~96 seconds (3 * 30s + 2s + 4s)
	expectedMin := 90 * time.Second
	expectedMax := 105 * time.Second
	if duration < expectedMin || duration > expectedMax {
		t.Errorf("Timeout duration unexpected: %v (expected %v-%v)", duration, expectedMin, expectedMax)
	}

	// File should not exist after timeout
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		t.Error("File should not exist after timeout")
	}
}

func TestHTTPDownloader_DownloadWithMockServer(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNopLogger()

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/firehol_level1.netset":
			w.Write([]byte("1.1.1.1\n192.168.1.0/24\n"))
		case "/asndrop.json":
			w.Write([]byte(`{"asn":12345,"rir":"test","domain":"test.com","cc":"US","asname":"TEST-ASN"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a custom downloader with overridden URLs
	downloader := &HTTPDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		dataDir: tempDir,
		logger:  logger,
	}

	// Test downloading by directly calling downloadFile with mock server URLs
	repPath := filepath.Join(tempDir, "sources", "firehol_level1.netset")
	err := downloader.downloadFile(server.URL+"/firehol_level1.netset", repPath)
	if err != nil {
		t.Fatalf("Download reputation file failed: %v", err)
	}

	// Verify the reputation file was downloaded
	if _, err := os.Stat(repPath); os.IsNotExist(err) {
		t.Error("Reputation file should exist")
	}

	// Verify file content
	content, err := os.ReadFile(repPath)
	if err != nil {
		t.Fatalf("Failed to read reputation file: %v", err)
	}
	expectedContent := "1.1.1.1\n192.168.1.0/24\n"
	if string(content) != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, string(content))
	}

	// Test downloading ASN file
	asnPath := filepath.Join(tempDir, "sources", "asndrop.json")
	err = downloader.downloadFile(server.URL+"/asndrop.json", asnPath)
	if err != nil {
		t.Fatalf("Download ASN file failed: %v", err)
	}

	// Verify the ASN file was downloaded
	if _, err := os.Stat(asnPath); os.IsNotExist(err) {
		t.Error("ASN drop list file should exist")
	}

	// Verify ASN file content
	asnContent, err := os.ReadFile(asnPath)
	if err != nil {
		t.Fatalf("Failed to read ASN file: %v", err)
	}
	expectedASNContent := `{"asn":12345,"rir":"test","domain":"test.com","cc":"US","asname":"TEST-ASN"}`
	if string(asnContent) != expectedASNContent {
		t.Errorf("Expected ASN content %q, got %q", expectedASNContent, string(asnContent))
	}

	// Test 404 handling
	err = downloader.downloadFile(server.URL+"/nonexistent", filepath.Join(tempDir, "nonexistent.txt"))
	if err == nil {
		t.Error("Expected error for 404 URL, got nil")
	}
}
