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

func (k Keeper) LessorReputationAll(ctx context.Context, req *types.QueryAllLessorReputationRequest) (*types.QueryAllLessorReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var lessorReputations []types.LessorReputation

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	lessorReputationStore := prefix.NewStore(store, types.KeyPrefix(types.LessorReputationKeyPrefix))

	pageRes, err := query.Paginate(lessorReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var lessorReputation types.LessorReputation
		if err := k.cdc.Unmarshal(value, &lessorReputation); err != nil {
			return err
		}

		lessorReputations = append(lessorReputations, lessorReputation)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLessorReputationResponse{LessorReputation: lessorReputations, Pagination: pageRes}, nil
}

func (k Keeper) LessorReputation(ctx context.Context, req *types.QueryGetLessorReputationRequest) (*types.QueryGetLessorReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLessorReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetLessorReputationResponse{LessorReputation: val}, nil
}
