package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/subnet/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListSubnet(ctx context.Context, req *types.QueryListSubnetRequest) (*types.QueryListSubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnets []types.Subnet
	var count uint64
	var skipped uint64

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
			// Parse the key as uint64 (subnet ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	// Use collections to iterate through subnets with pagination
	err := k.Subnets.Walk(ctx, nil, func(key uint64, value types.Subnet) (bool, error) {
		count++

		// If using key-based pagination, skip until we reach the start key
		if startKey != nil && key <= *startKey {
			return false, nil
		}

		// Skip entries until we reach the offset (for offset-based pagination)
		if startKey == nil && skipped < offset {
			skipped++
			return false, nil
		}

		// Stop if we've collected enough items
		if uint64(len(subnets)) >= limit {
			return true, nil
		}

		subnets = append(subnets, value)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	var lastKey uint64
	if uint64(len(subnets)) == limit {
		// Convert the last subnet ID to bytes for next key
		if len(subnets) > 0 {
			lastKey = subnets[len(subnets)-1].Id
			nextKey = make([]byte, 8)
			binary.BigEndian.PutUint64(nextKey, lastKey)
		}
	}

	return &types.QueryListSubnetResponse{
		Subnet: subnets,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}
