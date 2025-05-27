package netintel

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/yl2chen/cidranger"
)

// ReputationClient handles IP reputation/blacklist lookups
type ReputationClient struct {
    rangers map[int]cidranger.Ranger
    dataDir string
}

// NewReputationClient creates a new reputation client
func NewReputationClient(dataDir string) *ReputationClient {
    return &ReputationClient{
        rangers: make(map[int]cidranger.Ranger),
        dataDir: dataDir,
    }
}

// LoadLevel loads a netset file from local storage
func (rc *ReputationClient) LoadLevel(level int) error {
    filePath := fmt.Sprintf("%s/sources/firehol_level%d.netset", rc.dataDir, level)

    file, err := os.Open(filePath)
    if err != nil {
       return err
    }
    defer file.Close()

    ranger := cidranger.NewPCTrieRanger()
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
       line := strings.TrimSpace(scanner.Text())

       // Skip comments and empty lines
       if line == "" || strings.HasPrefix(line, "#") {
          continue
       }

       // Parse CIDR or individual IP
       _, network, err := net.ParseCIDR(line)
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
       }

       ranger.Insert(cidranger.NewBasicRangerEntry(*network))
    }

    rc.rangers[level] = ranger
    return scanner.Err()
}

// LoadAllLevels loads all reputation levels (1-4)
func (rc *ReputationClient) LoadAllLevels() error {
    for level := 1; level <= 4; level++ {
        if err := rc.LoadLevel(level); err != nil {
            return fmt.Errorf("failed to load level %d: %w", level, err)
        }
    }
    return nil
}

// Lookup returns the number of occurrences for a given IP/subnet in a specific level
// For Level1: checks subnet matches, for others: checks IP matches
func (rc *ReputationClient) Lookup(level int, target string) int {
    ranger, exists := rc.rangers[level]
    if !exists {
       return 0
    }

    if level == 1 {
       // For Level1, check subnet matches
       _, subnet, err := net.ParseCIDR(target)
       if err != nil {
          return 0
       }

       entries, err := ranger.ContainingNetworks(subnet.IP)
       if err != nil {
          return 0
       }
       return len(entries)
    } else {
       // For Level2-4, check each IP in the provided subnet
       _, subnet, err := net.ParseCIDR(target)
       if err != nil {
          return 0
       }

       count := 0

       // Iterate through each IP in the subnet
       for ip := subnet.IP.Mask(subnet.Mask); subnet.Contains(ip); inc(ip) {
          entries, err := ranger.ContainingNetworks(ip)
          if err == nil && len(entries) > 0 {
             count++
          }
       }

       return count
    }
}

// GetLoadedLevels returns which levels are currently loaded
func (rc *ReputationClient) GetLoadedLevels() []int {
    levels := make([]int, 0, len(rc.rangers))
    for level := range rc.rangers {
        levels = append(levels, level)
    }
    return levels
}

// IsLevelLoaded checks if a specific level is loaded
func (rc *ReputationClient) IsLevelLoaded(level int) bool {
    _, exists := rc.rangers[level]
    return exists
}

// Clear removes all loaded data
func (rc *ReputationClient) Clear() {
    rc.rangers = make(map[int]cidranger.Ranger)
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