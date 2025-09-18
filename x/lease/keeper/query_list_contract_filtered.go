package keeper

import (
	"context"
	"encoding/binary"
	"strings"

	"ipex/x/lease/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListContractFiltered(goCtx context.Context, req *types.QueryListContractFilteredRequest) (*types.QueryListContractFilteredResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Owner organization ID is required
	if req.OwnerOrg == 0 {
		return nil, status.Error(codes.InvalidArgument, "owner organization ID is required")
	}

	// Validate required mask field
	if req.Mask < 8 || req.Mask > 24 {
		return nil, errorsmod.Wrap(types.ErrInvalidPrefixLength, "mask must be between 8 and 24")
	}

	// Validate optional fields if provided (strict validation - no trimming)
	if req.Rir != "" && req.Rir != "none" && req.Rir != "any" {
		if err := validateRir(req.Rir); err != nil {
			return nil, errorsmod.Wrap(types.ErrUnknownRIR, err.Error())
		}
	} else {
		req.Rir = "" // Normalize placeholder values to empty string
	}

	if req.Country != "" && req.Country != "none" && req.Country != "any" {
		if err := validateCountryCode(req.Country); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidCountryCode, err.Error())
		}
	} else {
		req.Country = "" // Normalize placeholder values to empty string
	}

	if req.NetworkAddress != 0 {
		if err := validateNetworkAddress(req.NetworkAddress); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidIPFormat, err.Error())
		}
	}

	// Validate first octet field if provided
	if req.FirstOctet != 0 {
		if err := validateOctet(req.FirstOctet, "first octet"); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidIPFormat, err.Error())
		}
	}

	// Validate price range
	if req.MinPrice > 0 && req.MaxPrice > 0 && req.MinPrice > req.MaxPrice {
		return nil, errorsmod.Wrap(types.ErrInvalidPrice, "minimum price cannot be greater than maximum price")
	}

	// Validate date ranges
	if req.CreatedAtFrom != 0 && req.CreatedAtTo != 0 && req.CreatedAtFrom > req.CreatedAtTo {
		return nil, status.Error(codes.InvalidArgument, "createdAtFrom cannot be greater than createdAtTo")
	}
	if req.ActivatedAtFrom != 0 && req.ActivatedAtTo != 0 && req.ActivatedAtFrom > req.ActivatedAtTo {
		return nil, status.Error(codes.InvalidArgument, "activatedAtFrom cannot be greater than activatedAtTo")
	}
	if req.PrepaidUntilFrom != 0 && req.PrepaidUntilTo != 0 && req.PrepaidUntilFrom > req.PrepaidUntilTo {
		return nil, status.Error(codes.InvalidArgument, "prepaidUntilFrom cannot be greater than prepaidUntilTo")
	}
	if req.UpdatedAtFrom != 0 && req.UpdatedAtTo != 0 && req.UpdatedAtFrom > req.UpdatedAtTo {
		return nil, status.Error(codes.InvalidArgument, "updatedAtFrom cannot be greater than updatedAtTo")
	}

	var contracts []types.Contract
	var totalCount uint64
	var lastKey uint64

	// Set up pagination parameters
	limit := uint64(20) // Default limit
	offset := uint64(0)
	var startKey *uint64

	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		// Support key-based pagination
		if len(req.Pagination.Key) > 0 {
			// Parse the key as uint64 (contract ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	var skipped uint64
	var collectedCount uint64
	var collectingItems = true

	// Use owner index to efficiently get contracts for this organization (reverse for latest first)
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.OwnerOrg).Descending()
	err := k.ContractsByOwner.Walk(goCtx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// Get the contract ID from the key
		contractId := key.K2()

		// Fetch the full contract
		contract, found := k.GetContract(goCtx, contractId)
		if !found {
			// Contract might have been deleted, skip
			return false, nil
		}

		// Apply filters first - skip contract if it doesn't match criteria

		// Filter by lessee organization (0 means no filter)
		if req.LesseeOrg != 0 && contract.LesseeOrg != req.LesseeOrg {
			return false, nil
		}

		// Filter by mask (0 means no filter)
		if req.Mask != 0 && contract.Mask != req.Mask {
			return false, nil
		}

		// Filter by RIR (empty means no filter)
		if req.Rir != "" && !strings.EqualFold(contract.Rir, req.Rir) {
			return false, nil
		}

		// Filter by country (empty means no filter)
		if req.Country != "" && !strings.EqualFold(contract.Country, req.Country) {
			return false, nil
		}

		// Filter by price range (0 means no filter)
		contractPrice := contract.Price.Amount.Uint64()
		if req.MinPrice > 0 && contractPrice < req.MinPrice {
			return false, nil
		}
		if req.MaxPrice > 0 && contractPrice > req.MaxPrice {
			return false, nil
		}

		// Filter by network address (0 means no filter)
		if req.NetworkAddress != 0 && contract.NetworkAddress != req.NetworkAddress {
			return false, nil
		}

		// Filter by first octet (0 means no filter)
		if req.FirstOctet != 0 {
			contractFirstOctet := (contract.NetworkAddress >> 24) & 0xFF
			if contractFirstOctet != req.FirstOctet {
				return false, nil
			}
		}

		// Filter by status (0 means no filter)
		if req.Status != 0 && uint64(contract.Status) != req.Status {
			return false, nil
		}

		// Filter by revocation period (0 means no filter)
		if req.RevocationPeriodMonth != 0 && contract.RevocationPeriodMonth != req.RevocationPeriodMonth {
			return false, nil
		}

		// Filter by created timestamp range (0 means no filter)
		// CreatedAt is time.Time (non-nullable), check for zero value
		if !contract.CreatedAt.IsZero() {
			createdTimestamp := uint64(contract.CreatedAt.Unix())
			if req.CreatedAtFrom != 0 && createdTimestamp < req.CreatedAtFrom {
				return false, nil
			}
			if req.CreatedAtTo != 0 && createdTimestamp > req.CreatedAtTo {
				return false, nil
			}
		}

		// Filter by activated timestamp range (0 means no filter)
		if contract.ActivatedAt != nil {
			activatedTimestamp := uint64(contract.ActivatedAt.Unix())
			if req.ActivatedAtFrom != 0 && activatedTimestamp < req.ActivatedAtFrom {
				return false, nil
			}
			if req.ActivatedAtTo != 0 && activatedTimestamp > req.ActivatedAtTo {
				return false, nil
			}
		}

		// Filter by prepaid until timestamp range (0 means no filter)
		if contract.PrepaidUntil != nil {
			prepaidTimestamp := uint64(contract.PrepaidUntil.Unix())
			if req.PrepaidUntilFrom != 0 && prepaidTimestamp < req.PrepaidUntilFrom {
				return false, nil
			}
			if req.PrepaidUntilTo != 0 && prepaidTimestamp > req.PrepaidUntilTo {
				return false, nil
			}
		}

		// Filter by updated timestamp range (0 means no filter)
		// UpdatedAt is time.Time (non-nullable), check for zero value
		if !contract.UpdatedAt.IsZero() {
			updatedTimestamp := uint64(contract.UpdatedAt.Unix())
			if req.UpdatedAtFrom != 0 && updatedTimestamp < req.UpdatedAtFrom {
				return false, nil
			}
			if req.UpdatedAtTo != 0 && updatedTimestamp > req.UpdatedAtTo {
				return false, nil
			}
		}

		// Contract passed all filters - count it
		totalCount++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key (reverse order)
			if startKey != nil && contractId >= *startKey {
				return false, nil
			}

			// Skip entries until we reach the offset (for offset-based pagination)
			if startKey == nil && skipped < offset {
				skipped++
				return false, nil
			}

			// Stop collecting items if we've collected enough
			if collectedCount >= limit {
				collectingItems = false
				return false, nil // Continue counting but stop collecting
			}

			contracts = append(contracts, contract)
			lastKey = contractId
			collectedCount++
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if collectedCount == limit {
		// Convert the last contract ID to bytes for next key
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, lastKey)
	}

	return &types.QueryListContractFilteredResponse{
		Contract: contracts,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
