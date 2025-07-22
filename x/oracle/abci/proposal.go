package abci

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/oracle/types"
	subnettypes "ipex/x/subnet/types"
	trusttypes "ipex/x/trust/types"
)

// NetworkValidationResults defines the consolidated results of all network resource validations
// performed by validators and aggregated by stake-weighted voting power. This structure
// contains the final validation outcomes for organizations, subnets, and their associated
// properties, along with the extended commit info used to calculate these results.

type NetworkValidationResults struct {
	OrgValidationResults       map[uint64]bool
	SubnetSyncResults          map[uint64]subnettypes.SubnetInfo
	RpkiValidationResults      map[uint64]bool
	GeoValidationResults       map[uint64]bool
	SubnetOwnerValidation      map[uint64]bool
	SubnetBlacklistSyncResults map[uint64]trusttypes.SubnetBlacklistCount
	SubnetOriginCheckResults   map[uint64]bool
	ExtendedCommitInfo         abci.ExtendedCommitInfo
}

type ProposalHandler struct {
	logger       log.Logger
	orgKeeper    types.OrgKeeper
	subnetKeeper types.SubnetKeeper
	leaseKeeper  types.LeaseKeeper
	trustKeeper  types.TrustKeeper
	valStore     baseapp.ValidatorStore
}

func NewProposalHandler(
	logger log.Logger,
	orgK types.OrgKeeper,
	subnetK types.SubnetKeeper,
	leaseK types.LeaseKeeper,
	trustK types.TrustKeeper,
	valStore baseapp.ValidatorStore,
) *ProposalHandler {
	return &ProposalHandler{
		logger:       logger,
		orgKeeper:    orgK,
		subnetKeeper: subnetK,
		leaseKeeper:  leaseK,
		trustKeeper:  trustK,
		valStore:     valStore,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		proposalTxs := req.Txs

		//Only execute if vote extensions are enabled
		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			// Attempt to validate vote extensions, but do not abort proposal on error
			if err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit); err != nil {
				h.logger.Error("failed to validate vote extensions; injecting empty validation results", "err", err)
			}

			// Compute consensus on validation results (produce empty maps on error)
			orgValidationResults, subnetSyncResults, rpkiValidationResults, geoValidationResults, subnetOwnershipResults, subnetBlacklistSyncResults, subnetOriginCheckResults, err := h.computeValidationResults(ctx, req.LocalLastCommit)
			if err != nil {
				h.logger.Error("failed to compute validation results; injecting empty maps", "err", err)
				orgValidationResults = make(map[uint64]bool)
				subnetSyncResults = make(map[uint64]subnettypes.SubnetInfo)
				rpkiValidationResults = make(map[uint64]bool)
				geoValidationResults = make(map[uint64]bool)
				subnetOwnershipResults = make(map[uint64]bool)
				subnetBlacklistSyncResults = make(map[uint64]trusttypes.SubnetBlacklistCount)
				subnetOriginCheckResults = make(map[uint64]bool)
			}

			// Create the injected transaction with validation results
			injectedVoteExtTx := NetworkValidationResults{
				OrgValidationResults:       orgValidationResults,
				SubnetSyncResults:          subnetSyncResults,
				RpkiValidationResults:      rpkiValidationResults,
				GeoValidationResults:       geoValidationResults,
				SubnetOwnerValidation:      subnetOwnershipResults,
				SubnetBlacklistSyncResults: subnetBlacklistSyncResults,
				SubnetOriginCheckResults:   subnetOriginCheckResults,
				ExtendedCommitInfo:         req.LocalLastCommit,
			}

			// Marshal to JSON for injection
			bz, err := json.Marshal(injectedVoteExtTx)
			if err != nil {
				h.logger.Error("failed to encode injected vote extension tx; skipping injection", "err", err)
			} else {
				// Inject as first transaction in the block
				proposalTxs = append([][]byte{bz}, proposalTxs...)
			}
		}

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

