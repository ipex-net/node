package netintel

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/yl2chen/cidranger"
)

// GeolocationClient handles country-based IP lookups
type GeolocationClient struct {
	countryRangers map[string]cidranger.Ranger
	dataDir        string
}

// NewGeolocationClient creates a new geolocation client
func NewGeolocationClient(dataDir string) *GeolocationClient {
	return &GeolocationClient{
		countryRangers: make(map[string]cidranger.Ranger),
		dataDir:        dataDir,
	}
}

// LoadCountry loads country IP lists from ALL sources and merges them
// Creates a super permissive matcher using data from MaxMind, IP2Location, IPDeny, and IPIP
func (gc *GeolocationClient) LoadCountry(countryCode string) error {
	countryCode = strings.ToLower(countryCode)
	ranger := cidranger.NewPCTrieRanger()

	// Define all source file patterns
	sources := []string{
		fmt.Sprintf("%s/sources/geolite2_country/country_%s.netset", gc.dataDir, countryCode),
		fmt.Sprintf("%s/sources/ip2location_country/ip2location_country_%s.netset", gc.dataDir, countryCode),
		fmt.Sprintf("%s/sources/ipdeny_country/id_country_%s.netset", gc.dataDir, countryCode),
		fmt.Sprintf("%s/sources/ipip_country/ipip_country_%s.netset", gc.dataDir, countryCode),
	}

	loadedSources := 0

	// Load from all available sources
	for _, filePath := range sources {
		err := gc.loadCountryFromFile(filePath, ranger)
		if err == nil {
			loadedSources++
		}
		// Continue even if some sources fail - we want maximum coverage
	}

	if loadedSources == 0 {
		return fmt.Errorf("no country data found for %s from any source", countryCode)
	}

	gc.countryRangers[countryCode] = ranger
	return nil
}

// LoadAllCountries loads all available countries from all sources
func (gc *GeolocationClient) LoadAllCountries() error {
	// ISO 3166-1 Alpha-2 country codes
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

	successCount := 0
	for _, countryCode := range countryCodes {
		if err := gc.LoadCountry(countryCode); err == nil {
			successCount++
		}
		// Continue loading other countries even if some fail
	}

	if successCount == 0 {
		return fmt.Errorf("failed to load any country data")
	}

	return nil
}

// loadCountryFromFile helper function to load from a single file
func (gc *GeolocationClient) loadCountryFromFile(filePath string, ranger cidranger.Ranger) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

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

		// Insert into the shared ranger (duplicates will be handled by the trie)
		ranger.Insert(cidranger.NewBasicRangerEntry(*network))
	}

	return scanner.Err()
}

// CheckCountry checks if an individual IP belongs to a specific country
// Returns true if the IP is located in the specified country (super permissive - any source match)
func (gc *GeolocationClient) CheckCountry(ipAddress string, countryCode string) bool {
	ranger, exists := gc.countryRangers[strings.ToLower(countryCode)]
	if !exists {
		return false
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	// Check if any networks in the country ranger contain this IP
	entries, err := ranger.ContainingNetworks(ip)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

// GetLoadedCountries returns which countries are currently loaded
func (gc *GeolocationClient) GetLoadedCountries() []string {
	countries := make([]string, 0, len(gc.countryRangers))
	for country := range gc.countryRangers {
		countries = append(countries, country)
	}
	return countries
}

// IsCountryLoaded checks if a specific country is loaded
func (gc *GeolocationClient) IsCountryLoaded(countryCode string) bool {
	_, exists := gc.countryRangers[strings.ToLower(countryCode)]
	return exists
}

// Clear removes all loaded data
func (gc *GeolocationClient) Clear() {
	gc.countryRangers = make(map[string]cidranger.Ranger)
}
