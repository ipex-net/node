package keeper

import (
	"context"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RegisterOrg(goCtx context.Context, msg *types.MsgRegisterOrg) (*types.MsgRegisterOrgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var org = types.Org{
		Creator:            msg.Creator,
		Name:               msg.Name,
		Admin:              msg.Creator,
		LegalName:          msg.LegalName,
		Address:            msg.Address,
		RegistrationNumber: msg.RegistrationNumber,
		CountryCode:        msg.CountryCode,
		TaxId:              msg.TaxId,
		Website:            msg.Website,
		AbuseEmail:         msg.AbuseEmail,
		ContactEmail:       msg.ContactEmail,
		DomainVerified:     false,
		Treasury:           msg.Creator,
		CreatedAt:          ctx.BlockTime(),
		UpdatedAt:          ctx.BlockTime(),
	}

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if address is already a creator of any org
	_, err = k.OrgsByCreator.Get(ctx, msg.Creator)
	if err == nil {
		return nil, errorsmod.Wrapf(types.ErrAlreadyAdmin, "address %s is already creator of an organization", msg.Creator)
	}

	// Check if address is already an admin of any org
	_, err = k.OrgsByAdmin.Get(ctx, msg.Creator)
	if err == nil {
		return nil, errorsmod.Wrapf(types.ErrAlreadyAdmin, "address %s is already admin of an organization", msg.Creator)
	}

	var missingFields []string

	if msg.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if msg.LegalName == "" {
		missingFields = append(missingFields, "legal_name")
	}
	if msg.Address == "" {
		missingFields = append(missingFields, "address")
	}
	if msg.RegistrationNumber == "" {
		missingFields = append(missingFields, "registration_number")
	}
	if msg.CountryCode == "" {
		missingFields = append(missingFields, "country_code")
	}
	if msg.TaxId == "" {
		missingFields = append(missingFields, "tax_id")
	}
	if msg.Website == "" {
		missingFields = append(missingFields, "website")
	}
	if msg.AbuseEmail == "" {
		missingFields = append(missingFields, "abuse_email")
	}
	if msg.ContactEmail == "" {
		missingFields = append(missingFields, "contact_email")
	}

	if len(missingFields) > 0 {
		return nil, errorsmod.Wrapf(types.ErrRequiredFields, "missing required fields: %s", strings.Join(missingFields, ", "))
	}

	//Validating the format
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.-]{3,50}$`)
	if !nameRegex.MatchString(msg.Name) {
		return nil, errorsmod.Wrap(types.ErrInvalidName, "name must be 3-50 characters and can only contain letters, numbers, spaces, dots, and hyphens")
	}

	legalNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.,&'-]{3,100}$`)
	if !legalNameRegex.MatchString(msg.LegalName) {
		return nil, errorsmod.Wrap(types.ErrInvalidLegalName, "legal name must be 3-100 characters and can contain letters, numbers, spaces, and basic punctuation")
	}

	addressRegex := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]{5,200}$`)
	if !addressRegex.MatchString(msg.Address) {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "address must be 5-200 characters and can contain letters, numbers, spaces, commas, dots, and hyphens")
	}

	regNumRegex := regexp.MustCompile(`^[A-Z0-9-]{5,20}$`)
	if !regNumRegex.MatchString(msg.RegistrationNumber) {
		return nil, errorsmod.Wrap(types.ErrInvalidRegistrationNumber, "registration number must be 5-20 characters and can only contain capital letters, numbers, and hyphens")
	}
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !countryCodeRegex.MatchString(msg.CountryCode) {
		return nil, errorsmod.Wrap(types.ErrInvalidCountryCode, "country code must be exactly 2 capital letters")
	}

	parsedUrl, err := url.Parse(msg.Website)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidWebsiteUrl, "invalid geofeed URL format (%s)", msg.Website)
	}

	// Check if host is present
	if parsedUrl.Host == "" {
		return nil, errorsmod.Wrapf(types.ErrInvalidWebsiteUrl, "website URL must have a valid host (%s)", msg.Website)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(msg.AbuseEmail) {
		return nil, errorsmod.Wrap(types.ErrInvalidEmail, "invalid abuse email")
	}

	if !emailRegex.MatchString(msg.ContactEmail) {
		return nil, errorsmod.Wrap(types.ErrInvalidEmail, "invalid contact email")
	}

	id := k.AppendOrg(
		ctx,
		org,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.OrgRegisteredEventType,
			sdk.NewAttribute(types.OrgRegisteredEventCreator, msg.Creator),
			sdk.NewAttribute(types.OrgRegisteredEventId, strconv.FormatUint(id, 10)),
			sdk.NewAttribute(types.OrgRegisteredEventName, msg.Name),
			sdk.NewAttribute(types.OrgRegisteredEventLegalName, msg.LegalName),
			sdk.NewAttribute(types.OrgRegisteredEventAddress, msg.Address),
			sdk.NewAttribute(types.OrgRegisteredEventRegistrationNumber, msg.RegistrationNumber),
			sdk.NewAttribute(types.OrgRegisteredEventCountryCode, msg.CountryCode),
			sdk.NewAttribute(types.OrgRegisteredEventTaxId, msg.TaxId),
			sdk.NewAttribute(types.OrgRegisteredEventWebsite, msg.Website),
			sdk.NewAttribute(types.OrgRegisteredEventAbuseEmail, msg.AbuseEmail),
			sdk.NewAttribute(types.OrgRegisteredEventContactEmail, msg.ContactEmail),
		),
	)

	return &types.MsgRegisterOrgResponse{
		Id: id,
	}, nil
}
