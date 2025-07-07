package netintel

import (
	"sync/atomic"
	"time"
)

// AtomicCache provides thread-safe cache with atomic hot-swapping
type AtomicCache struct {
	// Atomic pointers for zero-copy reads during updates
	reputationChecker  atomic.Pointer[ReputationClient]
	geolocationChecker atomic.Pointer[GeolocationClient]
	asnChecker         atomic.Pointer[ASNClient]

	// Metadata
	lastUpdate atomic.Int64 // Unix timestamp of last update
}

// NewAtomicCache creates a new atomic cache
func NewAtomicCache() *AtomicCache {
	return &AtomicCache{}
}

// GetReputationChecker returns current reputation checker (thread-safe read)
// Returns nil if not yet initialized
func (c *AtomicCache) GetReputationChecker() *ReputationClient {
	return c.reputationChecker.Load()
}

// GetGeolocationChecker returns current geolocation checker (thread-safe read)
// Returns nil if not yet initialized
func (c *AtomicCache) GetGeolocationChecker() *GeolocationClient {
	return c.geolocationChecker.Load()
}

// GetASNChecker returns current ASN checker (thread-safe read)
// Returns nil if not yet initialized
func (c *AtomicCache) GetASNChecker() *ASNClient {
	return c.asnChecker.Load()
}

// UpdateReputationChecker atomically updates reputation checker
// This is a hot-swap - existing readers continue with old data until this completes
func (c *AtomicCache) UpdateReputationChecker(checker *ReputationClient) {
	c.reputationChecker.Store(checker)
	c.updateTimestamp()
}

// UpdateGeolocationChecker atomically updates geolocation checker
// This is a hot-swap - existing readers continue with old data until this completes
func (c *AtomicCache) UpdateGeolocationChecker(checker *GeolocationClient) {
	c.geolocationChecker.Store(checker)
	c.updateTimestamp()
}

// UpdateASNChecker atomically updates ASN checker
// This is a hot-swap - existing readers continue with old data until this completes
func (c *AtomicCache) UpdateASNChecker(checker *ASNClient) {
	c.asnChecker.Store(checker)
	c.updateTimestamp()
}

// UpdateBoth atomically updates both checkers
// Note: This is NOT a single atomic operation - there's a tiny window where
// one is updated and the other isn't. If you need true atomicity, use a single
// struct containing both checkers.
func (c *AtomicCache) UpdateBoth(repChecker *ReputationClient, geoChecker *GeolocationClient) {
	c.reputationChecker.Store(repChecker)
	c.geolocationChecker.Store(geoChecker)
	c.updateTimestamp()
}

// GetLastUpdate returns when cache was last updated (thread-safe)
func (c *AtomicCache) GetLastUpdate() time.Time {
	timestamp := c.lastUpdate.Load()
	if timestamp == 0 {
		return time.Time{} // Zero time if never updated
	}
	return time.Unix(timestamp, 0)
}

// IsReady returns true if geolocation checker is loaded (core requirement)
// Reputation and ASN checkers are optional and don't affect readiness
func (c *AtomicCache) IsReady() bool {
	return c.geolocationChecker.Load() != nil
}

// IsReputationReady returns true if reputation checker is loaded
func (c *AtomicCache) IsReputationReady() bool {
	return c.reputationChecker.Load() != nil
}

// IsGeolocationReady returns true if geolocation checker is loaded
func (c *AtomicCache) IsGeolocationReady() bool {
	return c.geolocationChecker.Load() != nil
}

// IsASNReady returns true if ASN checker is loaded
func (c *AtomicCache) IsASNReady() bool {
	return c.asnChecker.Load() != nil
}

// Clear removes all data from cache (sets to nil)
func (c *AtomicCache) Clear() {
	c.reputationChecker.Store(nil)
	c.geolocationChecker.Store(nil)
	c.asnChecker.Store(nil)
	c.updateTimestamp()
}

// updateTimestamp updates the last update timestamp to current time
func (c *AtomicCache) updateTimestamp() {
	c.lastUpdate.Store(time.Now().Unix())
}

// GetStatus returns detailed cache status
func (c *AtomicCache) GetStatus() CacheStatus {
	repChecker := c.reputationChecker.Load()
	geoChecker := c.geolocationChecker.Load()
	asnChecker := c.asnChecker.Load()

	status := CacheStatus{
		LastUpdate:       c.GetLastUpdate(),
		ReputationReady:  repChecker != nil,
		GeolocationReady: geoChecker != nil,
		ASNReady:         asnChecker != nil,
		LoadedLevels:     []int{},
		LoadedCountries:  []string{},
		LoadedASNCount:   0,
	}

	// Get loaded levels if reputation checker exists
	if repChecker != nil {
		status.LoadedLevels = repChecker.GetLoadedLevels()
	}

	// Get loaded countries if geolocation checker exists
	if geoChecker != nil {
		status.LoadedCountries = geoChecker.GetLoadedCountries()
	}

	// Get loaded ASN count if ASN checker exists
	if asnChecker != nil {
		status.LoadedASNCount = asnChecker.GetLoadedASNCount()
	}

	return status
}

// CacheStatus represents the current status of the cache
type CacheStatus struct {
	LastUpdate       time.Time
	ReputationReady  bool
	GeolocationReady bool
	ASNReady         bool
	LoadedLevels     []int
	LoadedCountries  []string
	LoadedASNCount   int
}

// Convenience methods for common operations

// CheckReputation is a convenience method that checks if cache is ready and performs lookup
// Returns 0 if cache is not ready or lookup fails
func (c *AtomicCache) CheckReputation(level int, target string) int {
	checker := c.GetReputationChecker()
	if checker == nil {
		return 0
	}
	return checker.Lookup(level, target)
}

// CheckCountry is a convenience method that checks if cache is ready and performs lookup
// Returns false if cache is not ready or lookup fails
func (c *AtomicCache) CheckCountry(ipAddress string, countryCode string) bool {
	checker := c.GetGeolocationChecker()
	if checker == nil {
		return false
	}
	return checker.CheckCountry(ipAddress, countryCode)
}

// CheckASNDrop is a convenience method that checks if cache is ready and performs ASN DROP lookup
// Returns false if cache is not ready or lookup fails
func (c *AtomicCache) CheckASNDrop(asn int32) bool {
	checker := c.GetASNChecker()
	if checker == nil {
		return false
	}
	return checker.CheckASNDrop(asn)
}
