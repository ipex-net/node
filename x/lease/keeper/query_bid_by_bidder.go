package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BidByBidder(goCtx context.Context, req *types.QueryBidByBidderRequest) (*types.QueryBidByBidderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Validate input parameters
	if req.BidderOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "bidder org ID cannot be zero")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use the existing helper function to get all bids for this bidder
	bids := k.GetBidsByBidder(ctx, req.BidderOrgId)

	// Set up pagination defaults
	limit := uint64(100)
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = req.Pagination.Limit
	}

	var startIdx uint64 = 0
	if req.Pagination != nil && len(req.Pagination.Key) > 0 {
		// Decode startIdx from pagination key
		if len(req.Pagination.Key) >= 8 {
			startIdx = binary.BigEndian.Uint64(req.Pagination.Key[:8])
		}
	}

	// Apply pagination to the results
	var paginatedBids []types.Bid
	var nextKey []byte

	endIdx := startIdx + limit
	if endIdx > uint64(len(bids)) {
		endIdx = uint64(len(bids))
	}

	if startIdx < uint64(len(bids)) {
		paginatedBids = bids[startIdx:endIdx]
		
		// Set nextKey if there are more results
		if endIdx < uint64(len(bids)) {
			nextKey = make([]byte, 8)
			binary.BigEndian.PutUint64(nextKey, endIdx)
		}
	}

	return &types.QueryBidByBidderResponse{
		Bid: paginatedBids,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   uint64(len(paginatedBids)),
		},
	}, nil
}
