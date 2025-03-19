package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) TerminateContract(goCtx context.Context, msg *types.MsgTerminateContract) (*types.MsgTerminateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgTerminateContractResponse{}, nil
}
