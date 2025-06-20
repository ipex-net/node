package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogByLessee(ctx context.Context, req *types.QueryBlacklistLogByLesseeRequest) (*types.QueryBlacklistLogByLesseeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklistLogs []types.BlacklistLog

	// Iterate over all logs with this lessee ID using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.LesseeOrgId)
	err := k.BlacklistLogsByLessee.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
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

	return &types.QueryBlacklistLogByLesseeResponse{
		BlacklistLog: blacklistLogs,
		Pagination:   &query.PageResponse{Total: uint64(len(blacklistLogs))},
	}, nil
}
