package keeper

import (
	"context"

	"ipex/x/abuse/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ComplaintAll(ctx context.Context, req *types.QueryAllComplaintRequest) (*types.QueryAllComplaintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var complaints []types.Complaint

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	complaintStore := prefix.NewStore(store, types.KeyPrefix(types.ComplaintKeyPrefix))

	pageRes, err := query.Paginate(complaintStore, req.Pagination, func(key []byte, value []byte) error {
		var complaint types.Complaint
		if err := k.cdc.Unmarshal(value, &complaint); err != nil {
			return err
		}

		complaints = append(complaints, complaint)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllComplaintResponse{Complaint: complaints, Pagination: pageRes}, nil
}

func (k Keeper) Complaint(ctx context.Context, req *types.QueryGetComplaintRequest) (*types.QueryGetComplaintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetComplaint(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetComplaintResponse{Complaint: val}, nil
}
