package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateOrg{}

func NewMsgUpdateOrg(creator string, id uint64, admin string, name string, legalName string, address string, registrationNumber string, countryCode string, taxId string, website string, abuseEmail string, contactEmail string, treasury string) *MsgUpdateOrg {
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
		Treasury:           treasury,
	}
}
