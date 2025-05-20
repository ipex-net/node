package abci

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"ipex/x/oracle/types"
	subnettypes "ipex/x/subnet/types"

	"cosmossdk.io/log"
	"github.com/openrdap/rdap"
	geoip "github.com/oschwald/geoip2-golang"
	"golang.org/x/sync/errgroup"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
    DNSLookupTimeout  = 1 * time.Second
    RDAPQueryTimeout  = 1 * time.Second
	RPKIQueryTimeout  = 1 * time.Second
    TracerouteTimeout = 5 * time.Second
	rDNSLookupTimeout = 1 * time.Second
)

type VoteExtHandler struct {
	logger        log.Logger
	currentBlock  int64        

    OrgKeeper    types.OrgKeeper
    SubnetKeeper types.SubnetKeeper
    LeaseKeeper  types.LeaseKeeper
}

func NewVoteExtHandler(
	logger log.Logger, 
	orgK types.OrgKeeper,
    subnetK types.SubnetKeeper,
    leaseK types.LeaseKeeper,
	) *VoteExtHandler {
	return &VoteExtHandler{
		logger:       logger,
        OrgKeeper:    orgK,
        SubnetKeeper: subnetK,
        LeaseKeeper:  leaseK,
	}
}


type OracleVoteExtension struct {
	Height              int64
	OrgDomainValidation map[uint64]bool
	SubnetRirSync map[uint64]subnettypes.SubnetInfo
	RpkiValidation map[uint64]bool
	SubnetGeoValidation map[uint64]bool
	SubnetOwnerValidation map[uint64]bool
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// Get list of organizations that need validation
		orgsToValidate := h.OrgKeeper.GetOrgsForDomainValidation(ctx)
		subnetsToRirSync := h.SubnetKeeper.GetSubnetsForRirSync(ctx)
		subnetsForRPKIValidation := h.SubnetKeeper.GetSubnetsForRPKIValidation(ctx)
		subnetsForGeoValidation := h.SubnetKeeper.GetSubnetsForGeoValidation(ctx)
		subnetsForOwnershipVerification := h.SubnetKeeper.GetSubnetsForOwnershipVerification(ctx)

		if len(orgsToValidate) == 0 && len(subnetsToRirSync) == 0 && len(subnetsForRPKIValidation) == 0 && len(subnetsForGeoValidation) == 0 && len(subnetsForOwnershipVerification) == 0 {
			// Nothing to validate or sync, returning empty Vote
			return &abci.ResponseExtendVote{VoteExtension: []byte{}}, nil
		}
		h.currentBlock = req.Height
		h.logger.Info("performing validations for vote extension", 
		"height", req.Height, 
		"orgs_count", len(orgsToValidate),
		"subnets_rir_count", len(subnetsToRirSync),
		"subnets_rpki_count", len(subnetsForRPKIValidation),
		"subnets_geo_count", len(subnetsForGeoValidation))
		h.logger.Info("performing domain validations for vote extension", "height", req.Height)
		// Map to store validation results
		domainValidationResults := make(map[uint64]bool)
		subnetSyncResults := make(map[uint64]subnettypes.SubnetInfo)
		subnetRpkiValidationResults := make(map[uint64]bool)
		subnetGeoValidationResults := make(map[uint64]bool)
		subnetOwnershipVerifResults := make(map[uint64]bool)

		if len(orgsToValidate) > 0 {
			g := new(errgroup.Group)
			var mu sync.Mutex

			for _, org := range orgsToValidate {
				// Create local variables to avoid closure issues
				org := org

				g.Go(func() error {
					// Set a context with timeout for DNS lookup
					dnsCtx, cancel := context.WithTimeout(context.Background(), DNSLookupTimeout)
					defer cancel()

					// Perform DNS lookup
					isValid, err := h.validateDomain(dnsCtx, org.Id, org.Website)
					if err != nil {
						h.logger.Error("domain validation failed", "org_id", org.Id, "domain", org.Website, "error", err)
					}

					mu.Lock()
					domainValidationResults[org.Id] = isValid
					mu.Unlock()
					
					return nil
				})
			}

			// Wait for all DNS validations to complete
			if err := g.Wait(); err != nil {
				h.logger.Error("some domain validations failed", "error", err)
				// Continue even if some validations failed
			}

		}

		if len(subnetsForOwnershipVerification) > 0 {

			// Running the rDNS verification first as it can be executed concurrently
			g := new(errgroup.Group)
			var mu sync.Mutex

			for _, subnet := range subnetsForOwnershipVerification {
				// Create local variables to avoid closure issues
				subnet := subnet

				g.Go(func() error {
					// Set a context with timeout for DNS lookup
					rdnsCtx, cancel := context.WithTimeout(context.Background(), rDNSLookupTimeout)
					defer cancel()

					// Perform rDNS verification
					ip := convertDecimalIPToString(subnet.NetworkAddress+1)
					isValid, err := CheckSubnetRdnsVerification(rdnsCtx, ip, subnet.Id)
					if err != nil {
						h.logger.Error("subnet ownership verification through rDNS failed", "subnet_id", subnet.Id, "rDNS host", ip, "error", err)
					}

					mu.Lock()
					subnetOwnershipVerifResults[subnet.Id] = isValid
					mu.Unlock()
					
					return nil
				})
			}

			// Wait for all DNS validations to complete
			if err := g.Wait(); err != nil {
				h.logger.Error("Some subnet ownership verifications through rDNS failed", "error", err)
				// Continue even if some validations failed
			}
		}

		if len(subnetsToRirSync) > 0 {
			for _, subnet := range subnetsToRirSync {
				subnetDetails:= h.SyncSubnetData(subnet.NetworkAddress, subnet.Mask)
				subnetSyncResults[subnet.Id] = subnetDetails
			}
		}

		if len(subnetsForRPKIValidation) > 0 {
			for _, subnet := range subnetsForRPKIValidation {
				prefix :=  fmt.Sprintf("%s/%d", convertDecimalIPToString(subnet.NetworkAddress), subnet.Mask)
				isValid := h.IsASNValidForPrefix(prefix, subnet.OriginAsn)
				subnetRpkiValidationResults[subnet.Id] = isValid
			}
		}

		// This is the longest process. We only want to execute it if there's no other processes in this vote extension.
		if len(subnetsForGeoValidation) > 0 && len(orgsToValidate) == 0 && len(subnetsToRirSync) == 0 && len(subnetsForRPKIValidation) == 0 {
			subnetGeoValidationResults = h.VerifyGeoTrace(subnetsForGeoValidation)
		}


		// Create vote extension
		voteExt := OracleVoteExtension{
			Height:              req.Height,
			OrgDomainValidation: domainValidationResults,
			SubnetRirSync : subnetSyncResults,
			RpkiValidation : subnetRpkiValidationResults,
			SubnetGeoValidation: subnetGeoValidationResults,
		}

		// Log the results
		h.logger.Info("Vote extension complete")

		// Marshal the results
		bz, err := json.Marshal(voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// Helper method to validate a domain
func (h *VoteExtHandler) validateDomain(ctx context.Context, orgID uint64, website string) (bool, error) {
	// Extract domain from website URL
	domain, err := extractDomain(website)
	if err != nil {
		return false, fmt.Errorf("failed to extract domain from website %s: %w", website, err)
	}

	// Use net.DefaultResolver with context for timeout control
	resolver := &net.Resolver{}

	// Look up TXT records
	txtRecords, err := resolver.LookupTXT(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed for domain %s: %w", domain, err)
	}

	// Check for the expected TXT record format: "ipex=<orgID>"
	expectedValue := fmt.Sprintf("ipex=%d", orgID)

	for _, record := range txtRecords {
		if record == expectedValue {
			return true, nil
		}
	}

	// No matching record found
	return false, nil
}

// extractDomain parses a URL and returns just the domain part
func extractDomain(websiteURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}

	// Get the host part (domain)
	domain := parsedURL.Hostname()
	if domain == "" {
		return "", fmt.Errorf("empty domain")
	}

	return domain, nil
}

// SyncSubnetData looks up subnet information via RDAP for validation and metadata extraction.
// It queries the RDAP registry for information about the given IP/subnet and validates the following:
//
// 1. That the subnet fits within the allocated range returned by RDAP
// 2. Organization information (ID and name)
// 4. Valid abuse contact email
//
// Parameters:
//   - ipDecimal: The starting IP address of the subnet in decimal format
//   - mask: The subnet mask length (e.g., 24 for a /24 subnet)
//
// Returns:
//
//	A SubnetInfo struct containing:
//	- Valid: true if all validations pass, otherwise false
//	- OrgID: The organization identifier from RDAP
//	- OrgName: The organization name from RDAP
//	- AbuseEmail: Contact email for abuse reports
//	- AllocationType: The allocation type from RDAP (e.g., "ALLOCATED")
//	- RIR: corresponding Regional Internet Registry("RIPE", "ARIN")
//
// The function returns an invalid SubnetInfo (Valid=false) if any validation step fails
// or if required information cannot be extracted.
func (h *VoteExtHandler) SyncSubnetData(ipDecimal uint64, mask uint64) subnettypes.SubnetInfo {
	// Convert decimal IP to string format for RDAP query
	ipStr := convertDecimalIPToString(ipDecimal)
	// Prepare empty result
	result := subnettypes.SubnetInfo{
		Valid:          false,
		OrgID:          "",
		OrgName:        "",
		AbuseEmail:     "",
		AllocationType: "",
		Rir: "",
	}
	// Initialize RDAP client using the http client with timeout
	client := &rdap.Client{
		HTTP: &http.Client{
			Timeout: RDAPQueryTimeout,
		},
	}

	// Perform the RDAP query
	response, err := client.QueryIP(ipStr)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "No RDAP servers responded successfully") {
			h.logger.Info("Rate limit detected: %v\n", err)
		} else {
			h.logger.Info("RDAP query failed: %v\n", err)
		}
		return result
	}

	// Extract and validate RIR
	responseRIR := getRIRFromWhoisServer(response.Port43)

	// We only support ARIN and RIPE
	if responseRIR != "RIPE" && responseRIR != "ARIN" {
		return result
	}

	
	// Calculate end IP based on the input IP and mask
	endIp := calculateSubnetEndIP(ipDecimal, mask)

	// Get the inetnum end IP from the RDAP response
	inetnumEndIp, err := IpToDecimal(response.EndAddress)
	if err != nil {
		h.logger.Info("Error converting RDAP end address to decimal: %v\n", err)
		return result
	}

	// Checking that the subnet we've received is within the inetnum range
	if endIp > inetnumEndIp {
		h.logger.Info("Subnet exceeds the inetnum range")
		return result
	}

	// Extract allocation type
	result.AllocationType = response.Type

	// Extract registrant information
	result.OrgID, result.OrgName = extractOrgInfoFromEntities(response.Entities)

	// Extract abuse email
	result.AbuseEmail = findAbuseEmail(response.Entities)

	// Set the RIR 
	result.Rir = responseRIR

	// If no orgId, orgName or abuseEmail extracted
	if result.OrgID == "" || result.OrgName == "" || result.AbuseEmail == "" {
		return result
	}
	// Otherwise subnet is valid
	result.Valid = true
	// Return the validation result and extracted data
	return result
}

