package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateSubnetUser(goCtx context.Context, msg *types.MsgUpdateSubnetUser) (*types.MsgUpdateSubnetUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// checking that user org exists
	if msg.UserOrgId == 0 || !k.orgKeeper.HasOrg(ctx, msg.UserOrgId) {
		return nil, errorsmod.Wrapf(ErrInvalidOrgId, "org with id %d doesn't exist", msg.UserOrgId)
	}

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

	//Updating the subnet
	subnet.UserOrgId = msg.UserOrgId
	k.SetSubnet(ctx, subnet)

	return &types.MsgUpdateSubnetUserResponse{}, nil
}
