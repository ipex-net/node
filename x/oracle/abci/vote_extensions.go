package abci

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"cosmossdk.io/log"
	"github.com/openrdap/rdap"
	"golang.org/x/sync/errgroup"

	"ipex/x/oracle/types"
	subnettypes "ipex/x/subnet/types"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VoteExtHandler struct {
	logger        log.Logger
	currentBlock  int64         // current block height
	lastCheckTime time.Time     // last time we performed domain validations
	checkTimeout  time.Duration // timeout for lookups

    OrgKeeper    types.OrgKeeper
    SubnetKeeper types.SubnetKeeper
    LeaseKeeper  types.LeaseKeeper
}

func NewVoteExtHandler(
	logger log.Logger, 
	timeout time.Duration, 
	orgK types.OrgKeeper,
    subnetK types.SubnetKeeper,
    leaseK types.LeaseKeeper,
	) *VoteExtHandler {
	return &VoteExtHandler{
		logger:       logger,
        checkTimeout: timeout,
        OrgKeeper:    orgK,
        SubnetKeeper: subnetK,
        LeaseKeeper:  leaseK,
	}
}


type OracleVoteExtension struct {
	Height              int64
	OrgDomainValidation map[uint64]bool
	SubnetRirSync map[uint64]subnettypes.SubnetInfo
	AsnPrefixValidation map[uint64]bool
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// Get list of organizations that need validation
		orgsToValidate := h.OrgKeeper.GetOrgsForDomainValidation(ctx)
		subnetsToRirSync := h.SubnetKeeper.GetSubnetsForRirSync(ctx)
		subnetsForRPKIValidation := h.SubnetKeeper.GetSubnetsForRPKIValidation(ctx)

		if len(orgsToValidate) == 0 && len(subnetsToRirSync) == 0 && len(subnetsForRPKIValidation) == 0 {
			// Nothing to validate or sync, returning empty Vote
			return &abci.ResponseExtendVote{VoteExtension: []byte{}}, nil
		}
		h.currentBlock = req.Height
		h.lastCheckTime = time.Now()
		h.logger.Info("performing domain validations for vote extension", "height", req.Height, "time", h.lastCheckTime)
		// Map to store validation results
		domainValidationResults := make(map[uint64]bool)
		subnetSyncResults := make(map[uint64]subnettypes.SubnetInfo)
		subnetRPKIValidationResults := make(map[uint64]bool)

		if len(orgsToValidate) > 0 {
			// Use errgroup to perform DNS lookups concurrently
			g := new(errgroup.Group)

			// Mutex to protect concurrent map access
			var mu sync.Mutex

			for _, org := range orgsToValidate {
				// Create local variables to avoid closure issues
				org := org

				g.Go(func() error {
					// Set a context with timeout for DNS lookup
					dnsCtx, cancel := context.WithTimeout(context.Background(), h.checkTimeout)
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
				subnetRPKIValidationResults[subnet.Id] = isValid
			}
		}


		// Create vote extension
		voteExt := OracleVoteExtension{
			Height:              req.Height,
			OrgDomainValidation: domainValidationResults,
			SubnetRirSync : subnetSyncResults,
			AsnPrefixValidation : subnetRPKIValidationResults,
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
	// Initialize RDAP client
	client := &rdap.Client{}

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
		Timeout: time.Second * 5,
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