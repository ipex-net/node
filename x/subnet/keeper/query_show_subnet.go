package keeper

import (
	"context"

	"ipex/x/subnet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ShowSubnet(goCtx context.Context, req *types.QueryShowSubnetRequest) (*types.QueryShowSubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	subnet, found := k.GetSubnet(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	//Access control. We can implemented permissioned output here. Example: only show certain fields to owner or user

	return &types.QueryShowSubnetResponse{Subnet: subnet}, nil
}
