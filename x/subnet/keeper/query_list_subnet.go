package keeper

import (
	"context"

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
	limit := uint64(100) // Default limit
	offset := uint64(0)
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
	}

	// Use collections to iterate through subnets with pagination
	err := k.Subnets.Walk(ctx, nil, func(key uint64, value types.Subnet) (bool, error) {
		count++

		// Skip entries until we reach the offset
		if skipped < offset {
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
	if uint64(len(subnets)) == limit && count > offset+limit {
		// There are more items available
		nextKey = []byte("has_more")
	}

	return &types.QueryListSubnetResponse{
		Subnet: subnets,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}
