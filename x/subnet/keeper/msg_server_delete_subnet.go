package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DeleteSubnet(goCtx context.Context, msg *types.MsgDeleteSubnet) (*types.MsgDeleteSubnetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Id == 0 {
		return nil, errorsmod.Wrap(types.ErrSubnetNotFound, "subnet id not provided",)
	}

	// fetch the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	// check that the message creator is subnet owner
	if subnet.OwnerOrgId > 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.OwnerOrgId)
	}

	// check that subnet is not active on the marketplace
	if subnet.Status == types.SubnetStatus_SUBNET_STATUS_ACTIVE {
		return nil, errorsmod.Wrapf(types.ErrIncorrectSubnetStatus, "can not delete subnet in ACTIVE status")
	}

	// check that subnet doesn't have a user
	if subnet.UserOrgId != 0 {
		return nil, errorsmod.Wrapf(types.ErrIncorrectSubnetStatus, "can not delete subnet with assigned user")
	}
	
	k.RemoveSubnet(ctx, msg.Id)
	return &types.MsgDeleteSubnetResponse{}, nil
}
