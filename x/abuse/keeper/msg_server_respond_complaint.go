package keeper

import (
	"context"

	"ipex/x/abuse/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RespondComplaint(goCtx context.Context, msg *types.MsgRespondComplaint) (*types.MsgRespondComplaintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRespondComplaintResponse{}, nil
}
