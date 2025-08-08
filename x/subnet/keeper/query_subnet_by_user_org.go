package keeper

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ipex/x/subnet/types"
)

func (k Keeper) SubnetByUserOrg(ctx context.Context, req *types.QuerySubnetByUserOrgRequest) (*types.QuerySubnetByUserOrgResponse, error) {
	if req == nil || req.UserOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnets []types.Subnet
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
			// Parse the key as uint64 (subnet ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	var skipped uint64
	var collectedCount uint64

	// Iterate over all subnets with this user org ID using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.UserOrgId)
	err := k.SubnetsByUserOrg.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// key.K2() is the subnet ID
		subnetId := key.K2()
		subnet, found := k.GetSubnet(ctx, subnetId)
		if !found {
			return false, nil
		}

		// Count total matching subnets
		totalCount++

		// If using key-based pagination, skip until we reach the start key
		if startKey != nil && subnetId <= *startKey {
			return false, nil
		}

		// Skip entries until we reach the offset (for offset-based pagination)
		if startKey == nil && skipped < offset {
			skipped++
			return false, nil
		}

		// Stop if we've collected enough items
		if collectedCount >= limit {
			return true, nil
		}

		subnets = append(subnets, subnet)
		lastKey = subnetId
		collectedCount++

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if collectedCount == limit {
		// Convert the last subnet ID to bytes for next key
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, lastKey)
	}

	return &types.QuerySubnetByUserOrgResponse{
		Subnet:     subnets,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
