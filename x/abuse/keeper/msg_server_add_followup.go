package keeper

import (
	"context"
	"regexp"
	"strings"
	"time"

	"ipex/x/abuse/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AddFollowup(goCtx context.Context, msg *types.MsgAddFollowup) (*types.MsgAddFollowupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Content validation
	minContentLength := 20
	maxContentLength := 200
	// Clean content
	content := strings.TrimSpace(msg.Content)
	content = strings.Join(strings.Fields(content), " ")

	// Validate content size
	if len(content) < minContentLength {
		return nil, errorsmod.Wrapf(types.ErrInvalidSummary,
			"summary too short (minimum %d characters)",
			minContentLength)
	}

	if len(content) > maxContentLength {
		return nil, errorsmod.Wrapf(types.ErrInvalidSummary,
			"summary too long (maximum %d characters)",
			maxContentLength)
	}

	// Validate evidence hash (IPFS)
	evidenceHash := strings.TrimSpace(msg.EvidenceHash)

	// Evidence hash is optional - complaints can be submitted without evidence
	if len(evidenceHash) > 0 {
		// IPFS hashes can be in two formats:
		// - CIDv0: starts with "Qm" (46 chars, base58)
		// - CIDv1: starts with "bafy" or "bafk" (59+ chars, base32)

		// Basic validation - check if it looks like an IPFS hash
		validIPFSv0 := regexp.MustCompile(`^Qm[1-9A-HJ-NP-Za-km-z]{44}$`)
		validIPFSv1 := regexp.MustCompile(`^ba[a-z2-7]{57,}$`)

		if !validIPFSv0.MatchString(evidenceHash) && !validIPFSv1.MatchString(evidenceHash) {
			return nil, errorsmod.Wrapf(types.ErrInvalidEvidenceHash,
				"evidence hash does not appear to be a valid IPFS hash")
		}
	}

	complaint, found := k.GetComplaint(ctx, msg.ComplaintId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrComplaintNotFound,
			"can not find complaint for id %d",
			msg.ComplaintId)
	}

	if complaint.State == types.ComplaintState_STATE_CLOSED {
		return nil, errorsmod.Wrap(types.ErrComplaintClosed,
			"complaint already closed")
	}

	if !k.subnetKeeper.VerifySubnetOwner(ctx, complaint.SubnetId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrUnauthorized,
			"address %s is not authorized to submit lessor follow up for complaint id %d",
			msg.Creator,
			msg.ComplaintId)
	}

	currentTime := ctx.BlockTime()
	var threadMessage = types.ThreadMessage{
		MessageType:  types.MessageType_MESSAGE_TYPE_LESSOR_FOLLOWUP,
		Content:      content,
		EvidenceHash: evidenceHash,
		SentAt:       &currentTime,
	}

	// Add the response to the complaint thread
	complaint.Thread = append(complaint.Thread, &threadMessage)

	// Update complaint state and activity tracking
	complaint.State = types.ComplaintState_STATE_AWAITING_RESPONSE

	// Extend deadline if not already exceeded
	if !complaint.ResolutionFailed && currentTime.Before(*complaint.ResolutionDeadline) {
		// Calculate new deadline: current time + 24 hours
		newDeadline := currentTime.Add(24 * time.Hour)

		// Only update if the new deadline is later than the existing one
		// This ensures we always use the furthest deadline
		if newDeadline.After(*complaint.ResolutionDeadline) {
			complaint.ResolutionDeadline = &newDeadline
		}
	}

	complaint.LastActivityAt = &currentTime

	// Save the updated complaint
	k.SetComplaint(ctx, complaint)

	// Emit event for response
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"abuse_follow_up_added",
			sdk.NewAttribute("complaint_id", string(rune(msg.ComplaintId))),
			sdk.NewAttribute("responder", msg.Creator),
			sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
		),
	)

	return &types.MsgAddFollowupResponse{}, nil
}
