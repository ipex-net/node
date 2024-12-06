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
