package types

import (
	"net/url"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateOrg{}

func NewMsgUpdateOrg(creator string, id uint64, admin string, name string, legalName string, address string, registrationNumber string, countryCode string, taxId string, website string, abuseEmail string, contactEmail string) *MsgUpdateOrg {
	return &MsgUpdateOrg{
		Creator:            creator,
		Id:                 id,
		Name:               name,
		LegalName:          legalName,
		Address:            address,
		RegistrationNumber: registrationNumber,
		CountryCode:        countryCode,
		TaxId:              taxId,
		Website:            website,
		AbuseEmail:         abuseEmail,
		ContactEmail:       contactEmail,
	}
}

func (msg *MsgUpdateOrg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
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
		return errorsmod.Wrapf(ErrNothingToUpdate, "at least one field must be provided for update")
	}

	// Name validation - allow letters, numbers, dots, and hyphens, 3-50 characters
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9.-]{3,50}$`)
	if msg.Name != "" && !nameRegex.MatchString(msg.Name) {
		return errorsmod.Wrap(ErrInvalidName, "name must be 3-50 characters and can only contain letters, numbers, dots, and hyphens")
	}

	// Legal Name validation - allow letters, numbers, spaces, and common punctuation, 3-100 characters
	legalNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.,&'-]{3,100}$`)
	if msg.LegalName != "" && !legalNameRegex.MatchString(msg.LegalName) {
		return errorsmod.Wrap(ErrInvalidLegalName, "legal name must be 3-100 characters and can contain letters, numbers, spaces, and basic punctuation")
	}

	// Address validation - allow common address characters, 5-200 characters
	addressRegex := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]{5,200}$`)
	if msg.Address != "" && !addressRegex.MatchString(msg.Address) {
		return errorsmod.Wrap(ErrInvalidAddress, "address must be 5-200 characters and can contain letters, numbers, spaces, commas, dots, and hyphens")
	}

	// Registration Number validation - allow alphanumeric and hyphens, 5-20 characters
	regNumRegex := regexp.MustCompile(`^[A-Z0-9-]{5,20}$`)
	if msg.RegistrationNumber != "" && !regNumRegex.MatchString(msg.RegistrationNumber) {
		return errorsmod.Wrap(ErrInvalidRegistrationNumber, "registration number must be 5-20 characters and can only contain capital letters, numbers, and hyphens")
	}

	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if msg.CountryCode != "" && !countryCodeRegex.MatchString(msg.CountryCode) {
		return errorsmod.Wrap(ErrInvalidCountryCode, "country code must be exactly 2 capital letters")
	}

	parsedUrl, err := url.Parse(msg.Website)
	if msg.Website != "" && err != nil {
		return errorsmod.Wrapf(ErrInvalidWebsiteUrl, "invalid geofeed URL format (%s)", msg.Website)
	}

	// Check if host is present
	if msg.Website != "" && parsedUrl.Host == "" {
		return errorsmod.Wrapf(ErrInvalidWebsiteUrl, "website URL must have a valid host (%s)", msg.Website)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if msg.AbuseEmail != "" && !emailRegex.MatchString(msg.AbuseEmail) {
		return errorsmod.Wrap(ErrInvalidEmail, "invalid abuse email")
	}

	if msg.ContactEmail != "" && !emailRegex.MatchString(msg.ContactEmail) {
		return errorsmod.Wrap(ErrInvalidEmail, "invalid contact email")
	}
	return nil
}
