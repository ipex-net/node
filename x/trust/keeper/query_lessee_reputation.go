package keeper

import (
	"context"

	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Helper function to convert LesseeReputation to LesseeReputationDisplay
func (k Keeper) convertLesseeReputationToDisplay(value types.LesseeReputation) types.LesseeReputationDisplay {
	display := types.LesseeReputationDisplay{
		Index:                       value.Index,
		Level1Count:                 value.Level1Count,
		Level2Count:                 value.Level2Count,
		Level3Count:                 value.Level3Count,
		Level4Count:                 value.Level4Count,
		LastUpdated:                 value.LastUpdated,
		AbuseResolved:               value.AbuseResolved,
		AbuseFailed:                 value.AbuseFailed,
		AbuseInProgress:             value.AbuseInProgress,
		IpMonthsTotal:               value.IpMonthsTotal,
		IpsRentedNow:                value.IpsRentedNow,
		ContractsActive:             value.ContractsActive,
		ContractsTerminated:         value.ContractsTerminated,
		ContractsTerminatedViolated: value.ContractsTerminatedViolated,
	}

	// Convert map to slice, only include initialized entries
	if value.AbuseByType != nil {
		for abuseType := abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA; abuseType <= abusetypes.AbuseType_ABUSE_TYPE_OTHER; abuseType++ {
			if stats, exists := value.AbuseByType[uint32(abuseType)]; exists && stats != nil {
				display.AbuseByType = append(display.AbuseByType, &types.AbuseTypeStatsDisplay{
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

func (k Keeper) LesseeReputationAll(ctx context.Context, req *types.QueryAllLesseeReputationRequest) (*types.QueryAllLesseeReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var lesseeReputationsDisplay []types.LesseeReputationDisplay
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

	// Use collections to iterate through lessee reputations with pagination
	err := k.LesseeReputations.Walk(ctx, nil, func(key uint64, value types.LesseeReputation) (bool, error) {
		count++

		// Skip entries until we reach the offset
		if skipped < offset {
			skipped++
			return false, nil
		}

		// Stop if we've collected enough items
		if uint64(len(lesseeReputationsDisplay)) >= limit {
			return true, nil
		}

		// Use helper function to convert to display format
		display := k.convertLesseeReputationToDisplay(value)
		lesseeReputationsDisplay = append(lesseeReputationsDisplay, display)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if uint64(len(lesseeReputationsDisplay)) == limit && count > offset+limit {
		// There are more items available
		nextKey = []byte("has_more")
	}

	return &types.QueryAllLesseeReputationResponse{
		LesseeReputation: lesseeReputationsDisplay,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}

func (k Keeper) LesseeReputation(ctx context.Context, req *types.QueryGetLesseeReputationRequest) (*types.QueryGetLesseeReputationResponse, error) {
	if req == nil || req.Index == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLesseeReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	// Use helper function to convert to display format
	display := k.convertLesseeReputationToDisplay(val)

	return &types.QueryGetLesseeReputationResponse{LesseeReputation: display}, nil
}
