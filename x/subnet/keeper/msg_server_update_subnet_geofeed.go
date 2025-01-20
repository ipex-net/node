package keeper

import (
	"context"
	"net/url"
	"strings"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateSubnetGeofeed(goCtx context.Context, msg *types.MsgUpdateSubnetGeofeed) (*types.MsgUpdateSubnetGeofeedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	//Validate Geofeed
	if msg.Geofeed == "" {
		return nil, errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL not provided")
	}

	// Check if URL starts with http:// or https://
	if strings.HasPrefix(strings.ToLower(msg.Geofeed), "http://") ||
		strings.HasPrefix(strings.ToLower(msg.Geofeed), "https://") {
		return nil, errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL should not include http:// or https://")
	}

	//Check if URL parses
	parsedUrl, err := url.Parse("https://" + msg.Geofeed)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidGeofeedURL, "invalid geofeed URL format (%s)", msg.Geofeed)
	}

	// Check if host is present
	if parsedUrl.Host == "" {
		return nil, errorsmod.Wrapf(ErrInvalidGeofeedURL, "geofeed URL must have a valid host (%s)", msg.Geofeed)
	}

	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	//Checking if message creator is subnet owner
	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, subnet.OwnerOrgId)
	}
	subnet.GeofeedUrl = msg.Geofeed
	k.SetSubnet(ctx, subnet)

	return &types.MsgUpdateSubnetGeofeedResponse{}, nil
}
