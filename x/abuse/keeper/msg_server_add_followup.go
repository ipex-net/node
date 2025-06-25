package keeper

import (
	"context"

	"ipex/x/abuse/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AddFollowup(goCtx context.Context, msg *types.MsgAddFollowup) (*types.MsgAddFollowupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgAddFollowupResponse{}, nil
}
