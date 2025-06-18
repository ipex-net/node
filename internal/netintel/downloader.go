package netintel

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// HTTPDownloader handles downloading data from remote sources
type HTTPDownloader struct {
	client       *http.Client
	dataDir      string
	logger       log.Logger
	lastDownload time.Time
	mu           sync.RWMutex
}

// NewHTTPDownloader creates a new HTTP downloader
func NewHTTPDownloader(dataDir string, logger log.Logger) *HTTPDownloader {
	return &HTTPDownloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		dataDir: dataDir,
		logger:  logger,
	}
}

// DownloadReputationLevel downloads a specific reputation level file
func (d *HTTPDownloader) DownloadReputationLevel(level int) error {
	if level < 1 || level > 4 {
		return fmt.Errorf("invalid reputation level: %d (must be 1-4)", level)
	}

	// TODO: TURN THIS BACK ON FOR PRODUCTION:
	// url := fmt.Sprintf("https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level%d.netset", level)
	// DEBUG:
	url := fmt.Sprintf("https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level%d.netset", level)
	destPath := filepath.Join(d.dataDir, "sources", fmt.Sprintf("firehol_level%d.netset", level))

	d.logger.Info("[NetIntel] Downloading reputation level", "level", level, "url", url)

	if err := d.downloadFile(url, destPath); err != nil {
		return fmt.Errorf("failed to download reputation level %d: %w", level, err)
	}

	d.updateLastDownload()
	d.logger.Info("[NetIntel] Successfully downloaded reputation level", "level", level, "path", destPath)
	return nil
}

// DownloadAllReputationLevels downloads all reputation levels (1-4)
func (d *HTTPDownloader) DownloadAllReputationLevels() error {
	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	// Download all levels concurrently
	for level := 1; level <= 4; level++ {
		wg.Add(1)
		go func(l int) {
			defer wg.Done()
			if err := d.DownloadReputationLevel(l); err != nil {
				errChan <- err
			}
		}(level)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to download %d reputation levels: %v", len(errors), errors)
	}

	d.logger.Info("[NetIntel] Successfully downloaded all reputation levels")
	return nil
}

// DownloadCountryData downloads country data for specific country from all providers
func (d *HTTPDownloader) DownloadCountryData(countryCode string) error {
	// Define all country data sources
	sources := []struct {
		name    string
		urlPath string
		dirName string
	}{
		{
			name:    "GeoLite2",
			urlPath: fmt.Sprintf("geolite2_country/country_%s.netset", countryCode),
			dirName: "geolite2_country",
		},
		{
			name:    "IP2Location",
			urlPath: fmt.Sprintf("ip2location_country/ip2location_country_%s.netset", countryCode),
			dirName: "ip2location_country",
		},
		{
			name:    "IPDeny",
			urlPath: fmt.Sprintf("ipdeny_country/id_country_%s.netset", countryCode),
			dirName: "ipdeny_country",
		},
		{
			name:    "IPIP",
			urlPath: fmt.Sprintf("ipip_country/ipip_country_%s.netset", countryCode),
			dirName: "ipip_country",
		},
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sources))
	successCount := 0
	var mu sync.Mutex

	// Download from all sources concurrently
	for _, source := range sources {
		wg.Add(1)
		go func(src struct {
			name    string
			urlPath string
			dirName string
		}) {
			defer wg.Done()

			url := fmt.Sprintf("https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/%s", src.urlPath)
			destDir := filepath.Join(d.dataDir, "sources", src.dirName)
			destPath := filepath.Join(destDir, filepath.Base(src.urlPath))

			d.logger.Debug("Downloading country data", "country", countryCode, "source", src.name, "url", url)

			if err := d.downloadFile(url, destPath); err != nil {
				d.logger.Warn("Failed to download country data from source", "country", countryCode, "source", src.name, "error", err)
				errChan <- fmt.Errorf("source %s: %w", src.name, err)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
				d.logger.Debug("Successfully downloaded country data", "country", countryCode, "source", src.name)
			}
		}(source)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if successCount == 0 {
		return fmt.Errorf("failed to download country %s from any source: %v", countryCode, errors)
	}

	d.updateLastDownload()
	d.logger.Info("[NetIntel] Downloaded country data", "country", countryCode, "sources", successCount, "total", len(sources))
	return nil
}

