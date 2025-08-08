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

func (k Keeper) BidByContract(goCtx context.Context, req *types.QueryBidByContractRequest) (*types.QueryBidByContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Validate input parameters
	if req.ContractId == 0 {
		return nil, status.Error(codes.InvalidArgument, "contract ID cannot be zero")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use the existing helper function to get all bids for this contract
	bids := k.GetBidsByContract(ctx, req.ContractId)

	// Set up pagination defaults
	limit := uint64(20)
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

	return &types.QueryBidByContractResponse{
		Bid: paginatedBids,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   uint64(len(bids)),
		},
	}, nil
}
