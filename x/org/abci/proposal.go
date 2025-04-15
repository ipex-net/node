package abci

import (
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/org/keeper"
)

// StakeDomainValidation defines the structure a proposer should use to calculate
// and submit the stake-weighted domain validation results, in addition to the vote
// extensions used to calculate them.
type StakeDomainValidation struct {
	ValidationResults   map[uint64]bool
	ExtendedCommitInfo  abci.ExtendedCommitInfo
}

type ProposalHandler struct {
	logger   log.Logger
	keeper   keeper.Keeper
	valStore baseapp.ValidatorStore
}

func NewProposalHandler(logger log.Logger, keeper keeper.Keeper, valStore baseapp.ValidatorStore) *ProposalHandler {
	return &ProposalHandler{
		logger:   logger,
		keeper:   keeper,
		valStore: valStore,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		proposalTxs := req.Txs

		if req.Height > ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			// Validate vote extensions
			err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			// Compute consensus on domain validation results
			validationResults, err := h.computeValidationResults(ctx, req.LocalLastCommit)
			if err != nil {
				return nil, errors.New("failed to compute domain validation results")
			}

			// Create the injected transaction with validation results
			injectedVoteExtTx := StakeDomainValidation{
				ValidationResults: validationResults,
				ExtendedCommitInfo: req.LocalLastCommit,
			}

			// Marshal to JSON
			bz, err := json.Marshal(injectedVoteExtTx)
			if err != nil {
				h.logger.Error("failed to encode injected vote extension tx", "err", err)
				return nil, errors.New("failed to encode injected vote extension tx")
			}

			// Inject as first transaction in the block
			proposalTxs = append([][]byte{bz}, proposalTxs...)
		}

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

// computeValidationResults aggregates validation results from vote extensions
func (h *ProposalHandler) computeValidationResults(_ sdk.Context, ci abci.ExtendedCommitInfo) (map[uint64]bool, error) {
	// Map to store aggregated validation results
	// org ID -> validation result
	validationResults := make(map[uint64]bool)
	
	// Maps to track voting power for each validation
	validPower := make(map[uint64]int64)
	invalidPower := make(map[uint64]int64)
	
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
		
		// Aggregate validation results based on voting power
		for orgID, isValid := range voteExt.OrgDomainValidation {
			if isValid {
				validPower[orgID] += v.Validator.Power
			} else {
				invalidPower[orgID] += v.Validator.Power
			}
		}
	}
	
	// Determine final validation result based on majority voting power
	for orgID := range validPower {
		total := validPower[orgID] + invalidPower[orgID]
		if total > 0 && validPower[orgID] > total/2 {
			// More than 50% of voting power says valid
			validationResults[orgID] = true
		} else {
			// Less than 50% says valid
			validationResults[orgID] = false
		}
	}
	
	return validationResults, nil
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
		var injectedVoteExtTx StakeDomainValidation
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Validate the vote extensions
		err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			h.logger.Error("failed to validate vote extensions", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Verify the proposer's validation results by computing the same calculation
		validationResults, err := h.computeValidationResults(ctx, injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			h.logger.Error("failed to compute validation results", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		// Compare the proposer's validation results with our own calculation
		if err := compareValidationResults(injectedVoteExtTx.ValidationResults, validationResults); err != nil {
			h.logger.Error("validation results don't match", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// compareValidationResults checks if two validation result maps match
func compareValidationResults(r1, r2 map[uint64]bool) error {
	// Check if both maps have the same number of entries
	if len(r1) != len(r2) {
		return fmt.Errorf("validation result maps have different sizes: %d vs %d", len(r1), len(r2))
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

func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	res := &sdk.ResponsePreBlock{}
	
	// Skip if no transactions or if vote extensions aren't enabled yet
	if len(req.Txs) == 0 || ctx.BlockHeight() <= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
		return res, nil
	}
	
	// Unmarshal the injected vote extension transaction
	var injectedVoteExtTx StakeDomainValidation
	if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
		h.logger.Error("failed to decode injected vote extension tx", "err", err)
		return nil, err
	}
	
	// Update the validation status for each organization
	for orgID, isValid := range injectedVoteExtTx.ValidationResults {
		if err := h.keeper.UpdateOrgDomainValidationStatus(ctx, orgID, isValid); err != nil {
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
	
	return res, nil
}