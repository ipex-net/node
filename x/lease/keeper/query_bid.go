package keeper

import (
	"context"

	"ipex/x/lease/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BidAll(ctx context.Context, req *types.QueryAllBidRequest) (*types.QueryAllBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	bids := k.GetAllBids(ctx)

	// Create a mock pagination response for compatibility
	pageRes := &query.PageResponse{
		Total: uint64(len(bids)),
	}

	return &types.QueryAllBidResponse{Bid: bids, Pagination: pageRes}, nil
}

func (k Keeper) Bid(ctx context.Context, req *types.QueryGetBidRequest) (*types.QueryGetBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Get bid using contract ID and bidder org ID
	bid, found := k.GetBid(ctx, req.ContractId, req.BidderOrgId)
	if !found {
		return nil, status.Error(codes.NotFound, "bid not found")
	}

	return &types.QueryGetBidResponse{Bid: bid}, nil
}
