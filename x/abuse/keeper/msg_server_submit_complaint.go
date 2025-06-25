package keeper

import (
	"context"

	"ipex/x/abuse/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SubmitComplaint(goCtx context.Context, msg *types.MsgSubmitComplaint) (*types.MsgSubmitComplaintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgSubmitComplaintResponse{}, nil
}
