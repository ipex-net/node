package keeper

import (
	"context"

	"ipex/x/trust/types"
	abusetypes "ipex/x/abuse/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LesseeReputationAll(ctx context.Context, req *types.QueryAllLesseeReputationRequest) (*types.QueryAllLesseeReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var lesseeReputationsDisplay []types.LesseeReputationDisplay

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	lesseeReputationStore := prefix.NewStore(store, types.KeyPrefix(types.LesseeReputationKeyPrefix))

	pageRes, err := query.Paginate(lesseeReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var lesseeReputation types.LesseeReputation
		if err := k.cdc.Unmarshal(value, &lesseeReputation); err != nil {
			return err
		}

		// Convert to display format
		display := types.LesseeReputationDisplay{
			Index:           lesseeReputation.Index,
			Level1Count:     lesseeReputation.Level1Count,
			Level2Count:     lesseeReputation.Level2Count,
			Level3Count:     lesseeReputation.Level3Count,
			Level4Count:     lesseeReputation.Level4Count,
			LastUpdated:     lesseeReputation.LastUpdated,
			AbuseResolved:   lesseeReputation.AbuseResolved,
			AbuseFailed:     lesseeReputation.AbuseFailed,
			AbuseInProgress: lesseeReputation.AbuseInProgress,
		}

		// Convert map to slice, only include initialized entries
		if lesseeReputation.AbuseByType != nil {
			for abuseType := abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA; abuseType <= abusetypes.AbuseType_ABUSE_TYPE_OTHER; abuseType++ {
				if stats, exists := lesseeReputation.AbuseByType[uint32(abuseType)]; exists && stats != nil {
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

		lesseeReputationsDisplay = append(lesseeReputationsDisplay, display)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLesseeReputationResponse{LesseeReputation: lesseeReputationsDisplay, Pagination: pageRes}, nil
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

	// Convert to display format
	display := types.LesseeReputationDisplay{
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

	return &types.QueryGetLesseeReputationResponse{LesseeReputation: display}, nil
}
