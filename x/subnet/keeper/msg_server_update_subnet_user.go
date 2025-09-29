package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateSubnetUser sets the user of the subnet outside of the marketplace
func (k msgServer) UpdateSubnetUser(goCtx context.Context, msg *types.MsgUpdateSubnetUser) (*types.MsgUpdateSubnetUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// checking that user org exists
	if msg.UserOrgId == 0 || !k.orgKeeper.HasOrg(ctx, msg.UserOrgId) {
		return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "org with id %d doesn't exist", msg.UserOrgId)
	}

	if msg.Id == 0 {
		return nil, errorsmod.Wrap(types.ErrSubnetNotFound, "subnet id not provided",)
	}

	// fetch the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	// check that subnet is not active on the marketplace
	if subnet.Status == types.SubnetStatus_SUBNET_STATUS_ACTIVE {
		return nil, errorsmod.Wrapf(types.ErrIncorrectSubnetStatus, "can not update user for subnet in active state")
	}

	// check that the message creator is subnet owner
	if subnet.OwnerOrgId > 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.OwnerOrgId)
	}

	// update the subnet
	subnet.UserOrgId = msg.UserOrgId
	subnet.UpdatedAt = ctx.BlockTime()
	k.SetSubnet(ctx, subnet)

	return &types.MsgUpdateSubnetUserResponse{}, nil
}
