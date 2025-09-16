package keeper

import (
	"context"

	"ipex/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListPaymentFiltered(goCtx context.Context, req *types.QueryListPaymentFilteredRequest) (*types.QueryListPaymentFilteredResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Process the query
	_ = ctx

	return &types.QueryListPaymentFilteredResponse{}, nil
}
