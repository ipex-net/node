package keeper

import (
	"context"

	"ipex/x/abuse/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ComplaintByLessor(goCtx context.Context, req *types.QueryComplaintByLessorRequest) (*types.QueryComplaintByLessorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Process the query
	_ = ctx

	return &types.QueryComplaintByLessorResponse{}, nil
}
