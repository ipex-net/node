package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		Treasury:			treasury,
	}
}

func (msg *MsgRegisterOrg) ValidateBasic() error {




	return nil
}
