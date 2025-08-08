package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/lease/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListContract(ctx context.Context, req *types.QueryListContractRequest) (*types.QueryListContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var contracts []types.Contract
	var totalCount uint64
	var lastKey uint64

	// Set up pagination parameters
	limit := uint64(100) // Default limit
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

	// Use collections to iterate through contracts with pagination
	err := k.Contracts.Walk(ctx, nil, func(key uint64, value types.Contract) (bool, error) {
		totalCount++

		// If using key-based pagination, skip until we reach the start key
		if startKey != nil && key <= *startKey {
			return false, nil
		}

		// Skip entries until we reach the offset (for offset-based pagination)
		if startKey == nil && skipped < offset {
			skipped++
			return false, nil
		}

		if collectingItems {
			// Stop if we've collected enough items
			if collectedCount >= limit {
				collectingItems = false
				return false, nil
			}

			contracts = append(contracts, value)
			lastKey = key
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

	return &types.QueryListContractResponse{
		Contract: contracts,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