// DownloadAllCountryData downloads country data for all countries
func (d *HTTPDownloader) DownloadAllCountryData() error {
	// ISO 3166-1 Alpha-2 country codes (same as in geolocation.go)
	countryCodes := []string{
		"ad", "ae", "af", "ag", "ai", "al", "am", "ao", "aq", "ar", "as", "at", "au", "aw", "ax", "az",
		"ba", "bb", "bd", "be", "bf", "bg", "bh", "bi", "bj", "bl", "bm", "bn", "bo", "bq", "br", "bs",
		"bt", "bv", "bw", "by", "bz", "ca", "cc", "cd", "cf", "cg", "ch", "ci", "ck", "cl", "cm", "cn",
		"co", "cr", "cu", "cv", "cw", "cx", "cy", "cz", "de", "dj", "dk", "dm", "do", "dz", "ec", "ee",
		"eg", "eh", "er", "es", "et", "fi", "fj", "fk", "fm", "fo", "fr", "ga", "gb", "gd", "ge", "gf",
		"gg", "gh", "gi", "gl", "gm", "gn", "gp", "gq", "gr", "gs", "gt", "gu", "gw", "gy", "hk", "hm",
		"hn", "hr", "ht", "hu", "id", "ie", "il", "im", "in", "io", "iq", "ir", "is", "it", "je", "jm",
		"jo", "jp", "ke", "kg", "kh", "ki", "km", "kn", "kp", "kr", "kw", "ky", "kz", "la", "lb", "lc",
		"li", "lk", "lr", "ls", "lt", "lu", "lv", "ly", "ma", "mc", "md", "me", "mf", "mg", "mh", "mk",
		"ml", "mm", "mn", "mo", "mp", "mq", "mr", "ms", "mt", "mu", "mv", "mw", "mx", "my", "mz", "na",
		"nc", "ne", "nf", "ng", "ni", "nl", "no", "np", "nr", "nu", "nz", "om", "pa", "pe", "pf", "pg",
		"ph", "pk", "pl", "pm", "pn", "pr", "ps", "pt", "pw", "py", "qa", "re", "ro", "rs", "ru", "rw",
		"sa", "sb", "sc", "sd", "se", "sg", "sh", "si", "sj", "sk", "sl", "sm", "sn", "so", "sr", "ss",
		"st", "sv", "sx", "sy", "sz", "tc", "td", "tf", "tg", "th", "tj", "tk", "tl", "tm", "tn", "to",
		"tr", "tt", "tv", "tw", "tz", "ua", "ug", "um", "us", "uy", "uz", "va", "vc", "ve", "vg", "vi",
		"vn", "vu", "wf", "ws", "ye", "yt", "za", "zm", "zw",
	}

	d.logger.Info("[NetIntel] Starting download of all country data", "countries", len(countryCodes))

	// Use a worker pool to limit concurrent downloads
	const maxWorkers = 10
	countryQueue := make(chan string, len(countryCodes))
	var wg sync.WaitGroup

	// Fill the queue
	for _, countryCode := range countryCodes {
		countryQueue <- countryCode
	}
	close(countryQueue)

	// Start workers
	var successCount int32
	var mu sync.Mutex

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for countryCode := range countryQueue {
				if err := d.DownloadCountryData(countryCode); err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	d.logger.Info("[NetIntel] Completed download of country data", "successful", successCount, "total", len(countryCodes))
	return nil
}

// DownloadAll downloads all data (reputation levels + all countries)
func (d *HTTPDownloader) DownloadAll() error {
	d.logger.Info("[NetIntel] Starting download of all NetIntel data")

	// Download reputation levels first (smaller, faster)
	if err := d.DownloadAllReputationLevels(); err != nil {
		return fmt.Errorf("failed to download reputation levels: %w", err)
	}

	// Download country data
	if err := d.DownloadAllCountryData(); err != nil {
		return fmt.Errorf("failed to download country data: %w", err)
	}

	d.logger.Info("[NetIntel] Successfully downloaded all NetIntel data")
	return nil
}

// GetLastDownload returns when data was last downloaded
func (d *HTTPDownloader) GetLastDownload() time.Time {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastDownload
}

// downloadFile downloads a file from URL to destination path
func (d *HTTPDownloader) downloadFile(url, destPath string) error {
	// Create directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	// Create temporary file
	tempPath := destPath + ".tmp"

	// Download with retries
	var lastErr error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := d.downloadFileAttempt(url, tempPath); err != nil {
			lastErr = err
			if attempt < maxRetries {
				d.logger.Warn("Download attempt failed, retrying", "url", url, "attempt", attempt, "error", err)
				time.Sleep(time.Duration(attempt) * 2 * time.Second) // Exponential backoff
				continue
			}
		} else {
			// Success - move temp file to final location
			if err := os.Rename(tempPath, destPath); err != nil {
				os.Remove(tempPath) // Clean up temp file
				return fmt.Errorf("failed to move downloaded file: %w", err)
			}
			return nil
		}
	}

	// Clean up temp file on final failure
	os.Remove(tempPath)
	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// downloadFileAttempt performs a single download attempt
func (d *HTTPDownloader) downloadFileAttempt(url, destPath string) error {
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// updateLastDownload updates the last download timestamp
func (d *HTTPDownloader) updateLastDownload() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastDownload = time.Now().UTC()
}
