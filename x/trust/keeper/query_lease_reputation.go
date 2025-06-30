package keeper

import (
	"context"

	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LeaseReputationAll(ctx context.Context, req *types.QueryAllLeaseReputationRequest) (*types.QueryAllLeaseReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Use collections-based GetAll instead of legacy KV store operations
	leaseReputations := k.GetAllLeaseReputation(ctx)
	var leaseReputationsDisplay []types.LeaseReputationDisplay

	for _, leaseReputation := range leaseReputations {
		// Convert to display format
		display := k.convertLeaseReputationToDisplay(leaseReputation)
		leaseReputationsDisplay = append(leaseReputationsDisplay, display)
	}

	return &types.QueryAllLeaseReputationResponse{LeaseReputation: leaseReputationsDisplay, Pagination: nil}, nil
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
