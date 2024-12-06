package keeper

import (
	"context"
	"fmt"

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
	//TODO: Email validation, website validation, country validation
	if msg.Admin != "" {org.Admin = msg.Admin}
	if msg.LegalName != "" {org.LegalName = msg.LegalName}
	if msg.Address != "" {org.Address = msg.Address}
	if msg.RegistrationNumber != "" {org.RegistrationNumber = msg.RegistrationNumber}
	if msg.CountryCode != "" {org.CountryCode = msg.CountryCode}
	if msg.TaxId != "" {org.TaxId = msg.TaxId}
	if msg.Website != "" {org.Website = msg.Website}
	if msg.AbuseEmail != "" {org.AbuseEmail = msg.AbuseEmail}
	if msg.ContactEmail != "" {org.ContactEmail = msg.ContactEmail}

	k.SetOrg(
		ctx,
		org,
	)

	return &types.MsgUpdateOrgResponse{}, nil
}
