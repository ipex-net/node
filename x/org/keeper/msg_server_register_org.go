package keeper

import (
	"context"

	"ipex/x/org/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterOrg(goCtx context.Context, msg *types.MsgRegisterOrg) (*types.MsgRegisterOrgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRegisterOrgResponse{}, nil
}
