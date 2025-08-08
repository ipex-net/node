package keeper

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/collections"
	"ipex/x/trust/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistLogAll(ctx context.Context, req *types.QueryAllBlacklistLogRequest) (*types.QueryAllBlacklistLogResponse, error) {
	if req == nil {
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

	// Create range for reverse iteration (latest first)
	rng := new(collections.Range[uint64]).Descending()
	err := k.BlacklistLogs.Walk(ctx, rng, func(key uint64, value types.BlacklistLog) (bool, error) {
		totalCount++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key (reverse order)
			if startKey != nil && key <= *startKey {
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

			blacklistLogs = append(blacklistLogs, value)
			lastKey = key
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

	return &types.QueryAllBlacklistLogResponse{
		BlacklistLog: blacklistLogs,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}

func (k Keeper) BlacklistLog(ctx context.Context, req *types.QueryGetBlacklistLogRequest) (*types.QueryGetBlacklistLogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	blacklistLog, found := k.GetBlacklistLog(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetBlacklistLogResponse{BlacklistLog: blacklistLog}, nil
}
