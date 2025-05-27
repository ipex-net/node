package netintel

import (
	"time"
)

// NetIntelService is the main interface for network intelligence operations
type NetIntelService interface {
	// Service lifecycle
	Start() error
	Stop() error
	IsReady() bool
	
	// Reputation checking
	CheckReputation(ip string, level int) int
	ScanSubnetReputation(subnet string, level int) int
	
	// Geolocation checking
	CheckCountry(ip string, countryCode string) bool
	
	// Service status
	GetLastUpdate() time.Time
	GetLoadedLevels() []int
	GetLoadedCountries() []string
}

// ReputationChecker handles IP reputation/blacklist lookups
type ReputationChecker interface {
	// LoadLevel loads a specific reputation level (1-4)
	LoadLevel(level int) error
	
	// Lookup checks reputation for IP/subnet at specific level
	Lookup(level int, target string) int
	
	// GetLoadedLevels returns which levels are currently loaded
	GetLoadedLevels() []int
}

// GeolocationChecker handles country-based IP lookups
type GeolocationChecker interface {
	// LoadCountry loads country data from all sources
	LoadCountry(countryCode string) error
	
	// CheckCountry checks if IP belongs to specific country
	CheckCountry(ip string, countryCode string) bool
	
	// GetLoadedCountries returns which countries are currently loaded
	GetLoadedCountries() []string
}

// Downloader handles fetching data from remote sources
type Downloader interface {
	// DownloadReputationLevel downloads a specific level file
	DownloadReputationLevel(level int) error
	
	// DownloadCountryData downloads country data for specific country and all providers
	DownloadCountryData(countryCode string) error
	
	// DownloadAll downloads all configured data
	DownloadAll() error
	
	// GetLastDownload returns when data was last downloaded
	GetLastDownload() time.Time
}

// Cache represents thread-safe cache with atomic updates
type Cache interface {
	// GetReputationChecker returns current reputation checker (thread-safe read)
	GetReputationChecker() ReputationChecker
	
	// GetGeolocationChecker returns current geolocation checker (thread-safe read)
	GetGeolocationChecker() GeolocationChecker
	
	// UpdateReputationChecker atomically updates reputation checker
	UpdateReputationChecker(checker ReputationChecker)
	
	// UpdateGeolocationChecker atomically updates geolocation checker  
	UpdateGeolocationChecker(checker GeolocationChecker)
	
	// GetLastUpdate returns when cache was last updated
	GetLastUpdate() time.Time
}

// Config holds configuration for the NetIntel service
type Config struct {
	// Service settings
	Enabled        bool          `mapstructure:"enabled"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
	DataDir        string        `mapstructure:"data_dir"`
	
	// Reputation settings
	ReputationLevels []int `mapstructure:"reputation_levels"`
	
	// Geolocation settings
	Countries []string `mapstructure:"countries"`
	
	// Download settings
	DownloadTimeout time.Duration `mapstructure:"download_timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
}

// ServiceStatus represents the current status of the service
type ServiceStatus struct {
	Running         bool
	Ready           bool
	LastUpdate      time.Time
	LastDownload    time.Time
	LoadedLevels    []int
	LoadedCountries []string
	ErrorCount      int
	LastError       error
}

// Common errors
var (
	ErrServiceNotReady    = NewNetIntelError("service not ready")
	ErrServiceNotRunning  = NewNetIntelError("service not running")
	ErrInvalidLevel       = NewNetIntelError("invalid reputation level")
	ErrInvalidCountry     = NewNetIntelError("invalid country code")
	ErrDownloadFailed     = NewNetIntelError("download failed")
	ErrLoadFailed         = NewNetIntelError("load failed")
)

// NetIntelError represents errors specific to the NetIntel service
type NetIntelError struct {
	Message string
	Cause   error
}

func (e *NetIntelError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *NetIntelError) Unwrap() error {
	return e.Cause
}

// NewNetIntelError creates a new NetIntel error
func NewNetIntelError(message string) *NetIntelError {
	return &NetIntelError{Message: message}
}

// NewNetIntelErrorWithCause creates a new NetIntel error with a cause
func NewNetIntelErrorWithCause(message string, cause error) *NetIntelError {
	return &NetIntelError{Message: message, Cause: cause}
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		UpdateInterval: 6 * time.Hour,
		DataDir:        "./data/firehol",
		ReputationLevels: []int{1, 2, 3, 4},
		Countries: []string{
			"us", "cn", "ru", "ir", "kp", "sy", "cu", "mm", "by", "ni",
			// Add more countries as needed
		},
		DownloadTimeout: 30 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
	}
}