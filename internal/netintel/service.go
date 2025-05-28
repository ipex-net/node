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
	countryUpdateInterval = 5 * time.Minute // Country updates every 24 hours
	// Startup timeout
	initialLoadTimeout = 5 * time.Minute
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

	// Start background update loops
	s.startBackgroundUpdates()

	s.logger.Info("[NetIntel] NetIntel service started successfully")
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

	// Update cache with loaded data
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

	// Start country update loop (every 24 hours)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.countryUpdateLoop()
	}()
}

// reputationUpdateLoop runs the reputation data update loop
func (s *Service) reputationUpdateLoop() {
	s.logger.Info("[NetIntel] Starting reputation update loop", "interval", reputationUpdateInterval)

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

// updateReputationData downloads and updates reputation data
func (s *Service) updateReputationData() {
	s.logger.Info("[NetIntel] Updating reputation data")

	start := time.Now().UTC()

	// Download fresh data
	if err := s.downloader.DownloadAllReputationLevels(); err != nil {
		s.logger.Error("Failed to download reputation data", "error", err)
		return
	}

	// Load fresh data into new client
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

	// Download fresh data (this takes a while)
	if err := s.downloader.DownloadAllCountryData(); err != nil {
		s.logger.Error("Failed to download country data", "error", err)
		return
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
		ErrorCount:      0,
		LastError:       nil,
	}
}