// getRIRFromWhoisServer maps a WHOIS server address to its corresponding RIR name.
// Returns an empty string if the server is not recognized.
func getRIRFromWhoisServer(port43 string) string {
	// Map of WHOIS servers to RIR names
	whoisToRIR := map[string]string{
		"whois.arin.net":    "ARIN",
		"whois.ripe.net":    "RIPE",
		"whois.apnic.net":   "APNIC",
		"whois.lacnic.net":  "LACNIC",
		"whois.afrinic.net": "AFRINIC",
	}

	// Look up the RIR name for the given WHOIS server
	return whoisToRIR[port43]
}

// extractEmailFromVCard extracts an email address from a VCard if present.
// Returns an empty string if no email is found or if vcard is nil.
func extractEmailFromVCard(vcard *rdap.VCard) string {
	// Handle nil VCard
	if vcard == nil {
		return ""
	}

	// Look for email property in VCard
	for _, prop := range vcard.Properties {
		if prop.Name == "email" {
			if value, ok := prop.Value.(string); ok {
				return value
			}
		}
	}

	// No email found
	return ""
}

// findAbuseEmail recursively searches for an abuse email in entities and their nested entities.
// Returns the first abuse email found, or an empty string if none is found.
func findAbuseEmail(entities []rdap.Entity) string {
	// Iterate through all entities at current level
	for _, entity := range entities {
		// Check if current entity has abuse role
		for _, role := range entity.Roles {
			if role == "abuse" {
				// Try to extract email from this entity
				email := extractEmailFromVCard(entity.VCard)
				if email != "" {
					return email
				}
			}
		}

		// If no email found and entity has nested entities, search those
		if len(entity.Entities) > 0 {
			// Recursive call to search nested entities
			email := findAbuseEmail(entity.Entities)
			if email != "" {
				return email
			}
		}
	}

	// No abuse email found in any entity
	return ""
}

