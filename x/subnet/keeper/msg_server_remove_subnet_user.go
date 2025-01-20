package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RemoveSubnetUser(goCtx context.Context, msg *types.MsgRemoveSubnetUser) (*types.MsgRemoveSubnetUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	//Checking if message creator is subnet owner
	// Verify creator is org admin
	if subnet.OwnerOrgId > 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.OwnerOrgId)
	}

	//TODO: Process contract termination logic

	subnet.Origin = 0
	//Updating the subnet
	subnet.UserOrgId = 0
	k.SetSubnet(ctx, subnet)

	return &types.MsgRemoveSubnetUserResponse{}, nil
}
