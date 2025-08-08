package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/payments/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) TransactionAll(ctx context.Context, req *types.QueryAllTransactionRequest) (*types.QueryAllTransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var transactions []types.Transaction
	var count uint64
	var skipped uint64
	var collectingItems = true

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
			// Parse the key as uint64 (transaction ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	var collectedCount uint64
	var lastKey uint64

	// Use collections to iterate through transactions with pagination
	err := k.Transactions.Walk(ctx, nil, func(key uint64, value types.Transaction) (bool, error) {
		// Count total transactions (always count for accurate total)
		count++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key
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

			transactions = append(transactions, value)
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
		// Convert the last transaction ID to bytes for next key
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, lastKey)
	}

	return &types.QueryAllTransactionResponse{
		Transaction: transactions,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}

func (k Keeper) Transaction(ctx context.Context, req *types.QueryGetTransactionRequest) (*types.QueryGetTransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	transaction, found := k.GetTransaction(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetTransactionResponse{Transaction: transaction}, nil
}
