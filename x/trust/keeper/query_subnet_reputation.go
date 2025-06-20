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

func (k Keeper) SubnetReputationAll(ctx context.Context, req *types.QueryAllSubnetReputationRequest) (*types.QueryAllSubnetReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnetReputations []types.SubnetReputation

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	subnetReputationStore := prefix.NewStore(store, types.KeyPrefix(types.SubnetReputationKeyPrefix))

	pageRes, err := query.Paginate(subnetReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var subnetReputation types.SubnetReputation
		if err := k.cdc.Unmarshal(value, &subnetReputation); err != nil {
			return err
		}

		subnetReputations = append(subnetReputations, subnetReputation)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSubnetReputationResponse{SubnetReputation: subnetReputations, Pagination: pageRes}, nil
}

func (k Keeper) SubnetReputation(ctx context.Context, req *types.QueryGetSubnetReputationRequest) (*types.QueryGetSubnetReputationResponse, error) {
	if req == nil || req.Index == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetSubnetReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetSubnetReputationResponse{SubnetReputation: val}, nil
}
