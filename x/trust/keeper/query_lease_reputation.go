package keeper

import (
	"context"

	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LeaseReputationAll(ctx context.Context, req *types.QueryAllLeaseReputationRequest) (*types.QueryAllLeaseReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var leaseReputationsDisplay []types.LeaseReputationDisplay

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	leaseReputationStore := prefix.NewStore(store, types.KeyPrefix(types.LeaseReputationKeyPrefix))

	pageRes, err := query.Paginate(leaseReputationStore, req.Pagination, func(key []byte, value []byte) error {
		var leaseReputation types.LeaseReputation
		if err := k.cdc.Unmarshal(value, &leaseReputation); err != nil {
			return err
		}

		// Convert to display format
		display := types.LeaseReputationDisplay{
			Index:           leaseReputation.Index,
			Level1Count:     leaseReputation.Level1Count,
			Level2Count:     leaseReputation.Level2Count,
			Level3Count:     leaseReputation.Level3Count,
			Level4Count:     leaseReputation.Level4Count,
			LastUpdated:     leaseReputation.LastUpdated,
			AbuseResolved:   leaseReputation.AbuseResolved,
			AbuseFailed:     leaseReputation.AbuseFailed,
			AbuseInProgress: leaseReputation.AbuseInProgress,
		}

		// Convert map to slice, only include initialized entries
		if leaseReputation.AbuseByType != nil {
			for abuseType := abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA; abuseType <= abusetypes.AbuseType_ABUSE_TYPE_OTHER; abuseType++ {
				if stats, exists := leaseReputation.AbuseByType[uint32(abuseType)]; exists && stats != nil {
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

		leaseReputationsDisplay = append(leaseReputationsDisplay, display)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLeaseReputationResponse{LeaseReputation: leaseReputationsDisplay, Pagination: pageRes}, nil
}

func (k Keeper) LeaseReputation(ctx context.Context, req *types.QueryGetLeaseReputationRequest) (*types.QueryGetLeaseReputationResponse, error) {
	if req == nil || req.Index == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetLeaseReputation(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	// Convert to display format
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

	return &types.QueryGetLeaseReputationResponse{LeaseReputation: display}, nil
}
