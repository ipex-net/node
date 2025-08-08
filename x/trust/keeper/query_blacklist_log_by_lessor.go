package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/trust/types"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogByLessor(ctx context.Context, req *types.QueryBlacklistLogByLessorRequest) (*types.QueryBlacklistLogByLessorResponse, error) {
	if req == nil || req.LessorOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklistLogs []types.BlacklistLog
	var totalCount uint64
	var lastKey uint64

	// Set up pagination parameters
	limit := uint64(20) // Default limit
	offset := uint64(0)
	var startKey *uint64

	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		// Support key-based pagination
		if len(req.Pagination.Key) > 0 {
			// Parse the key as uint64 (log ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	var skipped uint64
	var collectedCount uint64
	var collectingItems = true

	// Iterate over all logs with this lessor ID using prefix iteration (reverse for latest first)
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.LessorOrgId).Descending()
	err := k.BlacklistLogsByLessor.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// key.K2() is the log ID
		logId := key.K2()
		log, found := k.GetBlacklistLog(ctx, logId)
		if !found {
			return false, nil
		}

		totalCount++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key (reverse order)
			if startKey != nil && logId >= *startKey {
				return false, nil
			}

			// Skip entries until we reach the offset (for offset-based pagination)
			if startKey == nil && skipped < offset {
				skipped++
				return false, nil
			}

			// Stop collecting items if we've collected enough
			if collectedCount >= limit {
				collectingItems = false
				return false, nil // Continue counting but stop collecting
			}

			blacklistLogs = append(blacklistLogs, log)
			lastKey = logId
			collectedCount++
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if collectedCount == limit {
		// Convert the last log ID to bytes for next key
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, lastKey)
	}

	return &types.QueryBlacklistLogByLessorResponse{
		BlacklistLog: blacklistLogs,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
