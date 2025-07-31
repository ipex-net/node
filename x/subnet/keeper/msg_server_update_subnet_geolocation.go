package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/biter777/countries"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateSubnetGeolocation(goCtx context.Context, msg *types.MsgUpdateSubnetGeolocation) (*types.MsgUpdateSubnetGeolocationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check that address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that country code is correct
	if msg.Country == "" {
		return nil, errorsmod.Wrapf(ErrInvalidGeolocationData, "country data can not be empty")
	}
	if len(msg.Country) != 2 {
		return nil, errorsmod.Wrapf(ErrInvalidGeolocationData, "country code must be 2 letters")
	}
	if msg.Region != "" {
		err = validateGeo(msg.Country, msg.Region)
	} else {
		err = validateCountry(msg.Country)
	}
	if err != nil {
		return nil, errorsmod.Wrap(ErrInvalidGeolocationData, err.Error())
	}

	// Fetch the subnet
	subnet, found := k.GetSubnet(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(ErrSubnetNotFound, "can not find subnet with id %v", msg.Id)
	}

	// Check that data is actually different
	if subnet.Country == msg.Country && subnet.Region == msg.Region && subnet.City == msg.City && subnet.PostalCode == msg.PostalCode {
		return nil, errorsmod.Wrapf(ErrInvalidGeolocationData, "geolocation data can not be identical")
	}

	// Check 6 hour geo cooldown
	currentTime := ctx.BlockTime()
	if subnet.GeoDetailsChangedAt != nil && currentTime.Sub(*subnet.GeoDetailsChangedAt) < 6*time.Hour {
		// Calculate remaining time in hours (rounded up)
		remainingTime := (6 * time.Hour) - currentTime.Sub(*subnet.GeoDetailsChangedAt)
		remainingHours := int(remainingTime.Hours()) + 1 // Round up
		return nil, errorsmod.Wrapf(
			ErrGeoCooldownPeriod,
			"geo location cooldown period in effect - must wait %d hours before updating geolocation",
			remainingHours,
		)

	}

	isOwnerAdmin := k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.OwnerOrgId, msg.Creator)
	isUserAdmin := k.orgKeeper.VerifyOrgAdmin(goCtx, subnet.UserOrgId, msg.Creator)

	// Continue if signer is subnet admin and there's no lessor or if the signer is lessor
	if !((subnet.UserOrgId == 0 && isOwnerAdmin) || isUserAdmin) {
		return nil, errorsmod.Wrapf(
			ErrNotOrgAdmin,
			"address %s is not authorised to update geolocation for subnet %d",
			msg.Creator, subnet.Id,
		)
	}

	// Check if geo change is allowed for the user
	if isUserAdmin && !subnet.AllowUserGeoChange {
		return nil, ErrGeoChangeProhibited
	}

	// Update the timestamp
	subnet.GeoDetailsChangedAt = &currentTime
	// Reset the verification flag
	subnet.TraceGeoVerified = false
	// Reset the last check timestamp
	subnet.GeoVerifAttemptedAt = nil
	// Update geo details
	subnet.Country = msg.Country
	subnet.Region = msg.Region
	subnet.City = msg.City
	subnet.PostalCode = msg.PostalCode

	// Update the subnet change timestamp
	subnet.UpdatedAt = ctx.BlockTime()
	k.SetSubnet(ctx, subnet)

	return &types.MsgUpdateSubnetGeolocationResponse{}, nil
}

// validateGeo returns nil on success or an error explaining why it failed.
func validateGeo(alpha2, iso31662 string) error {
	country := countries.ByName(strings.ToUpper(alpha2))
	if !country.IsValid() { // ISO-3166-1 check
		return fmt.Errorf("invalid ISO-3166-1 alpha-2 country code: %s", alpha2)
	}

	sub := countries.SubdivisionCode(strings.ToUpper(iso31662))
	if !sub.IsValid() { // ISO-3166-2 check
		return fmt.Errorf("invalid ISO-3166-2 code: %s", iso31662)
	}
	if sub.Country() != country { // “US-TX” must match “US”
		return fmt.Errorf("%s does not belong to country %s", iso31662, alpha2)
	}
	return nil
}

// validateCountry return nil on success or an error if country invalid
func validateCountry(alpha2 string) error {
	country := countries.ByName(strings.ToUpper(alpha2))
	if !country.IsValid() { // ISO-3166-1 check
		return fmt.Errorf("invalid ISO-3166-1 alpha-2 country code: %s", alpha2)
	}
	return nil
}
