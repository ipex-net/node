package keeper

import (
	"context"
	"strconv"

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

	// Convert string index to uint64
	indexUint64, err := strconv.ParseUint(req.Index, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid index format")
	}

	// Find bid by index field
	// Walk through all bids to find the one with matching index
	allBids := k.GetAllBids(ctx)
	for _, bid := range allBids {
		if bid.Index == indexUint64 {
			return &types.QueryGetBidResponse{Bid: bid}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "bid not found")
}
