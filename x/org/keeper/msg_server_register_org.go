package keeper

import (
	"context"
	"strconv"

	"ipex/x/org/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		Verified:           false,
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
			sdk.NewAttribute(types.OrgRegisteredEventAdmin, msg.Creator),
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
