package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"ipex/x/trust/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogBySubnet(ctx context.Context, req *types.QueryBlacklistLogBySubnetRequest) (*types.QueryBlacklistLogBySubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklistLogs []types.BlacklistLog

	// Iterate over all logs with this subnet ID using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.SubnetId)
	err := k.BlacklistLogsBySubnet.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// key.K2() is the log ID
		logId := key.K2()
		log, found := k.GetBlacklistLog(ctx, logId)
		if found {
			blacklistLogs = append(blacklistLogs, log)
		}
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBlacklistLogBySubnetResponse{
		BlacklistLog: blacklistLogs,
		Pagination:   &query.PageResponse{Total: uint64(len(blacklistLogs))},
	}, nil
}
