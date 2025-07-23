package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) CanTerminateLeaseWithoutPenalty(goCtx context.Context, req *types.QueryCanTerminateLeaseWithoutPenaltyRequest) (*types.QueryCanTerminateLeaseWithoutPenaltyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	eligible := k.CanTerminateWithoutPenalty(ctx, req.ContractId)
	return &types.QueryCanTerminateLeaseWithoutPenaltyResponse{Eligible: eligible}, nil
}
