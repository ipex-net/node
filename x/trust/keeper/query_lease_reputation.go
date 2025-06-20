package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LeaseReputationAll(ctx context.Context, req *types.QueryAllLeaseReputationRequest) (*types.QueryAllLeaseReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var leaseReputations []types.LeaseReputation

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	leaseReputationStore := prefix.NewStore(store, types.KeyPrefix(types.LeaseReputationKeyPrefix))

	pageRes, err := query.Paginate(leaseReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var leaseReputation types.LeaseReputation
		if err := k.cdc.Unmarshal(value, &leaseReputation); err != nil {
			return err
		}

		leaseReputations = append(leaseReputations, leaseReputation)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLeaseReputationResponse{LeaseReputation: leaseReputations, Pagination: pageRes}, nil
}

func (k Keeper) LeaseReputation(ctx context.Context, req *types.QueryGetLeaseReputationRequest) (*types.QueryGetLeaseReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLeaseReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLeaseReputationResponse{LeaseReputation: val}, nil
}
