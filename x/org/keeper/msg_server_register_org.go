package keeper

import (
	"context"

	"ipex/x/org/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RegisterOrg(goCtx context.Context, msg *types.MsgRegisterOrg) (*types.MsgRegisterOrgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	//TODO: Email validation, website validation, country validation
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
		Verified:           false, // Setting default verification status
	}
	id := k.AppendOrg(
		ctx,
		org,
	)
	return &types.MsgRegisterOrgResponse{
		Id: id,
	}, nil
}
