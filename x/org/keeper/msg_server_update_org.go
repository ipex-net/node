package keeper

import (
	"context"
	"net/url"
	"regexp"
	"strconv"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateOrg(goCtx context.Context, msg *types.MsgUpdateOrg) (*types.MsgUpdateOrgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// First check if all fields are empty
	if msg.Name == "" &&
		msg.LegalName == "" &&
		msg.Address == "" &&
		msg.RegistrationNumber == "" &&
		msg.CountryCode == "" &&
		msg.TaxId == "" &&
		msg.Website == "" &&
		msg.AbuseEmail == "" &&
		msg.ContactEmail == "" {
		return nil, errorsmod.Wrapf(types.ErrNothingToUpdate, "at least one field must be provided for update")
	}

	// Name validation - allow letters, numbers, dots, and hyphens, 3-50 characters
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.-]{3,50}$`)
	if !nameRegex.MatchString(msg.Name) {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "name must be 3-50 characters and can only contain letters, numbers, spaces, dots, and hyphens")
	}

	// Legal Name validation - allow letters, numbers, spaces, and common punctuation, 3-100 characters
	legalNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.,&'-]{3,100}$`)
	if msg.LegalName != "" && !legalNameRegex.MatchString(msg.LegalName) {
		return nil, errorsmod.Wrap(types.ErrInvalidLegalName, "legal name must be 3-100 characters and can contain letters, numbers, spaces, and basic punctuation")
	}

	// Address validation - allow common address characters, 5-200 characters
	addressRegex := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]{5,200}$`)
	if msg.Address != "" && !addressRegex.MatchString(msg.Address) {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "address must be 5-200 characters and can contain letters, numbers, spaces, commas, dots, and hyphens")
	}

	// Registration Number validation - allow alphanumeric and hyphens, 5-20 characters
	regNumRegex := regexp.MustCompile(`^[A-Z0-9-]{5,20}$`)
	if msg.RegistrationNumber != "" && !regNumRegex.MatchString(msg.RegistrationNumber) {
		return nil, errorsmod.Wrap(types.ErrInvalidRegistrationNumber, "registration number must be 5-20 characters and can only contain capital letters, numbers, and hyphens")
	}

	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if msg.CountryCode != "" && !countryCodeRegex.MatchString(msg.CountryCode) {
		return nil, errorsmod.Wrap(types.ErrInvalidCountryCode, "country code must be exactly 2 capital letters")
	}

	parsedUrl, err := url.Parse(msg.Website)
	if msg.Website != "" && err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidWebsiteUrl, "invalid geofeed URL format (%s)", msg.Website)
	}

	// Check if host is present
	if msg.Website != "" && parsedUrl.Host == "" {
		return nil, errorsmod.Wrapf(types.ErrInvalidWebsiteUrl, "website URL must have a valid host (%s)", msg.Website)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if msg.AbuseEmail != "" && !emailRegex.MatchString(msg.AbuseEmail) {
		return nil, errorsmod.Wrapf(types.ErrInvalidEmail, "invalid abuse email (%s)", msg.AbuseEmail)
	}

	if msg.ContactEmail != "" && !emailRegex.MatchString(msg.ContactEmail) {
		return nil, errorsmod.Wrapf(types.ErrInvalidEmail, "invalid contact email (%s)", msg.ContactEmail)
	}

	// Fetch the org
	org, found := k.GetOrg(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrOrgNotFound, "org with id %d doesn't exist", msg.Id)
	}

	// Check permissions
	if msg.Creator != org.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect org creator")
	}

	// Keep track of changes
	var updatedFields []sdk.Attribute

	// Set trasferred parameters
	if msg.LegalName != "" {
		org.LegalName = msg.LegalName
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventLegalName, msg.LegalName))
	}
	if msg.Address != "" {
		org.Address = msg.Address
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventAddress, msg.Address))
	}
	if msg.RegistrationNumber != "" {
		org.RegistrationNumber = msg.RegistrationNumber
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventRegistrationNumber, msg.RegistrationNumber))
	}
	if msg.CountryCode != "" {
		org.CountryCode = msg.CountryCode
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventCountryCode, msg.CountryCode))
	}
	if msg.TaxId != "" {
		org.TaxId = msg.TaxId
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventTaxId, msg.TaxId))
	}
	if msg.Website != "" {
		org.Website = msg.Website
		org.DomainVerified = false
		org.DomainVerifAttemptedAt = nil
		org.DomainVerifiedAt = nil
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventWebsite, msg.Website))
	}
	if msg.AbuseEmail != "" {
		org.AbuseEmail = msg.AbuseEmail
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventAbuseEmail, msg.AbuseEmail))
	}
	if msg.ContactEmail != "" {
		org.ContactEmail = msg.ContactEmail
		updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventContactEmail, msg.ContactEmail))
	}

	org.UpdatedAt = ctx.BlockTime()

	// Set the changes
	k.SetOrg(
		ctx,
		org,
	)

	// Set base event attributes
	baseAttributes := []sdk.Attribute{
		sdk.NewAttribute(types.OrgUpdatedEventCreator, msg.Creator),
		sdk.NewAttribute(types.OrgUpdatedEventId, strconv.FormatUint(msg.Id, 10)),
	}

	// Append the changed attributes
	allAttributes := append(baseAttributes, updatedFields...)

	// Emit the organization update event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.OrgUpdatedEventType, allAttributes...),
	)
	return &types.MsgUpdateOrgResponse{}, nil
}
