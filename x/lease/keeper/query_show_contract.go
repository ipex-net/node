package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ShowContract(goCtx context.Context, req *types.QueryShowContractRequest) (*types.QueryShowContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Check that contractId is not nil
	if req.Id == 0 {
		return nil, types.ErrContractIdCantBeNil
	}
	
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, found := k.GetContract(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryShowContractResponse{Contract: contract}, nil
}
