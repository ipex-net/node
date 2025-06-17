package keeper

import (
	"context"

	"ipex/x/trust/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogBySubnet(goCtx context.Context, req *types.QueryBlacklistLogBySubnetRequest) (*types.QueryBlacklistLogBySubnetResponse, error) {
	if req == nil {
			return nil, status.Error(codes.InvalidArgument, "invalid request")
		}

	ctx := sdk.UnwrapSDKContext(goCtx)
	
	// Get all blacklist logs
	allLogs := k.GetAllBlacklistLog(ctx)
	
	// Filter logs by subnetId
	
	var filteredLogs []types.BlacklistLog
	for _, log := range allLogs {
		if log.SubnetId == req.SubnetId {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return &types.QueryBlacklistLogBySubnetResponse{
		BlacklistLog: filteredLogs,
}, nil
}
