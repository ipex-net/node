package netintel

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"cosmossdk.io/log"
)

const (
	// Update intervals
	reputationUpdateInterval = 1 * time.Minute // Reputation updates every minute
	//	countryUpdateInterval    = 24 * time.Hour   // Country updates every 24 hours
	countryUpdateInterval = 5 * time.Minute // Country updates every 5 minutes (dev mode)
	asnUpdateInterval     = 1 * time.Hour   // ASN DROP list updates every hour
	// Startup timeout for loading existing data from disk
	initialLoadTimeout = 30 * time.Second
)

// Service is the main NetIntel service that orchestrates all components
type Service struct {
	// Core components
	cache      *AtomicCache
	downloader *HTTPDownloader
	dataDir    string
	logger     log.Logger

	// Service state
	running  atomic.Bool
	stopChan chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup

	// Startup state
	initialLoadDone atomic.Bool

	// Test mode (disables background updates)
	testMode bool
}

// NewService creates a new NetIntel service
func NewService(dataDir string, logger log.Logger) *Service {
	return &Service{
		cache:      NewAtomicCache(),
		downloader: NewHTTPDownloader(dataDir, logger),
		dataDir:    dataDir,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// NewServiceForTesting creates a new NetIntel service with background updates disabled
func NewServiceForTesting(dataDir string, logger log.Logger) *Service {
	service := NewService(dataDir, logger)
	service.testMode = true
	return service
}

// Start starts the NetIntel service and background update loops
func (s *Service) Start() error {
	if !s.running.CompareAndSwap(false, true) {
		return fmt.Errorf("service is already running")
	}

	s.logger.Info("[NetIntel] Starting NetIntel service", "data_dir", s.dataDir)

	// Load initial data from disk (blocking)
	if err := s.loadInitialData(); err != nil {
		s.running.Store(false)
		return fmt.Errorf("failed to load initial data: %w", err)
	}

	// Start background update loops (unless in test mode)
	if !s.testMode {
		s.startBackgroundUpdates()
	}

	s.logger.Info("[NetIntel] NetIntel service started successfully", "test_mode", s.testMode)
	return nil
}

// Stop gracefully stops the NetIntel service
func (s *Service) Stop() error {
	if !s.running.Load() {
		return nil
	}

	s.logger.Info("[NetIntel] Stopping NetIntel service")

	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.running.Store(false)
	})

	// Wait for background goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Wait for shutdown with timeout
	select {
	case <-done:
		s.logger.Info("[NetIntel] NetIntel service stopped successfully")
	case <-time.After(30 * time.Second):
		s.logger.Warn("[NetIntel] NetIntel service shutdown timed out")
	}

	return nil
}

// IsRunning returns true if the service is running
func (s *Service) IsRunning() bool {
	return s.running.Load()
}

// IsReady returns true if the service is ready to serve requests
func (s *Service) IsReady() bool {
	return s.IsRunning() && s.initialLoadDone.Load() && s.cache.IsReady()
}

// CheckReputation checks IP reputation at specified level
func (s *Service) CheckReputation(ip string, level int) int {
	if !s.IsReady() {
		return 0
	}
	return s.cache.CheckReputation(level, ip)
}

// ScanSubnetReputation scans a subnet for reputation matches
func (s *Service) ScanSubnetReputation(subnet string, level int) int {
	if !s.IsReady() {
		return 0
	}
	return s.cache.CheckReputation(level, subnet)
}

// CheckCountry checks if IP belongs to specified country
func (s *Service) CheckCountry(ip string, countryCode string) bool {
	if !s.IsReady() {
		return false
	}
	return s.cache.CheckCountry(ip, countryCode)
}

// CheckASNDrop checks if ASN is in the DROP list
func (s *Service) CheckASNDrop(asn int32) bool {
	if !s.IsReady() {
		return false
	}
	return s.cache.CheckASNDrop(asn)
}

// GetLastUpdate returns when data was last updated
func (s *Service) GetLastUpdate() time.Time {
	return s.cache.GetLastUpdate()
}

