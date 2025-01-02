package keeper

import (
	"context"

	"ipex/x/subnet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateSubnet(goCtx context.Context, msg *types.MsgCreateSubnet) (*types.MsgCreateSubnetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCreateSubnetResponse{}, nil
}
