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

func (k Keeper) LesseeReputationAll(ctx context.Context, req *types.QueryAllLesseeReputationRequest) (*types.QueryAllLesseeReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var lesseeReputations []types.LesseeReputation

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	lesseeReputationStore := prefix.NewStore(store, types.KeyPrefix(types.LesseeReputationKeyPrefix))

	pageRes, err := query.Paginate(lesseeReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var lesseeReputation types.LesseeReputation
		if err := k.cdc.Unmarshal(value, &lesseeReputation); err != nil {
			return err
		}

		lesseeReputations = append(lesseeReputations, lesseeReputation)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLesseeReputationResponse{LesseeReputation: lesseeReputations, Pagination: pageRes}, nil
}

func (k Keeper) LesseeReputation(ctx context.Context, req *types.QueryGetLesseeReputationRequest) (*types.QueryGetLesseeReputationResponse, error) {
	if req == nil || req.Index == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLesseeReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLesseeReputationResponse{LesseeReputation: val}, nil
}
