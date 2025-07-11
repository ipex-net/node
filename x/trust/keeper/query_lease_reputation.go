package keeper

import (
	"context"

	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LeaseReputationAll(ctx context.Context, req *types.QueryAllLeaseReputationRequest) (*types.QueryAllLeaseReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var leaseReputationsDisplay []types.LeaseReputationDisplay
	var count uint64
	var skipped uint64

	// Set up pagination parameters
	limit := uint64(100) // Default limit
	offset := uint64(0)
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
	}

	// Use collections to iterate through lease reputations with pagination
	err := k.LeaseReputations.Walk(ctx, nil, func(key uint64, value types.LeaseReputation) (bool, error) {
		count++
		
		// Skip entries until we reach the offset
		if skipped < offset {
			skipped++
			return false, nil
		}
		
		// Stop if we've collected enough items
		if uint64(len(leaseReputationsDisplay)) >= limit {
			return true, nil
		}
		
		// Convert to display format
		display := k.convertLeaseReputationToDisplay(value)
		leaseReputationsDisplay = append(leaseReputationsDisplay, display)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if uint64(len(leaseReputationsDisplay)) == limit && count > offset+limit {
		// There are more items available
		nextKey = []byte("has_more")
	}

	return &types.QueryAllLeaseReputationResponse{
		LeaseReputation: leaseReputationsDisplay,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}

func (k Keeper) LeaseReputation(ctx context.Context, req *types.QueryGetLeaseReputationRequest) (*types.QueryGetLeaseReputationResponse, error) {
	if req == nil || req.Index == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLeaseReputation(ctx, req.Index)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	display := k.convertLeaseReputationToDisplay(val)
	return &types.QueryGetLeaseReputationResponse{LeaseReputation: display}, nil
}

// convertLeaseReputationToDisplay is a helper function to convert LeaseReputation to display format
func (k Keeper) convertLeaseReputationToDisplay(val types.LeaseReputation) types.LeaseReputationDisplay {
	display := types.LeaseReputationDisplay{
		Index:           val.Index,
		Level1Count:     val.Level1Count,
		Level2Count:     val.Level2Count,
		Level3Count:     val.Level3Count,
		Level4Count:     val.Level4Count,
		LastUpdated:     val.LastUpdated,
		AbuseResolved:   val.AbuseResolved,
		AbuseFailed:     val.AbuseFailed,
		AbuseInProgress: val.AbuseInProgress,
	}

	// Convert map to slice, only include initialized entries
	if val.AbuseByType != nil {
		for abuseType := abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA; abuseType <= abusetypes.AbuseType_ABUSE_TYPE_OTHER; abuseType++ {
			if stats, exists := val.AbuseByType[uint32(abuseType)]; exists && stats != nil {
				display.AbuseByType = append(display.AbuseByType, &types.AbuseStatsDisplay{
					AbuseType:    uint32(abuseType),
					Resolved:     stats.Resolved,
					Failed:       stats.Failed,
					InProgress:   stats.InProgress,
					LastIncident: stats.LastIncident,
				})
			}
		}
	}

	return display
}