// convertDecimalIPToString converts a uint64 decimal IP representation to a string
// in the standard dotted-decimal format (e.g., "192.168.1.1").
func convertDecimalIPToString(ipDecimal uint64) string {
	// Extract each octet from the decimal representation
	ipBytes := []byte{
		byte(ipDecimal >> 24),
		byte(ipDecimal >> 16),
		byte(ipDecimal >> 8),
		byte(ipDecimal),
	}

	// Format as a dotted-decimal string
	return fmt.Sprintf("%d.%d.%d.%d", ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3])
}

// calculateSubnetEndIP calculates the last IP address in a subnet
// given the start IP in decimal form and the subnet mask.
func calculateSubnetEndIP(ipDecimal uint64, mask uint64) uint64 {
	// Calculate number of host bits (32 - mask for IPv4)
	hostBits := uint64(32 - mask)

	// Create a mask with 1s in the host portion
	hostMask := (uint64(1) << hostBits) - 1

	// Calculate end IP by setting all host bits to 1
	// (This is equivalent to adding the maximum number of hosts to the start IP)
	return ipDecimal | hostMask
}

// extractOrgInfoFromEntities searches for a registrant entity in the provided entities list
// and extracts organization ID and name information.
// Returns empty strings if no registrant entity is found or if information is incomplete.
func extractOrgInfoFromEntities(entities []rdap.Entity) (orgID, orgName string) {
	// Default return values
	orgID = ""
	orgName = ""

	// Iterate through all entities
	for _, entity := range entities {
		// Check if entity has the registrant role
		for _, role := range entity.Roles {
			if role == "registrant" {
				// Extract organization ID from entity handle
				orgID = entity.Handle

				// Extract organization name from vCard
				if entity.VCard != nil {
					for _, prop := range entity.VCard.Properties {
						if prop.Name == "fn" {
							if value, ok := prop.Value.(string); ok {
								orgName = value
							}
						}
					}
				}

				// If we've found both pieces of information, return early
				if orgID != "" && orgName != "" {
					return orgID, orgName
				}
			}
		}
	}

	// Return whatever we found (may be incomplete or empty)
	return orgID, orgName
}

