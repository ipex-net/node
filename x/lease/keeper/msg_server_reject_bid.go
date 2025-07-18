package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RejectBid(goCtx context.Context, msg *types.MsgRejectBid) (*types.MsgRejectBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRejectBidResponse{}, nil
}
