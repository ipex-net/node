package keeper

import (
	"context"

	"ipex/x/lease/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BidAll(ctx context.Context, req *types.QueryAllBidRequest) (*types.QueryAllBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var bids []types.Bid

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bidStore := prefix.NewStore(store, types.KeyPrefix(types.BidKeyPrefix))

	pageRes, err := query.Paginate(bidStore, req.Pagination, func(key []byte, value []byte) error {
		var bid types.Bid
		if err := k.cdc.Unmarshal(value, &bid); err != nil {
			return err
		}

		bids = append(bids, bid)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBidResponse{Bid: bids, Pagination: pageRes}, nil
}

func (k Keeper) Bid(ctx context.Context, req *types.QueryGetBidRequest) (*types.QueryGetBidResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetBid(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBidResponse{Bid: val}, nil
}
