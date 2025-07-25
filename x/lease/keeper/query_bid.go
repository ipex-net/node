package keeper

import (
	"context"

	"ipex/x/lease/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Bid(ctx context.Context, req *types.QueryGetBidRequest) (*types.QueryGetBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Validate input parameters
	if req.ContractId == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract ID cannot be zero")
	}
	if req.BidderOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "bidder org ID cannot be zero")
	}

	// Get bid using contract ID and bidder org ID
	bid, found := k.GetBid(ctx, req.ContractId, req.BidderOrgId)
	if !found {
		return nil, status.Error(codes.NotFound, "bid not found")
	}

	return &types.QueryGetBidResponse{Bid: bid}, nil
}
