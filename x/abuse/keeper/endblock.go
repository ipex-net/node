package keeper

import (
	"context"
	"time"

	"ipex/x/abuse/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TESTING
// const autoCloseDelay = 48 * time.Hour
const autoCloseDelay = 5 * time.Minute

// Automatically closes or sets resolution as failed based on time passed and lease state
func (k Keeper) ProcessComplaints(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TESTING
	// // Only run every 50 blocks
	// if ctx.BlockHeight()%50 != 0 {
	// 	return
	// 	}

	currentTime := ctx.BlockTime()

	// Process RESPONDED complaints that need auto-closing
	k.processComplaintsByState(goCtx, types.ComplaintState_STATE_RESPONDED, func(complaint types.Complaint) {
		if !complaint.ResolutionFailed && currentTime.After(complaint.LastActivityAt.Add(autoCloseDelay)) {
			// Update complaint state
			complaint.State = types.ComplaintState_STATE_CLOSED
			complaint.ResolutionFailed = false
			complaint.LastActivityAt = &currentTime
			k.SetComplaint(goCtx, complaint)
			k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "inprogress", "resolved")
			// Emit event for response
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"abuse_complaint_closed_automatically",
					sdk.NewAttribute("complaint_id", string(rune(complaint.Id))),
					sdk.NewAttribute("lease_id", string(rune(complaint.LeaseId))),
					sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
					sdk.NewAttribute("reason", "no_followup"),
				),
			)
		}
	})

	// Process AWAITING_RESPONSE complaints for deadline failures
	k.processComplaintsByState(goCtx, types.ComplaintState_STATE_AWAITING_RESPONSE, func(complaint types.Complaint) {
		if !complaint.ResolutionFailed && currentTime.After(*complaint.ResolutionDeadline) {
			complaint.ResolutionFailed = true
			k.SetComplaint(goCtx, complaint)
			k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "inprogress", "failed")

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"abuse_failed_timely_resolution",
					sdk.NewAttribute("complaint_id", string(rune(complaint.Id))),
					sdk.NewAttribute("lease_id", string(rune(complaint.LeaseId))),
					sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
				),
			)
		}
	})

	// Process SUBMITTED complaints for deadline failures
	k.processComplaintsByState(goCtx, types.ComplaintState_STATE_SUBMITTED, func(complaint types.Complaint) {
		if !complaint.ResolutionFailed && currentTime.After(*complaint.ResolutionDeadline) {
			complaint.ResolutionFailed = true
			k.SetComplaint(goCtx, complaint)
			k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "inprogress", "failed")

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"abuse_failed_timely_resolution",
					sdk.NewAttribute("complaint_id", string(rune(complaint.Id))),
					sdk.NewAttribute("lease_id", string(rune(complaint.LeaseId))),
					sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
				),
			)
		}
	})

	// Process all non-closed complaints for subnet user validation
	states := []types.ComplaintState{
		types.ComplaintState_STATE_SUBMITTED,
		types.ComplaintState_STATE_RESPONDED,
		types.ComplaintState_STATE_AWAITING_RESPONSE,
	}
	for _, state := range states {
		k.processComplaintsByState(goCtx, state, func(complaint types.Complaint) {
			if !complaint.ResolutionFailed {
				subnet, found := k.subnetKeeper.GetSubnet(ctx, complaint.SubnetId)
				if !found || subnet.UserOrgId != complaint.LesseeOrgId {
					// Update complaint state
					complaint.State = types.ComplaintState_STATE_CLOSED
					complaint.ResolutionFailed = false
					complaint.LastActivityAt = &currentTime
					k.SetComplaint(goCtx, complaint)
					k.trustKeeper.UpdateAbuseStats(ctx, complaint.LeaseId, complaint.LesseeOrgId, uint32(complaint.AbuseType), "inprogress", "resolved")
					ctx.EventManager().EmitEvent(
						sdk.NewEvent(
							"abuse_complaint_closed_automatically",
							sdk.NewAttribute("complaint_id", string(rune(complaint.Id))),
							sdk.NewAttribute("lease_id", string(rune(complaint.LeaseId))),
							sdk.NewAttribute("subnet_id", string(rune(complaint.SubnetId))),
							sdk.NewAttribute("reason", "subnet_user_removed"),
						),
					)
				}
			}
		})
	}
}

// processComplaintsByState processes complaints by their state using the state index
func (k Keeper) processComplaintsByState(ctx context.Context, state types.ComplaintState, processor func(types.Complaint)) {
	// Use state index to get complaints by state efficiently
	rng := collections.NewPrefixedPairRange[int32, uint64](int32(state))
	err := k.ComplaintsByState.Walk(ctx, rng, func(key collections.Pair[int32, uint64], value bool) (bool, error) {
		// Get the complaint ID from the key
		complaintId := key.K2()

		// Fetch the full complaint
		complaint, found := k.GetComplaint(ctx, complaintId)
		if !found {
			// Complaint might have been deleted, skip
			return false, nil
		}

		// Process the complaint
		processor(complaint)
		return false, nil
	})

	if err != nil {
		k.Logger().Error("Error processing complaints by state", "state", state, "error", err)
	}
}
