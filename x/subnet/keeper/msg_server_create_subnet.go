package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateSubnet(goCtx context.Context, msg *types.MsgCreateSubnet) (*types.MsgCreateSubnetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check for existing network with the same network address
	subnets := k.GetAllSubnet(ctx)
	for _, existingSubnet := range subnets {
		if existingSubnet.NetworkAddress == msg.NetworkAddress {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrConflict,
				"contract already exists for network address %s and is not terminated",
				msg.NetworkAddress,
			)
		}
	}

	return &types.MsgCreateSubnetResponse{}, nil
}
