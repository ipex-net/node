package keeper

import (
	"context"

	"ipex/x/abuse/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CloseComplaint(goCtx context.Context, msg *types.MsgCloseComplaint) (*types.MsgCloseComplaintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgCloseComplaintResponse{}, nil
}
