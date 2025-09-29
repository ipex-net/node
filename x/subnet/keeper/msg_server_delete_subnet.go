package keeper

import (
	"context"

	"ipex/x/subnet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DeleteSubnet(goCtx context.Context, msg *types.MsgDeleteSubnet) (*types.MsgDeleteSubnetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgDeleteSubnetResponse{}, nil
}
