package keeper

import (
	"context"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateSubnetGeolocation(goCtx context.Context, msg *types.MsgUpdateSubnetGeolocation) (*types.MsgUpdateSubnetGeolocationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	//Checking that the provided values are not nil
	if msg.Country == "" || msg.Region == "" || msg.City == "" || msg.PostalCode == "" {
		return nil, errorsmod.Wrapf(ErrInvalidGeolocationData, "Geolocation data can not be empty")

	}
	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	//Owner can't change geolocation if there's a user
	//User can change geo anytime
	if subnet.UserOrgId == 0 && !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		//Subnet don't have a user and the msg.Creator is not an admin of owner org
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of owner org %d", msg.Creator, subnet.OwnerOrgId)
	} else if !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.UserOrgId, msg.Creator) {
		//Subnet has a user org but the msg.Creator is not an admin of the user org
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of user org %d", msg.Creator, subnet.UserOrgId)
	}

	//TODO: input field validation. We'd need to use a common source of truth for both on chain and off chain validation for geolocation data
	//Only things to validate are Country: a 2-letter ISO country code conforming to ISO 3166-1.
	//Region: a region code conforming to ISO 3166-2. For example, US-TX is the ISO 3166-2 code for Texas in the United States.
	//Applying the changes
	//currentTime := ctx.BlockTime()
	//subnet.GeoChangeRequested = true
	//subnet.geoDetailsChangedAt = &currentTime
	subnet.Country = msg.Country
	subnet.Region = msg.Region
	subnet.City = msg.City
	subnet.PostalCode = msg.PostalCode
	subnet.UpdatedAt = ctx.BlockTime()
	k.SetSubnet(ctx, subnet)

	return &types.MsgUpdateSubnetGeolocationResponse{}, nil
}