// computeValidationResults aggregates validation results from vote extensions
func (h *ProposalHandler) computeValidationResults(_ sdk.Context, ci abci.ExtendedCommitInfo) (map[uint64]bool, map[uint64]subnettypes.SubnetInfo, map[uint64]bool, map[uint64]bool, map[uint64]bool, map[uint64]trusttypes.SubnetBlacklistCount, map[uint64]bool, error) {
	// Maps to store aggregated validation results
	orgValidationResults := make(map[uint64]bool)
	subnetSyncResults := make(map[uint64]subnettypes.SubnetInfo)
	rpkiValidationResults := make(map[uint64]bool)
	geoValidationResults := make(map[uint64]bool)
	subnetOwnershipResults := make(map[uint64]bool)
	subnetBlacklistSyncResults := make(map[uint64]trusttypes.SubnetBlacklistCount)
	subnetOriginCheckResults :=make(map[uint64]bool)

	// Maps to track voting power for org validations
	orgValidPower := make(map[uint64]int64)
	orgInvalidPower := make(map[uint64]int64)

	// Maps to track voting power for RPKI validations
	rpkiValidPower := make(map[uint64]int64)
	rpkiInvalidPower := make(map[uint64]int64)

	// Maps to track voting power for TraceGeo validations
	geoValidPower := make(map[uint64]int64)
	geoInvalidPower := make(map[uint64]int64)

	// Maps to track voting power for subnet ownership validation
	subnetOwnerValidPower := make(map[uint64]int64)
	subnetOwnerInvalidPower := make(map[uint64]int64)

	// Map to track voting power for subnet info by subnet ID and unique SubnetInfo
	// Using string representation of SubnetInfo for map key to handle struct comparison
	type subnetVotePower struct {
		subnetID uint64
		infoJSON string
		power    int64
		info     subnettypes.SubnetInfo
	}
	// List to store subnet voting power entries
	subnetVotePowers := []subnetVotePower{}

	type blacklistVotePower struct {
		subnetID  uint64
		countJSON string
		power     int64
		count     trusttypes.SubnetBlacklistCount
	}
	blacklistVotePowers := []blacklistVotePower{}

	originValidPower := make(map[uint64]int64)
	originInvalidPower := make(map[uint64]int64)

	// Process each vote extension
	for _, v := range ci.Votes {
		// Skip votes that don't commit
		if v.BlockIdFlag != 2 { // BlockIDFlagCommit
			continue
		}

		// Skip empty vote extensions
		if len(v.VoteExtension) == 0 {
			continue
		}

		// Parse the vote extension
		var voteExt OracleVoteExtension
		if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
			h.logger.Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
			continue // Skip this vote extension but process others
		}

		// Aggregate org validation results based on voting power
		for orgID, isValid := range voteExt.OrgDomainValidation {
			if isValid {
				orgValidPower[orgID] += v.Validator.Power
			} else {
				orgInvalidPower[orgID] += v.Validator.Power
			}
		}

		// Aggregate RPKI validation results based on voting power
		for subnetID, isValid := range voteExt.RpkiValidation {
			if isValid {
				rpkiValidPower[subnetID] += v.Validator.Power
			} else {
				rpkiInvalidPower[subnetID] += v.Validator.Power
			}
		}

		// Aggregate subnet sync results
		for subnetID, subnetInfo := range voteExt.SubnetRirSync {
			// Convert the SubnetInfo to JSON for consistent comparison
			infoJSON, err := json.Marshal(subnetInfo)
			if err != nil {
				h.logger.Error("failed to marshal subnet info", "err", err, "subnet_id", subnetID)
				continue
			}

			// Add to subnet vote powers list
			subnetVotePowers = append(subnetVotePowers, subnetVotePower{
				subnetID: subnetID,
				infoJSON: string(infoJSON),
				power:    v.Validator.Power,
				info:     subnetInfo,
			})
		}

		// Aggregate geo validation results based on voting power
		for subnetID, isValid := range voteExt.SubnetGeoValidation {
			if isValid {
				geoValidPower[subnetID] += v.Validator.Power
			} else {
				geoInvalidPower[subnetID] += v.Validator.Power
			}
		}

		// Aggregate subnet ownership verification results based on voting power
		for subnetID, isValid := range voteExt.SubnetOwnerValidation {
			if isValid {
				subnetOwnerValidPower[subnetID] += v.Validator.Power
			} else {
				subnetOwnerInvalidPower[subnetID] += v.Validator.Power
			}
		}

		// Aggregate blacklist sync results based on voting power
		for subnetID, blacklistCount := range voteExt.SubnetBlacklistSync {
			// Convert the SubnetBlacklistCount to JSON for consistent comparison
			countJSON, err := json.Marshal(blacklistCount)
			if err != nil {
				h.logger.Error("failed to marshal blacklist count", "err", err, "subnet_id", subnetID)
				continue
			}

			// Add to blacklist vote powers list
			blacklistVotePowers = append(blacklistVotePowers, blacklistVotePower{
				subnetID:  subnetID,
				countJSON: string(countJSON),
				power:     v.Validator.Power,
				count:     blacklistCount,
			})
		}

		// Aggregate subnet origin check results based on voting power:
		for subnetID, isValid := range voteExt.SubnetOriginCheck {
			if isValid {
				originValidPower[subnetID] += v.Validator.Power
			} else {
				originInvalidPower[subnetID] += v.Validator.Power
			}
		}
	}

	// Build the union of all org IDs that appeared in any vote extension
	allOrgIDs := make(map[uint64]struct{})
	for orgID := range orgValidPower {
		allOrgIDs[orgID] = struct{}{}
	}
	for orgID := range orgInvalidPower {
		allOrgIDs[orgID] = struct{}{}
	}

	// Determine final org validation result based on majority voting power
	for orgID := range allOrgIDs {
		vp := orgValidPower[orgID]
		ip := orgInvalidPower[orgID]
		total := vp + ip
		if total > 0 && vp > total/2 {
			orgValidationResults[orgID] = true
		} else {
			orgValidationResults[orgID] = false
		}
	}

	// Build the union of all subnet IDs that appeared in RPKI validation
	allRpkiSubnetIDs := make(map[uint64]struct{})
	for subnetID := range rpkiValidPower {
		allRpkiSubnetIDs[subnetID] = struct{}{}
	}
	for subnetID := range rpkiInvalidPower {
		allRpkiSubnetIDs[subnetID] = struct{}{}
	}

	// Determine final RPKI validation result based on majority voting power
	for subnetID := range allRpkiSubnetIDs {
		vp := rpkiValidPower[subnetID]
		ip := rpkiInvalidPower[subnetID]
		total := vp + ip
		if total > 0 && vp > total/2 {
			rpkiValidationResults[subnetID] = true
		} else {
			rpkiValidationResults[subnetID] = false
		}
	}

	// Build the union of all subnet IDs that appeared in Geo validation
	allGeoSubnetIDs := make(map[uint64]struct{})
	for subnetID := range geoValidPower {
		allGeoSubnetIDs[subnetID] = struct{}{}
	}
	for subnetID := range geoInvalidPower {
		allGeoSubnetIDs[subnetID] = struct{}{}
	}

	// Determine final Geo validation result based on majority voting power
	for subnetID := range allGeoSubnetIDs {
		vp := geoValidPower[subnetID]
		ip := geoInvalidPower[subnetID]
		total := vp + ip
		if total > 0 && vp > total/2 {
			geoValidationResults[subnetID] = true
		} else {
			geoValidationResults[subnetID] = false
		}
	}

	// Build the union of all subnet IDs that appeared in ownership verification
	allOwnershipSubnetIDs := make(map[uint64]struct{})
	for subnetID := range subnetOwnerValidPower {
		allOwnershipSubnetIDs[subnetID] = struct{}{}
	}
	for subnetID := range subnetOwnerInvalidPower {
		allOwnershipSubnetIDs[subnetID] = struct{}{}
	}

	// Group subnet voting powers by subnet ID
	subnetInfoVotingPower := make(map[uint64]map[string]int64)
	subnetInfoMap := make(map[uint64]map[string]subnettypes.SubnetInfo)

	for _, sv := range subnetVotePowers {
		// Initialize maps for this subnet ID if they don't exist
		if _, exists := subnetInfoVotingPower[sv.subnetID]; !exists {
			subnetInfoVotingPower[sv.subnetID] = make(map[string]int64)
			subnetInfoMap[sv.subnetID] = make(map[string]subnettypes.SubnetInfo)
		}

		// Add voting power for this subnet info
		subnetInfoVotingPower[sv.subnetID][sv.infoJSON] += sv.power
		subnetInfoMap[sv.subnetID][sv.infoJSON] = sv.info
	}

	// For each subnet ID, find the SubnetInfo with the highest voting power
	for subnetID, infoPowers := range subnetInfoVotingPower {
		var maxPower int64
		var maxInfoJSON string

		// Find the SubnetInfo with the highest voting power
		for infoJSON, power := range infoPowers {
			if power > maxPower {
				maxPower = power
				maxInfoJSON = infoJSON
			}
		}

		// Only include subnets where at least one validator provided info
		if maxPower > 0 {
			subnetSyncResults[subnetID] = subnetInfoMap[subnetID][maxInfoJSON]
		}
	}

	// Determine final ownership verification result based on majority voting power
	for subnetID := range allOwnershipSubnetIDs {
		vp := subnetOwnerValidPower[subnetID]
		ip := subnetOwnerInvalidPower[subnetID]
		total := vp + ip
		if total > 0 && vp > total/2 {
			subnetOwnershipResults[subnetID] = true
		} else {
			subnetOwnershipResults[subnetID] = false
		}
	}

	// Group blacklist voting powers by subnet ID
	blacklistInfoVotingPower := make(map[uint64]map[string]int64)
	blacklistInfoMap := make(map[uint64]map[string]trusttypes.SubnetBlacklistCount)

	for _, bv := range blacklistVotePowers {
		// Initialize maps for this subnet ID if they don't exist
		if _, exists := blacklistInfoVotingPower[bv.subnetID]; !exists {
			blacklistInfoVotingPower[bv.subnetID] = make(map[string]int64)
			blacklistInfoMap[bv.subnetID] = make(map[string]trusttypes.SubnetBlacklistCount)
		}

		// Add voting power for this blacklist count
		blacklistInfoVotingPower[bv.subnetID][bv.countJSON] += bv.power
		blacklistInfoMap[bv.subnetID][bv.countJSON] = bv.count
	}

	// For each subnet ID, find the SubnetBlacklistCount with the highest voting power
	for subnetID, countPowers := range blacklistInfoVotingPower {
		var maxPower int64
		var maxCountJSON string

		// Find the SubnetBlacklistCount with the highest voting power
		for countJSON, power := range countPowers {
			if power > maxPower {
				maxPower = power
				maxCountJSON = countJSON
			}
		}

		// Only include subnets where at least one validator provided blacklist data
		if maxPower > 0 {
			subnetBlacklistSyncResults[subnetID] = blacklistInfoMap[subnetID][maxCountJSON]
		}
	}

	// Build the union of all subnet IDs that appeared in origin check
	allOriginCheckSubnetIDs := make(map[uint64]struct{})
	for subnetID := range originValidPower {
		allOriginCheckSubnetIDs[subnetID] = struct{}{}
	}
	for subnetID := range originInvalidPower {
		allOriginCheckSubnetIDs[subnetID] = struct{}{}
	}

	// Determine final origin check result based on majority voting power
	for subnetID := range allOriginCheckSubnetIDs {
		vp := originValidPower[subnetID]
		ip := originInvalidPower[subnetID]
		total := vp + ip
		if total > 0 && vp > total/2 {
			subnetOriginCheckResults[subnetID] = true
		} else {
			subnetOriginCheckResults[subnetID] = false
		}
	}
	return orgValidationResults, subnetSyncResults, rpkiValidationResults, geoValidationResults, subnetOwnershipResults, subnetBlacklistSyncResults, subnetOriginCheckResults, nil
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// If there are no transactions, accept the proposal
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		// Skip validation if vote extensions aren't enabled yet
		if req.Height <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		// Unmarshal the injected vote extension transaction
		var injectedVoteExtTx NetworkValidationResults
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			h.logger.Error("failed to decode injected vote extension tx; skipping validation", "err", err)
			// Accept proposal, skip validation verification
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		// Validate the vote extensions
		err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			h.logger.Error("failed to validate vote extensions", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Verify the proposer's validation results by computing the same calculation
		orgValidationResults, subnetSyncResults, rpkiValidationResults, geoValidationResults, subnetOwnershipResults, subnetBlacklistSyncResults, originCheckResults, err := h.computeValidationResults(ctx, injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			h.logger.Error("failed to compute validation results", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's organization validation results with our own calculation
		if err := compareOrgValidationResults(injectedVoteExtTx.OrgValidationResults, orgValidationResults); err != nil {
			h.logger.Error("organization validation results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's RPKI validation results with our own calculation
		if err := compareRpkiValidationResults(injectedVoteExtTx.RpkiValidationResults, rpkiValidationResults); err != nil {
			h.logger.Error("RPKI validation results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's origin check results with our own calculation
		if err := compareOriginCheckResults(injectedVoteExtTx.SubnetOriginCheckResults, originCheckResults); err != nil {
			h.logger.Error("Origin check results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's subnet sync results with our own calculation
		if err := compareSubnetSyncResults(injectedVoteExtTx.SubnetSyncResults, subnetSyncResults); err != nil {
			h.logger.Error("subnet sync results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's geo validation results with our own calculation
		if err := compareGeoValidationResults(injectedVoteExtTx.GeoValidationResults, geoValidationResults); err != nil {
			h.logger.Error("geo validation results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's subnet ownership verification results with our own calculation
		if err := compareSubnetOwnershipResults(injectedVoteExtTx.SubnetOwnerValidation, subnetOwnershipResults); err != nil {
			h.logger.Error("subnet ownership verification results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's blacklist sync results with our own calculation
		if err := compareBlacklistSyncResults(injectedVoteExtTx.SubnetBlacklistSyncResults, subnetBlacklistSyncResults); err != nil {
			h.logger.Error("blacklist sync results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// compareOrgValidationResults checks if two organization validation result maps match
func compareOrgValidationResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("organization validation result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for orgID, valid1 := range r1 {
		valid2, exists := r2[orgID]
		if !exists {
			return fmt.Errorf("organization ID %d missing from second result map", orgID)
		}
		if valid1 != valid2 {
			return fmt.Errorf("validation result mismatch for organization ID %d: %t vs %t", orgID, valid1, valid2)
		}
	}

	return nil
}

// compareRpkiValidationResults checks if two RPKI validation result maps match
func compareRpkiValidationResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("RPKI validation result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, valid1 := range r1 {
		valid2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second result map", subnetID)
		}
		if valid1 != valid2 {
			return fmt.Errorf("RPKI validation result mismatch for subnet ID %d: %t vs %t", subnetID, valid1, valid2)
		}
	}

	return nil
}

// compareOriginCheckResults checks if two origin check result maps match
func compareOriginCheckResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("Origin check result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, valid1 := range r1 {
		valid2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second result map", subnetID)
		}
		if valid1 != valid2 {
			return fmt.Errorf("Origin check result mismatch for subnet ID %d: %t vs %t", subnetID, valid1, valid2)
		}
	}

	return nil
}

// compareSubnetSyncResults checks if two subnet sync result maps match
func compareSubnetSyncResults(r1, r2 map[uint64]subnettypes.SubnetInfo) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("subnet sync result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, info1 := range r1 {
		info2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second result map", subnetID)
		}

		// Convert both SubnetInfo structs to JSON for comparison
		info1JSON, err := json.Marshal(info1)
		if err != nil {
			return fmt.Errorf("failed to marshal subnet info for comparison: %w", err)
		}

		info2JSON, err := json.Marshal(info2)
		if err != nil {
			return fmt.Errorf("failed to marshal subnet info for comparison: %w", err)
		}

		// Compare JSON representations
		if string(info1JSON) != string(info2JSON) {
			return fmt.Errorf("subnet sync result mismatch for subnet ID %d", subnetID)
		}
	}

	return nil
}

// compareGeoValidationResults checks if two GEO validation result maps match
func compareGeoValidationResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("geo validation result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, valid1 := range r1 {
		valid2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second result map", subnetID)
		}
		if valid1 != valid2 {
			return fmt.Errorf("geo validation result mismatch for subnet ID %d: %t vs %t", subnetID, valid1, valid2)
		}
	}

	return nil
}

// compareGeoValidationResults checks if two subnet ownership verification result maps match
func compareSubnetOwnershipResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("subnet ownership verification result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, valid1 := range r1 {
		valid2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second result map", subnetID)
		}
		if valid1 != valid2 {
			return fmt.Errorf("subnet ownership verification result mismatch for subnet ID %d: %t vs %t", subnetID, valid1, valid2)
		}
	}

	return nil
}

func compareBlacklistSyncResults(r1, r2 map[uint64]trusttypes.SubnetBlacklistCount) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("blacklist sync result maps have different sizes: %d vs %d", len(r1), len(r2))
	}

	// Check if all entries match
	for subnetID, count1 := range r1 {
		count2, exists := r2[subnetID]
		if !exists {
			return fmt.Errorf("subnet ID %d missing from second blacklist result map", subnetID)
		}

		// Convert both SubnetBlacklistCount structs to JSON for comparison
		count1JSON, err := json.Marshal(count1)
		if err != nil {
			return fmt.Errorf("failed to marshal blacklist count for comparison: %w", err)
		}

		count2JSON, err := json.Marshal(count2)
		if err != nil {
			return fmt.Errorf("failed to marshal blacklist count for comparison: %w", err)
		}

		// Compare JSON representations
		if string(count1JSON) != string(count2JSON) {
			return fmt.Errorf("blacklist sync result mismatch for subnet ID %d", subnetID)
		}
	}

	return nil
}

func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	res := &sdk.ResponsePreBlock{}

	// Skip if vote extensions aren't enabled yet
	if len(req.Txs) == 0 || ctx.BlockHeight() <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
		return res, nil
	}

	// Unmarshal the injected vote extension transaction
	var injectedVoteExtTx NetworkValidationResults
	if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
		h.logger.Error("failed to decode injected vote extension tx", "err", err)
		return nil, err
	}

	// Update the validation status for each organization
	for orgID, isValid := range injectedVoteExtTx.OrgValidationResults {
		if err := h.orgKeeper.UpdateOrgDomainValidationStatus(ctx, orgID, isValid); err != nil {
			// Log error but continue with other organizations
			h.logger.Error("failed to update org domain validation status",
				"org_id", orgID,
				"is_valid", isValid,
				"error", err)
		} else {
			h.logger.Info("updated org domain validation status",
				"org_id", orgID,
				"is_valid", isValid)
		}
	}

	// Update the subnet information for each subnet
	for subnetID, subnetInfo := range injectedVoteExtTx.SubnetSyncResults {
		if err := h.subnetKeeper.UpdateSubnetRirDetails(ctx, subnetID, subnetInfo); err != nil {
			// Log error but continue with other subnets
			h.logger.Error("failed to update subnet sync data",
				"subnet_id", subnetID,
				"valid", subnetInfo.Valid,
				"error", err)
		} else {
			h.logger.Info("updated subnet sync data",
				"subnet_id", subnetID,
				"valid", subnetInfo.Valid,
				"org_id", subnetInfo.OrgID,
				"org_name", subnetInfo.OrgName)
		}
	}

	// Update the RPKI validation status for each subnet
	for subnetID, isValid := range injectedVoteExtTx.RpkiValidationResults {
		if err := h.subnetKeeper.UpdateSubnetRpkiValidationStatus(ctx, subnetID, isValid); err != nil {
			// Log error but continue with other subnets
			h.logger.Error("failed to update subnet RPKI validation status",
				"subnet_id", subnetID,
				"is_valid", isValid,
				"error", err)
		} else {
			h.logger.Info("updated subnet RPKI validation status",
				"subnet_id", subnetID,
				"is_valid", isValid)
		}
	}

	// Process origin check results for each subnet
	for subnetId, isInASNDROP := range injectedVoteExtTx.SubnetOriginCheckResults {
		if isInASNDROP {
			if err := h.subnetKeeper.FlagUserRequestedASNDROPOrigin(ctx, subnetId); err != nil {
				h.logger.Error("failed to set ASNDROP origin transfer flag",
				"subnet_id", subnetId,
				"is_in_asndrop", isInASNDROP,
				"error", err)
			} else {
				h.logger.Info("ASNDROP origin transfer flag set",
				"subnet_id", subnetId,
				"is_in_asndrop", isInASNDROP)
			}
		} else {
			if err := h.subnetKeeper.ClearSubnetOriginCheckPending(ctx, subnetId); err != nil {
				h.logger.Error("failed to clear origin check flag",
				"subnet_id", subnetId,
				"is_in_asndrop", isInASNDROP,
				"error", err)
			} else {
				h.logger.Info("origin check cleared",
				"subnet_id", subnetId,
				"is_in_asndrop", isInASNDROP)
			}
		}
	}

	// Update the geo validation status for each subnet
	for subnetID, isValid := range injectedVoteExtTx.GeoValidationResults {
		if err := h.subnetKeeper.UpdateSubnetGeoValidationStatus(ctx, subnetID, isValid); err != nil {
			// Log error but continue with other subnets
			h.logger.Error("failed to update subnet geo validation status",
				"subnet_id", subnetID,
				"is_valid", isValid,
				"error", err)
		} else {
			h.logger.Info("updated subnet geo validation status",
				"subnet_id", subnetID,
				"is_valid", isValid)
		}
	}

	// Update the subnet ownership verification status for each subnet
	for subnetID, isValid := range injectedVoteExtTx.SubnetOwnerValidation {
		if err := h.subnetKeeper.UpdateSubnetOwnershipVerification(ctx, subnetID, isValid); err != nil {
			// Log error but continue with other subnets
			h.logger.Error("failed to update subnet ownership verification status",
				"subnet_id", subnetID,
				"is_valid", isValid,
				"error", err)
		} else {
			h.logger.Info("updated subnet ownership verification status",
				"subnet_id", subnetID,
				"is_valid", isValid)
		}
	}

	// Update the blacklist sync results for each subnet
	for subnetId, blacklistCount := range injectedVoteExtTx.SubnetBlacklistSyncResults {
		subnet, found := h.subnetKeeper.GetSubnet(ctx, subnetId)
		if !found {
			h.logger.Error("failed to update subnet blacklist count. Can't fetch subnet.",
				"subnet_id", subnetId,
				"l1_count", blacklistCount.Level1Count,
				"l2_count", blacklistCount.Level2Count,
				"l3_count", blacklistCount.Level3Count,
				"l4_count", blacklistCount.Level4Count,
				"error", "subnet not found")
			continue
		}

		// Get the contract details to extract the correct lessee org ID
		var contractId uint64 = 0
		var lesseeOrgId uint64 = 0

		if foundContractId, found := h.leaseKeeper.GetFirstLeasedContractBySubnet(ctx, subnetId); found {
			contractId = foundContractId
			if contract, contractFound := h.leaseKeeper.GetContract(ctx, foundContractId); contractFound {
				lesseeOrgId = contract.LesseeOrg
			}
		}

		if err := h.trustKeeper.UpdateSubnetReputation(ctx, subnetId, contractId, subnet.OwnerOrgId, lesseeOrgId, blacklistCount); err != nil {
			// Log error but continue with other subnets
			h.logger.Error("failed to update subnet blacklist count",
				"subnet_id", subnetId,
				"l1_count", blacklistCount.Level1Count,
				"l2_count", blacklistCount.Level2Count,
				"l3_count", blacklistCount.Level3Count,
				"l4_count", blacklistCount.Level4Count,
				"error", err.Error())
		} else {
			h.logger.Info("updated subnet blacklist count",
				"subnet_id", subnetId,
				"l1_count", blacklistCount.Level1Count,
				"l2_count", blacklistCount.Level2Count,
				"l3_count", blacklistCount.Level3Count,
				"l4_count", blacklistCount.Level4Count)
		}
	}

	return res, nil
}