// Helper function to convert IP string to decimal
func IpToDecimal(ipStr string) (uint64, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("not an IPv4 address: %s", ipStr)
	}

	return uint64(ip[0])<<24 | uint64(ip[1])<<16 | uint64(ip[2])<<8 | uint64(ip[3]), nil
}

func (h *VoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		// If extension is empty, accept it (no validations needed)
		if len(req.VoteExtension) == 0 {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
		}

		// Unmarshal the vote extension
		var voteExt OracleVoteExtension
		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		// Verify the height matches
		if voteExt.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
		}

		// Verify the structure of the validation results
		if voteExt.OrgDomainValidation == nil && voteExt.SubnetRirSync == nil {
			return nil, fmt.Errorf("vote extension data is nil")
		}

		// Check that all org IDs are valid
		if voteExt.OrgDomainValidation != nil {
			for orgID := range voteExt.OrgDomainValidation {
				// Verify the org exists in state
				org, found := h.OrgKeeper.GetOrg(ctx, orgID)
				if !found {
					return nil, fmt.Errorf("org ID %d in vote extension does not exist", orgID)
				}

				// Check if the org has a website
				if org.Website == "" {
					return nil, fmt.Errorf("org ID %d has no website", orgID)
				}
			}
		}

		// Check that all subnet IDs are valid
		if voteExt.SubnetRirSync != nil {
			for subnetID := range voteExt.SubnetRirSync {
				// Verify the org exists in state
				_, found := h.SubnetKeeper.GetSubnet(ctx, subnetID)
				if !found {
					return nil, fmt.Errorf("subnet ID %d in vote extension does not exist", subnetID)
				}
			}
		}
		// Accept the vote extension
		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

