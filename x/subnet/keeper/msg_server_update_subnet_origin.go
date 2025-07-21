package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateSubnetOrigin(goCtx context.Context, msg *types.MsgUpdateSubnetOrigin) (*types.MsgUpdateSubnetOriginResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	//Checking reserved AS numbers
	//0	Reserved by [RFC7607]	[RFC7607]
	//112	Used by the AS112 project to sink misdirected DNS queries; see [RFC7534]	[RFC7534]
	//23456	AS_TRANS; reserved by [RFC6793]	[RFC6793]
	//64496-64511	For documentation and sample code; reserved by [RFC5398]	[RFC5398]
	//64512-65534	For private use; reserved by [RFC6996]	[RFC6996]
	//65535	Reserved by [RFC7300]	[RFC7300]
	//65536-65551	For documentation and sample code; reserved by [RFC5398]	[RFC5398]
	//4200000000-4294967294	For private use; reserved by [RFC6996]	[RFC6996]
	//4294967295	Reserved by [RFC7300]	[RFC7300]

	if msg.As == 0 || msg.As == 112 || msg.As == 23456 || msg.As == 65535 || msg.As == 4294967295 ||
		(msg.As >= 64496 && msg.As <= 64511) ||
		(msg.As >= 65536 && msg.As <= 65551) ||
		(msg.As >= 64512 && msg.As <= 65534) ||
		(msg.As >= 4200000000 && msg.As <= 4294967294) {
		return nil, errorsmod.Wrapf(ErrReservedASNumber, "reserved AS number %d", msg.As)
	}

	if msg.As <= 0 || msg.As > 4294967295 {
		return nil, errorsmod.Wrapf(ErrASNumberOutOfRange, "AS number %d out of range", msg.As)
	}

	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	//Checking if message creator is an admin of the subnet user org
	//NB only user can request origin change
	if subnet.UserOrgId > 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.UserOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.UserOrgId)
	}

	// If there is no specific authorization to allow origin transfer anywhere we set the check to pending
	if !subnet.AllowUserASNDROPOriginTransfer{
		subnet.OriginCheckPending = true
	}
	
	subnet.OriginAsn = msg.As
	currentTime := ctx.BlockTime()

	subnet.OriginTransferRequestedAt = &currentTime
	subnet.RpkiOriginValid = false
	subnet.RpkiValidityChangedAt = &currentTime

	subnet.UpdatedAt = currentTime
	k.SetSubnet(ctx, subnet)
	return &types.MsgUpdateSubnetOriginResponse{}, nil
}
