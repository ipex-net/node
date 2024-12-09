package types

import (
	"net/url"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOrg{}

func NewMsgRegisterOrg(creator string, name string, legalName string, address string, registrationNumber string, countryCode string, taxId string, website string, abuseEmail string, contactEmail string) *MsgRegisterOrg {
	return &MsgRegisterOrg{
		Creator:            creator,
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

func (msg *MsgRegisterOrg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// First check if all fields are set
	if msg.Name == "" ||
		msg.LegalName == "" ||
		msg.Address == "" ||
		msg.RegistrationNumber == "" ||
		msg.CountryCode == "" ||
		msg.TaxId == "" ||
		msg.Website == "" ||
		msg.AbuseEmail == "" ||
		msg.ContactEmail == "" {
		return errorsmod.Wrapf(ErrRequiredFields, "all fields are required for organization registration")
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.-]{3,50}$`)
	if !nameRegex.MatchString(msg.Name) {
		return errorsmod.Wrap(ErrInvalidName, "name must be 3-50 characters and can only contain letters, numbers, spaces, dots, and hyphens")
	}

	legalNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s.,&'-]{3,100}$`)
	if !legalNameRegex.MatchString(msg.LegalName) {
		return errorsmod.Wrap(ErrInvalidLegalName, "legal name must be 3-100 characters and can contain letters, numbers, spaces, and basic punctuation")
	}

	addressRegex := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]{5,200}$`)
	if !addressRegex.MatchString(msg.Address) {
		return errorsmod.Wrap(ErrInvalidAddress, "address must be 5-200 characters and can contain letters, numbers, spaces, commas, dots, and hyphens")
	}

	regNumRegex := regexp.MustCompile(`^[A-Z0-9-]{5,20}$`)
	if !regNumRegex.MatchString(msg.RegistrationNumber) {
		return errorsmod.Wrap(ErrInvalidRegistrationNumber, "registration number must be 5-20 characters and can only contain capital letters, numbers, and hyphens")
	}
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !countryCodeRegex.MatchString(msg.CountryCode) {
		return errorsmod.Wrap(ErrInvalidCountryCode, "country code must be exactly 2 capital letters")
	}

	parsedUrl, err := url.Parse(msg.Website)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidWebsiteUrl, "invalid geofeed URL format (%s)", msg.Website)
	}

	// Check if host is present
	if parsedUrl.Host == "" {
		return errorsmod.Wrapf(ErrInvalidWebsiteUrl, "website URL must have a valid host (%s)", msg.Website)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(msg.AbuseEmail) {
		return errorsmod.Wrap(ErrInvalidEmail, "invalid abuse email")
	}

	if !emailRegex.MatchString(msg.ContactEmail) {
		return errorsmod.Wrap(ErrInvalidEmail, "invalid contact email")
	}

	return nil
}
