package keeper

import (
	"context"

	"ipex/x/trust/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogByLessee(goCtx context.Context, req *types.QueryBlacklistLogByLesseeRequest) (*types.QueryBlacklistLogByLesseeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Process the query
	_ = ctx

	return &types.QueryBlacklistLogByLesseeResponse{}, nil
}
