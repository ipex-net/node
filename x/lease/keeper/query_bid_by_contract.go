package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/lease/types"

	"cosmossdk.io/collections"
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

	// Set up pagination defaults
	limit := uint64(100)
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = req.Pagination.Limit
	}

	var startKey []byte
	if req.Pagination != nil && len(req.Pagination.Key) > 0 {
		startKey = req.Pagination.Key
	}

	var bids []types.Bid
	var nextKey []byte
	count := uint64(0)

	// Iterate through the Bids collection with prefix for this contract
	err := k.Bids.Walk(ctx, collections.NewPrefixedPairRange[uint64, uint64](req.ContractId), func(key collections.Pair[uint64, uint64], bid types.Bid) (bool, error) {
		// If we have a startKey, skip until we reach it
		if len(startKey) > 0 {
			keyBytes := make([]byte, 16)
			binary.BigEndian.PutUint64(keyBytes[:8], key.K1())
			binary.BigEndian.PutUint64(keyBytes[8:], key.K2())

			if string(keyBytes) <= string(startKey) {
				return false, nil
			}
		}

		// Add the bid to results
		bids = append(bids, bid)
		count++

		// If we've reached the limit, set nextKey and stop
		if count >= limit {
			keyBytes := make([]byte, 16)
			binary.BigEndian.PutUint64(keyBytes[:8], key.K1())
			binary.BigEndian.PutUint64(keyBytes[8:], key.K2())
			nextKey = keyBytes
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBidByContractResponse{
		Bid: bids,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}
