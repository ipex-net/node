package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RevokeBid(goCtx context.Context, msg *types.MsgRevokeBid) (*types.MsgRevokeBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRevokeBidResponse{}, nil
}
