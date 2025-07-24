package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RequestLesseeRefund(goCtx context.Context, msg *types.MsgRequestLesseeRefund) (*types.MsgRequestLesseeRefundResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRequestLesseeRefundResponse{}, nil
}
