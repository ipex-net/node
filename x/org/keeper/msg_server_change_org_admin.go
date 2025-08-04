package keeper

import (
	"context"
	"strconv"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ChangeOrgAdmin(goCtx context.Context, msg *types.MsgChangeOrgAdmin) (*types.MsgChangeOrgAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate new admin address format
	_, err = sdk.AccAddressFromBech32(msg.NewAdmin)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress,"invalid admin address: %s", err)
	}

	// Fetch the org
	org, found := k.GetOrg(ctx, msg.OrgId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrOrgNotFound, "org with id %d doesn't exist", msg.OrgId)
	}

	// Check permissions
	if msg.Creator != org.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect org creator")
	}

	// Reject if no change
	if org.Admin == msg.NewAdmin {
		return nil, types.ErrAlreadyAdmin
	}

	// Reject if address is already admin of a DIFFERENT org
	existingOrgId, err := k.OrgsByAdmin.Get(ctx, msg.NewAdmin)
	if err == nil && existingOrgId != msg.OrgId {
		return nil, errorsmod.Wrapf(types.ErrAlreadyAdmin, "address %s is already admin of organization %d", msg.NewAdmin, existingOrgId)
	}

	// Reject if address is creator of a DIFFERENT org
	existingCreatorOrgId, err := k.OrgsByCreator.Get(ctx, msg.NewAdmin)
	if err == nil && existingCreatorOrgId != msg.OrgId {
		return nil, errorsmod.Wrapf(types.ErrAlreadyAdmin, "address %s is already creator of organization %d", msg.NewAdmin, existingCreatorOrgId)
	}

	// Change admin
	org.Admin = msg.NewAdmin
	org.UpdatedAt = ctx.BlockTime()

	// Set the changes
	k.SetOrg(ctx,org)

	// Emit the admin change  event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.OrgAdminChangedEventType,
			sdk.NewAttribute(types.OrgAdminChangedCreator, msg.Creator),
			sdk.NewAttribute(types.OrgAdminChangedId, strconv.FormatUint(msg.OrgId, 10)),
			sdk.NewAttribute(types.OrgAdminChangedNewAdmin, msg.NewAdmin),
		),
	)
	return &types.MsgChangeOrgAdminResponse{}, nil
}
