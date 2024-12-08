package keeper

import (
	"context"
	"fmt"
	"strconv"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateOrg(goCtx context.Context, msg *types.MsgUpdateOrg) (*types.MsgUpdateOrgResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	org, found := k.GetOrg(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if msg.Creator != org.Admin {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	//Keeping track of changes
	var updatedFields []sdk.Attribute

	//Setting trasferred parameters
 if msg.Admin != "" {
        org.Admin = msg.Admin
        updatedFields = append(updatedFields, sdk.NewAttribute(types.OrgUpdatedEventAdmin, msg.Admin))
    }
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

	k.SetOrg(
		ctx,
		org,
	)
    // Always include these base attributes
    baseAttributes := []sdk.Attribute{
        sdk.NewAttribute(types.OrgUpdatedEventCreator, msg.Creator),
        sdk.NewAttribute(types.OrgUpdatedEventId, strconv.FormatUint(msg.Id, 10)),
    }

    // Combine base attributes with updated fields
    allAttributes := append(baseAttributes, updatedFields...)

    // Emit the organization update event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(types.OrgUpdatedEventType, allAttributes...),
    )
	return &types.MsgUpdateOrgResponse{}, nil
}
