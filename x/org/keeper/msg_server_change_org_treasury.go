package keeper

import (
	"context"
	"strconv"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) ChangeOrgTreasury(goCtx context.Context, msg *types.MsgChangeOrgTreasury) (*types.MsgChangeOrgTreasuryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate new treasury address
	_, err = sdk.AccAddressFromBech32(msg.NewTreasury)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid treasury address: %s", err)
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
	if org.Treasury == msg.NewTreasury {
		return nil, errorsmod.Wrap(types.ErrNoChange, "same treasury address already set up")
	}

	// Change treasury
	org.Treasury = msg.NewTreasury
	org.UpdatedAt = ctx.BlockTime()

	// Set the changes
	k.SetOrg(ctx, org)

	// Emit the treasury change  event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.OrgTreasuryChangedEventType,
			sdk.NewAttribute(types.OrgTreasuryChangedCreator, msg.Creator),
			sdk.NewAttribute(types.OrgTreasuryChangedId, strconv.FormatUint(msg.OrgId, 10)),
			sdk.NewAttribute(types.OrgTreasuryChangedNewTreasury, msg.NewTreasury),
		),
	)
	return &types.MsgChangeOrgTreasuryResponse{}, nil
}