// GetLoadedLevels returns which reputation levels are loaded
func (s *Service) GetLoadedLevels() []int {
	repChecker := s.cache.GetReputationChecker()
	if repChecker == nil {
		return []int{}
	}
	return repChecker.GetLoadedLevels()
}

// GetLoadedCountries returns which countries are loaded
func (s *Service) GetLoadedCountries() []string {
	geoChecker := s.cache.GetGeolocationChecker()
	if geoChecker == nil {
		return []string{}
	}
	return geoChecker.GetLoadedCountries()
}

// loadInitialData loads existing data from disk on startup
func (s *Service) loadInitialData() error {
	s.logger.Info("[NetIntel] Loading initial data from disk")

	ctx, cancel := context.WithTimeout(context.Background(), initialLoadTimeout)
	defer cancel()

	// Load reputation data
	repChecker := NewReputationClient(s.dataDir)
	if err := s.loadWithTimeout(ctx, "reputation", func() error {
		return repChecker.LoadAllLevels()
	}); err != nil {
		s.logger.Warn("Failed to load reputation data", "error", err)
	}

	// Load geolocation data
	geoChecker := NewGeolocationClient(s.dataDir)
	if err := s.loadWithTimeout(ctx, "geolocation", func() error {
		return geoChecker.LoadAllCountries()
	}); err != nil {
		s.logger.Warn("Failed to load geolocation data", "error", err)
	}

	// Load ASN DROP list data (optional - failure doesn't block startup)
	asnChecker := NewASNClient(s.dataDir)
	if err := s.loadWithTimeout(ctx, "ASN DROP list", func() error {
		return asnChecker.LoadASNDropList()
	}); err != nil {
		s.logger.Warn("[NetIntel] Failed to load ASN DROP list data on startup (will retry during updates)", "error", err)
	} else if asnChecker.GetLoadedASNCount() > 0 {
		s.cache.UpdateASNChecker(asnChecker)
		s.logger.Info("[NetIntel] Loaded ASN DROP list data", "count", asnChecker.GetLoadedASNCount())
	}

	// Update cache with loaded data (core services only)
	if len(repChecker.GetLoadedLevels()) > 0 {
		s.cache.UpdateReputationChecker(repChecker)
		s.logger.Info("[NetIntel] Loaded reputation data", "levels", repChecker.GetLoadedLevels())
	}

	if len(geoChecker.GetLoadedCountries()) > 0 {
		s.cache.UpdateGeolocationChecker(geoChecker)
		s.logger.Info("[NetIntel] Loaded geolocation data", "countries", len(geoChecker.GetLoadedCountries()))
	}

	s.initialLoadDone.Store(true)
	s.logger.Info("[NetIntel] Initial data loading completed")
	return nil
}

// loadWithTimeout loads data with context timeout
func (s *Service) loadWithTimeout(ctx context.Context, name string, loadFunc func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- loadFunc()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("%s data loading timed out", name)
	}
}

// startBackgroundUpdates starts the background update goroutines
func (s *Service) startBackgroundUpdates() {
	// Start reputation update loop (every minute)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.reputationUpdateLoop()
	}()

	// Start country update loop (every 5 minutes in dev, 24 hours in prod)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.countryUpdateLoop()
	}()

	// Start ASN update loop (every hour)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.asnUpdateLoop()
	}()
}

// reputationUpdateLoop runs the reputation data update loop
func (s *Service) reputationUpdateLoop() {
	s.logger.Info("[NetIntel] Starting reputation update loop", "interval", reputationUpdateInterval)
	s.updateReputationData()
	ticker := time.NewTicker(reputationUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("[NetIntel] Reputation update loop stopped")
			return
		case <-ticker.C:
			s.updateReputationData()
		}
	}
}

// countryUpdateLoop runs the country data update loop
func (s *Service) countryUpdateLoop() {
	s.logger.Info("[NetIntel] Starting country update loop", "interval", countryUpdateInterval)
	s.updateCountryData()
	ticker := time.NewTicker(countryUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("[NetIntel] Country update loop stopped")
			return
		case <-ticker.C:
			s.updateCountryData()
		}
	}
}

