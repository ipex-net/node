package abci

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"cosmossdk.io/log"
	"golang.org/x/sync/errgroup"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/org/keeper"
)


type VoteExtHandler struct {
    logger           log.Logger
    currentBlock     int64                  // current block height
    lastCheckTime    time.Time              // last time we performed domain validations
    checkTimeout     time.Duration          // timeout for DNS lookups
    
    Keeper           keeper.Keeper          // keeper of the org module itself
}

func NewVoteExtHandler(logger log.Logger, timeout time.Duration, k keeper.Keeper) *VoteExtHandler {
    return &VoteExtHandler{
        logger:       logger,
        checkTimeout: timeout,
        Keeper:       k,
    }
}

type OracleVoteExtension struct {
	Height int64
	OrgDomainValidation map[uint64]bool
   }


func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
    return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
        h.currentBlock = req.Height
        h.lastCheckTime = time.Now()
        h.logger.Info("performing domain validations for vote extension", "height", req.Height, "time", h.lastCheckTime)
        
        // Get list of organizations that need validation
        orgsToValidate := h.Keeper.GetOrgsForDomainValidation(ctx)
        
        if len(orgsToValidate) == 0 {
            // No orgs need validation, return empty extension
            return &abci.ResponseExtendVote{VoteExtension: []byte{}}, nil
        }
        
        // Map to store validation results
        validationResults := make(map[uint64]bool)
        
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
                    // Continue with other validations, don't fail completely
                    return nil
                }
                
                // Safely update the results map
                mu.Lock()
                validationResults[org.Id] = isValid
                mu.Unlock()
                
                return nil
            })
        }
        
        // Wait for all validations to complete
        if err := g.Wait(); err != nil {
            h.logger.Error("some domain validations failed", "error", err)
            // Continue even if some validations failed
        }
        
        // Create vote extension
        voteExt := OracleVoteExtension{
            Height: req.Height,
            OrgDomainValidation: validationResults,
        }
        
        // Log the results
        h.logger.Info("domain validation complete", "validated_count", len(validationResults))
        
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
        if voteExt.OrgDomainValidation == nil {
            return nil, fmt.Errorf("org domain validation map is nil")
        }
        
        // Check that all org IDs are valid
        for orgID := range voteExt.OrgDomainValidation {
            // Verify the org exists in state
            org, found := h.Keeper.GetOrg(ctx, orgID)
            if !found {
                return nil, fmt.Errorf("org ID %d in vote extension does not exist", orgID)
            }
            
            // Check if the org has a website
            if org.Website == "" {
                return nil, fmt.Errorf("org ID %d has no website", orgID)
            }
        }
        
        // Accept the vote extension
        return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
    }
}