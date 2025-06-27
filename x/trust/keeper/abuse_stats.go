package keeper

import (
	"context"
	"fmt"

	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) UpdateAbuseStats(ctx context.Context, leaseId uint64, lesseeOrgId uint64, abuseCategory uint32, oldStatus string, newStatus string) error {
	// Unwrap SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Get block time
	blockTime := sdkCtx.BlockTime()
	if blockTime.IsZero() {
		return fmt.Errorf("invalid block time: cannot be zero")
	}

	if oldStatus != "" && oldStatus != "failed" && oldStatus != "inprogress" && oldStatus != "resolved" {
		return fmt.Errorf("invalid old status %s", oldStatus)
	}
	
	if newStatus != "failed" && newStatus != "inprogress" && newStatus != "resolved" {
		return fmt.Errorf("invalid new status %s", newStatus)
	}

	// Convert abuseType uint to enum
	abuseType := abusetypes.AbuseType(abuseCategory)
	if abuseType < abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA || abuseType > abusetypes.AbuseType_ABUSE_TYPE_OTHER {
		return fmt.Errorf(
			"invalid abuse category %d",
			abuseCategory)
	}
	
	// Get the current lease reputation
	currentLeaseRep, found := k.GetLeaseReputation(ctx, leaseId)
	
	// Initialize lease rep if it doesn't exist
	if !found {
		currentLeaseRep = types.LeaseReputation {
			Index: leaseId,
			LastUpdated: &blockTime,
			AbuseByType: make(map[uint32]*types.AbuseStats),
		}
	}
	
	// Get the current lessee reputation
	currentLesseeRep, found := k.GetLesseeReputation(ctx, lesseeOrgId)
	
	// Initialize lessee rep if it doesn't exist
	if !found {
		currentLesseeRep = types.LesseeReputation {
			Index: lesseeOrgId,
			LastUpdated: &blockTime,
			AbuseByType: make(map[uint32]*types.AbuseTypeStats),
		}
	}
	
	// Initialize maps if they are nil
	if currentLesseeRep.AbuseByType == nil {
		currentLesseeRep.AbuseByType = make(map[uint32]*types.AbuseTypeStats)
	}
	if currentLeaseRep.AbuseByType == nil {
		currentLeaseRep.AbuseByType = make(map[uint32]*types.AbuseStats)
	}

	switch oldStatus {
	case "inprogress":
		if currentLesseeRep.AbuseInProgress > 0 {
			currentLesseeRep.AbuseInProgress--
		}
		if currentLeaseRep.AbuseInProgress > 0 {
			currentLeaseRep.AbuseInProgress--
		}
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok && currentLesseeRep.AbuseByType[abuseCategory].InProgress > 0 {
			currentLesseeRep.AbuseByType[abuseCategory].InProgress--
		}
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok && currentLeaseRep.AbuseByType[abuseCategory].InProgress > 0 {
			currentLeaseRep.AbuseByType[abuseCategory].InProgress--
		}

	case "failed":
		if currentLesseeRep.AbuseFailed > 0 {
			currentLesseeRep.AbuseFailed--
		}
		if currentLeaseRep.AbuseFailed > 0 {
			currentLeaseRep.AbuseFailed--
		}
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok && currentLesseeRep.AbuseByType[abuseCategory].Failed > 0 {
			currentLesseeRep.AbuseByType[abuseCategory].Failed--
		}
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok && currentLeaseRep.AbuseByType[abuseCategory].Failed > 0 {
			currentLeaseRep.AbuseByType[abuseCategory].Failed--
		}

	case "resolved":
		if currentLesseeRep.AbuseResolved > 0 {
			currentLesseeRep.AbuseResolved--
		}
		if currentLeaseRep.AbuseResolved > 0 {
			currentLeaseRep.AbuseResolved--
		}
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok && currentLesseeRep.AbuseByType[abuseCategory].Resolved > 0 {
			currentLesseeRep.AbuseByType[abuseCategory].Resolved--
		}
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok && currentLeaseRep.AbuseByType[abuseCategory].Resolved > 0 {
			currentLeaseRep.AbuseByType[abuseCategory].Resolved--
		}
	}

	// New status also used to initiate the stat maps if they are not set
	switch newStatus {
	case "inprogress":
		currentLesseeRep.AbuseInProgress++
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok {
			currentLesseeRep.AbuseByType[abuseCategory].InProgress++
			currentLesseeRep.AbuseByType[abuseCategory].LastIncident = &blockTime
		} else {
			// Initialize stats for category in case it's not set yet
			currentLesseeRep.AbuseByType[abuseCategory] = &types.AbuseTypeStats{
				InProgress: 1, 
				LastIncident: &blockTime,
			}
		}

		currentLeaseRep.AbuseInProgress++
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok {
			currentLeaseRep.AbuseByType[abuseCategory].InProgress++
			currentLeaseRep.AbuseByType[abuseCategory].LastIncident = &blockTime
		} else {
			currentLeaseRep.AbuseByType[abuseCategory] = &types.AbuseStats{
				InProgress: 1,
				LastIncident: &blockTime,
			}
		}
	// Note: we don't initialize the abuse stat by category for failed and resolved states, because initialisation should have been done on first complaint submission e.g. inprogress state
	case "failed":
		currentLesseeRep.AbuseFailed++
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok {
			currentLesseeRep.AbuseByType[abuseCategory].Failed++
		}

		currentLeaseRep.AbuseFailed++
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok {
			currentLeaseRep.AbuseByType[abuseCategory].Failed++
		}

	case "resolved":
		currentLesseeRep.AbuseResolved++
		if _, ok := currentLesseeRep.AbuseByType[abuseCategory]; ok {
			currentLesseeRep.AbuseByType[abuseCategory].Resolved++
		}

		currentLeaseRep.AbuseResolved++
		if _, ok := currentLeaseRep.AbuseByType[abuseCategory]; ok {
			currentLeaseRep.AbuseByType[abuseCategory].Resolved++
		}
	}
	currentLeaseRep.LastUpdated = &blockTime
	currentLesseeRep.LastUpdated = &blockTime
	k.SetLeaseReputation(ctx, currentLeaseRep)
	k.SetLesseeReputation(ctx, currentLesseeRep)
	return nil
}