// asnUpdateLoop runs the ASN data update loop
func (s *Service) asnUpdateLoop() {
	s.logger.Info("[NetIntel] Starting ASN update loop", "interval", asnUpdateInterval)
	s.updateASNData()
	ticker := time.NewTicker(asnUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Info("[NetIntel] ASN update loop stopped")
			return
		case <-ticker.C:
			s.updateASNData()
		}
	}
}

// updateReputationData downloads and updates reputation data
func (s *Service) updateReputationData() {
	s.logger.Info("[NetIntel] Updating reputation data")

	start := time.Now().UTC()

	// Download fresh reputation data (skip in test mode)
	if !s.testMode {
		if err := s.downloader.DownloadAllReputationLevels(); err != nil {
			s.logger.Error("Failed to download reputation data", "error", err)
			return
		}
	}

	// Load fresh reputation data into new client
	newRepChecker := NewReputationClient(s.dataDir)
	if err := newRepChecker.LoadAllLevels(); err != nil {
		s.logger.Error("Failed to load new reputation data", "error", err)
		return
	}

	// Atomic swap into cache
	s.cache.UpdateReputationChecker(newRepChecker)

	duration := time.Since(start)
	s.logger.Info("[NetIntel] Reputation data updated successfully",
		"duration", duration,
		"levels", newRepChecker.GetLoadedLevels())
}

// updateCountryData downloads and updates country data
func (s *Service) updateCountryData() {
	s.logger.Info("[NetIntel] Updating country data")

	start := time.Now().UTC()

	// Download fresh data (skip in test mode)
	if !s.testMode {
		if err := s.downloader.DownloadAllCountryData(); err != nil {
			s.logger.Error("Failed to download country data", "error", err)
			return
		}
	}

	// Load fresh data into new client
	newGeoChecker := NewGeolocationClient(s.dataDir)
	if err := newGeoChecker.LoadAllCountries(); err != nil {
		s.logger.Error("Failed to load new country data", "error", err)
		return
	}

	// Atomic swap into cache
	s.cache.UpdateGeolocationChecker(newGeoChecker)

	duration := time.Since(start)
	s.logger.Info("[NetIntel] Country data updated successfully",
		"duration", duration,
		"countries", len(newGeoChecker.GetLoadedCountries()))
}

// updateASNData downloads and updates ASN DROP list data
func (s *Service) updateASNData() {
	s.logger.Info("[NetIntel] Updating ASN DROP list data")

	start := time.Now().UTC()

	// Download fresh ASN DROP list (skip in test mode)
	if !s.testMode {
		if err := s.downloader.DownloadASNDropList(); err != nil {
			s.logger.Error("Failed to download ASN DROP list", "error", err)
			return
		}
	}

	// Load fresh ASN data into new client
	newASNChecker := NewASNClient(s.dataDir)
	if err := newASNChecker.LoadASNDropList(); err != nil {
		s.logger.Error("Failed to load new ASN DROP list data", "error", err)
		return
	}

	// Atomic swap into cache
	s.cache.UpdateASNChecker(newASNChecker)

	duration := time.Since(start)
	s.logger.Info("[NetIntel] ASN DROP list updated successfully",
		"duration", duration,
		"asn_count", newASNChecker.GetLoadedASNCount())
}

// GetStatus returns detailed service status
func (s *Service) GetStatus() ServiceStatus {
	cacheStatus := s.cache.GetStatus()

	return ServiceStatus{
		Running:         s.IsRunning(),
		Ready:           s.IsReady(),
		LastUpdate:      cacheStatus.LastUpdate,
		LastDownload:    s.downloader.GetLastDownload(),
		LoadedLevels:    cacheStatus.LoadedLevels,
		LoadedCountries: cacheStatus.LoadedCountries,
		LoadedASNCount:  cacheStatus.LoadedASNCount,
		ErrorCount:      0,
		LastError:       nil,
	}
}
