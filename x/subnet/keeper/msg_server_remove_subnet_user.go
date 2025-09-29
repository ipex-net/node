package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RemoveSubnetUser removes the user of the subnet outside of the marketplace
func (k msgServer) RemoveSubnetUser(goCtx context.Context, msg *types.MsgRemoveSubnetUser) (*types.MsgRemoveSubnetUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Id == 0 {
		return nil, errorsmod.Wrap(types.ErrSubnetNotFound, "subnet id not provided",)
	}

	// Fetch the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	// check that subnet is not active on the marketplace
	if subnet.Status == types.SubnetStatus_SUBNET_STATUS_ACTIVE {
		return nil, errorsmod.Wrapf(types.ErrIncorrectSubnetStatus, "can not remove user for subnet in active state")
	}

	// Verify creator is org admin
	if subnet.OwnerOrgId > 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.OwnerOrgId)
	}

	subnet.OriginAsn = 0
	subnet.OriginTransferRequestedAt = nil
	subnet.RpkiValidityChangedAt = nil
	subnet.RpkiOriginValid = false
	subnet.UpdatedAt = ctx.BlockTime()
	//Updating the subnet
	subnet.UserOrgId = 0
	k.SetSubnet(ctx, subnet)

	return &types.MsgRemoveSubnetUserResponse{}, nil
}