// IsASNValidForPrefix checks if the given ASN is a valid origin for the specified prefix
func (h *VoteExtHandler) IsASNValidForPrefix(prefix string, asn uint64) bool {
	routinatorURLs := []string{
		"http://127.0.0.1:1337",           // Local Routinator
		"https://rpki-validator.ripe.net", // RIPE Routinator (fallback)
	}

	client := &http.Client{
		Timeout: RPKIQueryTimeout,
	}

	var resp *http.Response
	var err error

	// Try each Routinator URL until one succeeds
	for i, baseURL := range routinatorURLs {
		url := fmt.Sprintf("%s/api/v1/validity/%d/%s", baseURL, asn, prefix)

		resp, err = client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break // Successfully connected to this Routinator
		}

		if resp != nil {
			resp.Body.Close() // Close the body if we got a response but it wasn't OK
		}

		if i < len(routinatorURLs)-1 {
			// Log fallback message only if there are more URLs to try
			h.logger.Info("Cannot reach Routinator at %s: %v. Trying fallback.\n", baseURL, err)
		} else {
			// All URLs failed
			h.logger.Error("Error querying all Routinator instances: %v\n", err)
			return false
		}
	}

	// At this point we have a valid response
	defer resp.Body.Close()

	// Parse the validation response
	var validationResp struct {
		ValidatedRoute struct {
			Route struct {
				OriginASN string `json:"origin_asn"`
				Prefix    string `json:"prefix"`
			} `json:"route"`
			Validity struct {
				State       string `json:"state"`
				Description string `json:"description"`
				Reason      string `json:"reason,omitempty"`
			} `json:"validity"`
		} `json:"validated_route"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		h.logger.Error("Error decoding validation response: %v\n", err)
		return false
	}

	// Check if state is "valid"
	return validationResp.ValidatedRoute.Validity.State == "valid"
}

func Traceroute(ip string, startHop int, stopHop int) (lastIP string, err error) {
	// Prepare traceroute arguments
	args := []string{"-n", "-m", strconv.Itoa(stopHop), "-q", "1", "-w", "1", "-f", strconv.Itoa(startHop), ip}
	cmd := exec.Command("traceroute", args...)

	// Capture combined stdout and stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Execute the command
	if err = cmd.Run(); err != nil {
		return
	}

	// Scan all output lines
	scanner := bufio.NewScanner(&out)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip the header line
		if strings.HasPrefix(line, "traceroute to") {
			continue
		}

		// Get the last non-target IP
		for _, tok := range strings.Fields(line) {
			t := strings.Trim(tok, "(),")
			if curIp := net.ParseIP(t); curIp != nil && curIp.String() != ip {
				lastIP = curIp.String()
				break
			}
		}
	}

	return
}

func LookupGeo(ip string) string {
	// Path to the MaxMind database file
	dbPath := "data/maxmind/GeoLite2-Country.mmdb" // This is a Country database

	// Open the MaxMind database
	db, err := geoip.Open(dbPath)
	if err != nil {
		return ""
	}
	defer db.Close()

	// Parse the IP address
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return ""
	}

	// Use Country lookup instead of City for a Country database
	record, err := db.Country(ipAddr)
	if err != nil {
		return ""
	}

	if record.Country.IsoCode != "" {
		return strings.ToUpper(record.Country.IsoCode) // Convert to uppercase
	}

	return ""
}

// VerifyGeoTrace performs geo validation on subnets by running traceroutes
// and checking if the geographic location of the last hop matches the expected location
func (h *VoteExtHandler) VerifyGeoTrace(subnets []subnettypes.Subnet) map[uint64]bool {
    // Set timeout context for the entire operation
    ctx, cancel := context.WithTimeout(context.Background(), TracerouteTimeout)
    defer cancel()
    
    // Record start time for total execution
    totalStart := time.Now()
    
    // Initialize results map to store validation outcomes by subnet ID
    results := make(map[uint64]bool)
    
    // Channel for receiving and ordering trace results
    type traceResult struct {
        subnetId        uint64
        country         string  // Store the country for comparison when processing results
        lastIp1_3       string
        lastIp3_5       string
        lastIp5_7       string
        lastIp7_9       string
        lastIp9_11      string
        lastIp11_13     string
        lastIp13_15     string
        lastIp15_17     string
        lastIp17_19     string
        lastIp19_21     string
        err             error
    }
    
    resultChan := make(chan traceResult, len(subnets))
    
    var wg sync.WaitGroup
    // Limit concurrent subnet processing (adjust based on system capacity)
    semaphore := make(chan struct{}, 20)
    
    h.logger.Info("Starting geo validation via traceroutes", "subnets_count", len(subnets))
    
    // Start the trace for each subnet in a goroutine
    for _, subnet := range subnets {
        select {
        case <-ctx.Done():
            h.logger.Info("Geo validation timeout reached, processing results so far")
            goto ProcessResults
        default:
            // Continue processing
        }
        
        wg.Add(1)
        
        go func(s subnettypes.Subnet) {
            // Acquire a spot in the semaphore
            semaphore <- struct{}{}
            defer func() { <-semaphore }() // Release when done
            defer wg.Done()
            
            // Calculate target IP (.1 in the subnet)
            targetIp := s.NetworkAddress + 1
            ipStr := convertDecimalIPToString(targetIp)
            
            // Run all traceroutes in parallel with their own wait group
            var traceWg sync.WaitGroup
            var lastIp1_3, lastIp3_5, lastIp5_7, lastIp7_9, lastIp9_11, lastIp11_13, lastIp13_15, lastIp15_17, lastIp17_19, lastIp19_21 string
            var err1, err2, err3, err4, err5, err6, err7, err8, err9, err10 error
            
            traceWg.Add(10)
            
            // Run traceroute 1-3 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp1_3, err1 = Traceroute(ipStr, 1, 3)
            }()
            
            // Run traceroute 3-5 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp3_5, err2 = Traceroute(ipStr, 3, 5)
            }()
            
            // Run traceroute 5-7 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp5_7, err3 = Traceroute(ipStr, 5, 7)
            }()
            
            // Run traceroute 7-9 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp7_9, err4 = Traceroute(ipStr, 7, 9)
            }()
            
            // Run traceroute 9-11 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp9_11, err5 = Traceroute(ipStr, 9, 11)
            }()
            
            // Run traceroute 11-13 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp11_13, err6 = Traceroute(ipStr, 11, 13)
            }()
            
            // Run traceroute 13-15 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp13_15, err7 = Traceroute(ipStr, 13, 15)
            }()
            
            // Run traceroute 15-17 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp15_17, err8 = Traceroute(ipStr, 15, 17)
            }()
            
            // Run traceroute 17-19 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp17_19, err9 = Traceroute(ipStr, 17, 19)
            }()
            
            // Run traceroute 19-21 (2 hops)
            go func() {
                defer traceWg.Done()
                lastIp19_21, err10 = Traceroute(ipStr, 19, 21)
            }()
            
            // Wait for all traceroutes to complete with timeout
            traceDone := make(chan struct{})
            go func() {
                traceWg.Wait()
                close(traceDone)
            }()
            
            select {
            case <-traceDone:
                // All traces completed normally
            case <-time.After(5 * time.Second):
                // Timeout occurred, proceed with what we have
                h.logger.Info("Traceroute timeout reached for subnet", "subnet_id", s.Id)
            }
            
            // Determine which error to report - prioritize reporting errors only if all failed
            var resultErr error
            if err1 != nil && err2 != nil && err3 != nil && err4 != nil &&
               err5 != nil && err6 != nil && err7 != nil && err8 != nil &&
               err9 != nil && err10 != nil {
                resultErr = fmt.Errorf("all traceroutes failed for subnet %d", s.Id)
            }
            
            // Send result back through channel
            resultChan <- traceResult{
                subnetId:    s.Id,
                country:     s.Country,  // Store the country for later comparison
                lastIp1_3:   lastIp1_3,
                lastIp3_5:   lastIp3_5,
                lastIp5_7:   lastIp5_7,
                lastIp7_9:   lastIp7_9,
                lastIp9_11:  lastIp9_11,
                lastIp11_13: lastIp11_13,
                lastIp13_15: lastIp13_15,
                lastIp15_17: lastIp15_17,
                lastIp17_19: lastIp17_19,
                lastIp19_21: lastIp19_21,
                err:         resultErr,
            }
        }(subnet)
    }
    
    // Wait in a separate goroutine and close the channel when done
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
ProcessResults:
    // Process results as they come in
    var matchCount, totalCount int
    
    for res := range resultChan {
        // If all traceroutes failed, mark subnet as not validated
        if res.err != nil {
            h.logger.Info("All traceroutes failed for subnet", "subnet_id", res.subnetId)
            results[res.subnetId] = false
            continue
        }
        
        // Prioritize farthest traceroute result that has an IP
        var lastIp string
        
        // Check traceroutes in reverse order (farthest first)
        if res.lastIp19_21 != "" {
            lastIp = res.lastIp19_21
        } else if res.lastIp17_19 != "" {
            lastIp = res.lastIp17_19
        } else if res.lastIp15_17 != "" {
            lastIp = res.lastIp15_17
        } else if res.lastIp13_15 != "" {
            lastIp = res.lastIp13_15
        } else if res.lastIp11_13 != "" {
            lastIp = res.lastIp11_13
        } else if res.lastIp9_11 != "" {
            lastIp = res.lastIp9_11
        } else if res.lastIp7_9 != "" {
            lastIp = res.lastIp7_9
        } else if res.lastIp5_7 != "" {
            lastIp = res.lastIp5_7
        } else if res.lastIp3_5 != "" {
            lastIp = res.lastIp3_5
        } else if res.lastIp1_3 != "" {
            lastIp = res.lastIp1_3
        }
        
        if lastIp == "" {
            h.logger.Info("No valid last hop found for subnet", "subnet_id", res.subnetId)
            results[res.subnetId] = false
            continue
        }
        
        // Look up geo location of the last hop
        lastHopGeo := LookupGeo(lastIp)
        
        // Check if geo location matches the expected subnet geo location
        isGeoMatched := (lastHopGeo == res.country)
        
        h.logger.Info("Geo validation result", 
            "subnet_id", res.subnetId,
            "expected_geo", res.country,
            "last_hop_ip", lastIp,
            "last_hop_geo", lastHopGeo,
            "match", isGeoMatched)
        
        // Update results
        results[res.subnetId] = isGeoMatched
        totalCount++
        
        if isGeoMatched {
            matchCount++
        }
    }
    
    if totalCount > 0 {
        h.logger.Info("Geo validation statistics",
            "total_verified", totalCount,
            "matches", matchCount,
            "match_percent", float64(matchCount)/float64(totalCount)*100)
    }
    
    h.logger.Info("Geo validation completed", "duration_seconds", time.Since(totalStart).Seconds())
    
    return results
}

// CheckSubnetRdnsVerification takes an IPv4 address and an ID, performs a reverse
// DNS lookup, and checks if a TXT record with "ipex=$id" exists on the
// second and third level domains of the resolved hostname.
func CheckSubnetRdnsVerification(ctx context.Context, ipv4 string, id uint64) (bool, error) {
	// Validate the IPv4 address
	ip := net.ParseIP(ipv4)
	if ip == nil || strings.Count(ipv4, ".") != 3 {
		return false, fmt.Errorf("invalid IPv4 address: %s", ipv4)
	}
	// Use net.DefaultResolver with context for timeout control
	resolver := &net.Resolver{}

	// Perform the reverse DNS lookup
	names, err := resolver.LookupAddr(ctx, ipv4)
	if err != nil {
		return false, fmt.Errorf("reverse DNS lookup failed: %w", err)
	}

	// If no results found
	if len(names) == 0 {
		return false, fmt.Errorf("no reverse DNS records found for %s", ipv4)
	}

	// Get the domain name (remove trailing dot if present)
	domain := strings.TrimSuffix(names[0], ".")

	// Extract second and third level domains
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false, fmt.Errorf("domain %s doesn't have enough levels", domain)
	}

	// Determine second and third level domains
	var secondLevelDomain, thirdLevelDomain string

	if len(parts) >= 3 {
		// For domains like sub.example.com:
		// Third level: sub.example.com
		// Second level: example.com
		secondLevelDomain = strings.Join(parts[len(parts)-2:], ".")
		thirdLevelDomain = domain
	} else {
		// For domains like example.com:
		// Second level: example.com
		// No third level
		secondLevelDomain = domain
		thirdLevelDomain = "" // No third level available
	}

	// Construct the expected TXT record value
	expectedTXT := "ipex=" + strconv.FormatUint(id, 10)

	// Check TXT records on the second level domain
	secondLevelTXTFound := false
	txtRecordsSecond, err := resolver.LookupTXT(ctx, secondLevelDomain)
	if err == nil {
		// Check if any of the TXT records match our expected value
		for _, txt := range txtRecordsSecond {
			if txt == expectedTXT {
				secondLevelTXTFound = true
				break
			}
		}
	} else {
		// If the error is not "no such host" or similar, return it
		if dnsErr, ok := err.(*net.DNSError); !ok || !dnsErr.IsNotFound {
			return false, fmt.Errorf("error checking TXT record for %s: %w", secondLevelDomain, err)
		}
	}

	// If no third level domain or we already found the TXT record on second level
	if thirdLevelDomain == "" || secondLevelTXTFound {
		return secondLevelTXTFound, nil
	}

	// Check TXT records on the third level domain
	thirdLevelTXTFound := false
	txtRecordsThird, err := resolver.LookupTXT(ctx, thirdLevelDomain)
	if err == nil {
		// Check if any of the TXT records match our expected value
		for _, txt := range txtRecordsThird {
			if txt == expectedTXT {
				thirdLevelTXTFound = true
				break
			}
		}
	} else {
		// If the error is not "no such host" or similar, return it
		if dnsErr, ok := err.(*net.DNSError); !ok || !dnsErr.IsNotFound {
			return false, fmt.Errorf("error checking TXT record for %s: %w", thirdLevelDomain, err)
		}
	}

	// Return true if either domain has the TXT record
	return secondLevelTXTFound || thirdLevelTXTFound, nil
}