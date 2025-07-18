package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AcceptBid(goCtx context.Context, msg *types.MsgAcceptBid) (*types.MsgAcceptBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgAcceptBidResponse{}, nil
}
