package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRegisterOrg{}

func NewMsgRegisterOrg(creator string, name string, legalName string, address string, registrationNumber string, countryCode string, taxId string, website string, abuseEmail string, contactEmail string, treasury string) *MsgRegisterOrg {
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

	return nil
}
