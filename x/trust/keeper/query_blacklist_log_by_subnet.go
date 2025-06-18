package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogBySubnet(goCtx context.Context, req *types.QueryBlacklistLogBySubnetRequest) (*types.QueryBlacklistLogBySubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var blacklistLogs []types.BlacklistLog
	
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	blacklistLogStore := prefix.NewStore(store, types.KeyPrefix(types.BlacklistLogKey))
	
	pageRes, err := query.Paginate(blacklistLogStore, req.Pagination, func(key []byte, value []byte) error {
		var blacklistLog types.BlacklistLog
		if err := k.cdc.Unmarshal(value, &blacklistLog); err != nil {
			return err
		}
		
		// Filter by subnetId - only add matching logs
		if blacklistLog.SubnetId == req.SubnetId {
			blacklistLogs = append(blacklistLogs, blacklistLog)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	
	return &types.QueryBlacklistLogBySubnetResponse{
		BlacklistLog: blacklistLogs,
		Pagination:   pageRes,
	}, nil
}