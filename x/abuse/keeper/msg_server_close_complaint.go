package keeper

import (
	"context"

	"ipex/x/abuse/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CloseComplaint(goCtx context.Context, msg *types.MsgCloseComplaint) (*types.MsgCloseComplaintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	complaint, found := k.GetComplaint(ctx, msg.ComplaintId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrComplaintNotFound,
			"can not find complaint with id %d",
			msg.ComplaintId)
	}

	if complaint.State == types.ComplaintState_STATE_CLOSED {
		return nil, errorsmod.Wrap(types.ErrComplaintClosed,
			"complaint already closed")
	}

	if !k.subnetKeeper.VerifySubnetOwner(ctx, complaint.SubnetId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorized,
			"address %s is not authorized to close complaint with id %d",
			msg.Creator,
			msg.ComplaintId)
	}

	currentTime := ctx.BlockTime()

	// Updating abuse stats
	if complaint.ResolutionFailed {
		k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "failed", "resolved")
	} else {
		k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "inprogress", "resolved")
	}

	// Update complaint state and activity tracking
	complaint.State = types.ComplaintState_STATE_CLOSED

	// Unset failed resolution flag
	complaint.ResolutionFailed = false

	complaint.LastActivityAt = &currentTime

	// Save the updated complaint
	k.SetComplaint(ctx, complaint)

	// Emit event for response
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"abuse_complaint_closed",
			sdk.NewAttribute("complaint_id", string(rune(msg.ComplaintId))),
			sdk.NewAttribute("responder", msg.Creator),
			sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
		),
	)

	return &types.MsgCloseComplaintResponse{}, nil
}
