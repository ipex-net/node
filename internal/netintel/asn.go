package netintel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SpamhausASNEntry represents a single ASN entry from Spamhaus ASN DROP list
type SpamhausASNEntry struct {
	ASN    int32  `json:"asn"`
	RIR    string `json:"rir"`
	Domain string `json:"domain"`
	CC     string `json:"cc"`
	ASName string `json:"asname"`
	Type   string `json:"type,omitempty"` // For metadata entries
}

// ASNClient handles ASN DROP list lookups
type ASNClient struct {
	dropList map[int32]SpamhausASNEntry
	dataDir  string
}

// NewASNClient creates a new ASN client
func NewASNClient(dataDir string) *ASNClient {
	return &ASNClient{
		dropList: make(map[int32]SpamhausASNEntry),
		dataDir:  dataDir,
	}
}

// LoadASNDropList loads ASN DROP list from local JSON file
func (ac *ASNClient) LoadASNDropList() error {
	filePath := fmt.Sprintf("%s/sources/asndrop.json", ac.dataDir)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open ASN DROP list file: %w", err)
	}
	defer file.Close()

	// Check if file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get ASN DROP list file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		// Empty file is valid - just means no ASNs to load
		ac.dropList = make(map[int32]SpamhausASNEntry)
		return nil
	}

	// Parse newline-delimited JSON (each line is a separate JSON object)
	newDropList := make(map[int32]SpamhausASNEntry)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		var entry SpamhausASNEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return fmt.Errorf("failed to decode ASN DROP list JSON: %w", err)
		}

		// Skip metadata entries
		if entry.Type == "metadata" {
			continue
		}
		// Only include entries with valid ASN numbers
		if entry.ASN > 0 {
			newDropList[entry.ASN] = entry
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading ASN DROP list file: %w", err)
	}

	// Atomically replace the drop list
	ac.dropList = newDropList

	return nil
}

// CheckASNDrop checks if the given ASN is in the DROP list
func (ac *ASNClient) CheckASNDrop(asn int32) bool {
	_, exists := ac.dropList[asn]
	return exists
}

// GetLoadedASNCount returns the number of ASNs loaded in the DROP list
func (ac *ASNClient) GetLoadedASNCount() int {
	return len(ac.dropList)
}
