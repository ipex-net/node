package keeper

import (
	"context"
	"time"

	"ipex/x/abuse/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)
const closeAfterHours = 48

// Automatically closes or sets resolution as failed based on time passed and lease state
func (k Keeper) ProcessComplaints(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Only run every 50 blocks
	if ctx.BlockHeight()%50 != 0 { 
		return
		}        
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(goCtx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()


	currentTime := ctx.BlockTime()
	for ; iterator.Valid(); iterator.Next() {
		var complaint types.Complaint
		k.cdc.MustUnmarshal(iterator.Value(), &complaint)
		// Closing complaints that were replied 48 hours ago and haven't failed resolution
		if !complaint.ResolutionFailed && complaint.State == types.ComplaintState_STATE_RESPONDED && currentTime.After(complaint.LastActivityAt.Add(closeAfterHours * time.Hour)) {
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
		// Setting resolution to failed if the deadline was missed in awaiting reponse or submitted states
		if !complaint.ResolutionFailed && (complaint.State == types.ComplaintState_STATE_AWAITING_RESPONSE || complaint.State == types.ComplaintState_STATE_SUBMITTED) && currentTime.After(*complaint.ResolutionDeadline) {
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
		// Closing all non-failed complaints if lessor org is not the user of subnet, and therefore can't reply
		if !complaint.ResolutionFailed && complaint.State != types.ComplaintState_STATE_CLOSED {
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
	}
}

