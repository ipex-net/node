package keeper

import (
	"context"

	"ipex/x/trust/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogAll(ctx context.Context, req *types.QueryAllBlacklistLogRequest) (*types.QueryAllBlacklistLogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklistLogs []types.BlacklistLog

	err := k.BlacklistLogs.Walk(ctx, nil, func(key uint64, value types.BlacklistLog) (bool, error) {
		blacklistLogs = append(blacklistLogs, value)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// For now, return without pagination support (can be improved later)
	return &types.QueryAllBlacklistLogResponse{
		BlacklistLog: blacklistLogs, 
		Pagination: &query.PageResponse{Total: uint64(len(blacklistLogs))},
	}, nil
}

func (k Keeper) BlacklistLog(ctx context.Context, req *types.QueryGetBlacklistLogRequest) (*types.QueryGetBlacklistLogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	blacklistLog, found := k.GetBlacklistLog(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetBlacklistLogResponse{BlacklistLog: blacklistLog}, nil
}
