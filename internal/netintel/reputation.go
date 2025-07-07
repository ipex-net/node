package netintel

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/yl2chen/cidranger"
)

// ReputationClient handles IP reputation/blacklist lookups
type ReputationClient struct {
	rangers       map[int]cidranger.Ranger
	individualIPs map[int]map[string]bool // Track individual IPs vs CIDR ranges
	dataDir       string
	mu            sync.RWMutex // Protect map access
}

// NewReputationClient creates a new reputation client
func NewReputationClient(dataDir string) *ReputationClient {
	return &ReputationClient{
		rangers:       make(map[int]cidranger.Ranger),
		individualIPs: make(map[int]map[string]bool),
		dataDir:       dataDir,
	}
}

// LoadLevel loads a netset file from local storage
func (rc *ReputationClient) LoadLevel(level int) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	filePath := fmt.Sprintf("%s/sources/firehol_level%d.netset", rc.dataDir, level)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ranger := cidranger.NewPCTrieRanger()
	individualIPSet := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse CIDR or individual IP
		_, network, err := net.ParseCIDR(line)
		isIndividualIP := false
		if err != nil {
			// Try parsing as individual IP
			ip := net.ParseIP(line)
			if ip == nil {
				continue // Skip invalid entries
			}

			// Convert single IP to /32 CIDR
			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}
			network = &net.IPNet{IP: ip, Mask: mask}
			isIndividualIP = true
		}

		ranger.Insert(cidranger.NewBasicRangerEntry(*network))

		// Track individual IPs
		if isIndividualIP {
			individualIPSet[network.IP.String()] = true
		}
	}

	rc.rangers[level] = ranger
	rc.individualIPs[level] = individualIPSet
	return scanner.Err()
}

// LoadAllLevels loads all available reputation levels (1-4)
// Missing files are skipped without error
func (rc *ReputationClient) LoadAllLevels() error {
	var errors []error
	loadedCount := 0

	for level := 1; level <= 4; level++ {
		if err := rc.LoadLevel(level); err != nil {
			// Check if it's a file not found error
			if os.IsNotExist(err) {
				continue // Skip missing files
			}
			errors = append(errors, fmt.Errorf("level %d: %w", level, err))
		} else {
			loadedCount++
		}
	}

	// Return error only if we had non-missing-file errors
	if len(errors) > 0 {
		return fmt.Errorf("failed to load reputation levels: %v", errors)
	}

	// If no levels were loaded at all, that might be an issue
	if loadedCount == 0 {
		return fmt.Errorf("no reputation levels could be loaded")
	}

	return nil
}

// Lookup returns the number of occurrences for a given IP/subnet in a specific level
// For individual IPs or /32 CIDRs: checks if the IP exists in reputation data (individual IPs or CIDR ranges)
// For subnets larger than /32: scans the subnet and counts individual IP entries (not CIDR ranges)
func (rc *ReputationClient) Lookup(level int, target string) int {
	rc.mu.RLock()
	ranger, rangerExists := rc.rangers[level]
	individualIPs, individualExists := rc.individualIPs[level]
	rc.mu.RUnlock()

	if !rangerExists || !individualExists {
		return 0
	}

	// Try parsing as CIDR first
	_, subnet, err := net.ParseCIDR(target)
	if err != nil {
		// Not a CIDR, try parsing as individual IP
		ip := net.ParseIP(target)
		if ip == nil {
			return 0 // Invalid input
		}

		// For individual IP lookups, check both individual IPs and CIDR ranges
		entries, err := ranger.ContainingNetworks(ip)
		if err == nil && len(entries) > 0 {
			return 1 // IP found in reputation data
		}
		return 0
	}

	// Check if this is a single IP lookup (/32 for IPv4, /128 for IPv6)
	ones, bits := subnet.Mask.Size()
	if ones == bits {
		// This is a single IP lookup - check if the IP exists in reputation data
		entries, err := ranger.ContainingNetworks(subnet.IP)
		if err == nil && len(entries) > 0 {
			return 1 // IP found in reputation data
		}
		return 0
	}

	// For subnet lookups larger than single IP, count only individual IP entries within the subnet
	count := 0
	for ip := subnet.IP.Mask(subnet.Mask); subnet.Contains(ip); inc(ip) {
		if individualIPs[ip.String()] {
			count++
		}
	}
	return count
}

// GetLoadedLevels returns which levels are currently loaded
func (rc *ReputationClient) GetLoadedLevels() []int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	levels := make([]int, 0, len(rc.rangers))
	for level := range rc.rangers {
		levels = append(levels, level)
	}
	return levels
}

// IsLevelLoaded checks if a specific level is loaded
func (rc *ReputationClient) IsLevelLoaded(level int) bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	_, exists := rc.rangers[level]
	return exists
}

// Clear removes all loaded data
func (rc *ReputationClient) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.rangers = make(map[int]cidranger.Ranger)
	rc.individualIPs = make(map[int]map[string]bool)
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
