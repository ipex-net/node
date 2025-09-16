package keeper

import (
	"context"
	"encoding/binary"
	"ipex/x/payments/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListPaymentFiltered(goCtx context.Context, req *types.QueryListPaymentFilteredRequest) (*types.QueryListPaymentFilteredResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	
	// Organization ID is required
	if req.OrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	var transactions []types.Transaction
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
			// Parse the key as uint64 (transaction ID)
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

	// Use org index to efficiently get transactions for this organization
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.OrgId)
	err := k.TransactionsByOrg.Walk(goCtx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// Get the transaction ID from the key
		transactionId := key.K2()

		// Fetch the full transaction
		transaction, found := k.GetTransaction(goCtx, transactionId)
		if !found {
			// Transaction might have been deleted, skip
			return false, nil
		}

		// Count total matching transactions (always count for accurate total)
		totalCount++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key
			if startKey != nil && transactionId <= *startKey {
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

			// Apply filters - skip transaction if it doesn't match criteria
			
			// Filter by transaction type (0 means no filter)
			if req.TransactionType != 0 && uint64(transaction.TransactionType) != req.TransactionType {
				return false, nil
			}
			
			// Filter by module ID (0 means no filter)  
			if req.ModuleId != 0 && uint64(transaction.RelatedModuleId) != req.ModuleId {
				return false, nil
			}
			
			// Filter by timestamp range (0 means no filter)
			if transaction.Timestamp != nil {
				transactionTimestamp := uint64(transaction.Timestamp.Unix())
				if req.FromTimestamp != 0 && transactionTimestamp < req.FromTimestamp {
					return false, nil
				}
				if req.ToTimestamp != 0 && transactionTimestamp > req.ToTimestamp {
					return false, nil
				}
			}
			
			// Filter by amount range (0 means no filter)
			transactionAmount := transaction.Amount.Amount
			if req.MinAmount != 0 {
				minAmount := math.NewIntFromUint64(req.MinAmount)
				if transactionAmount.LT(minAmount) {
					return false, nil
				}
			}
			if req.MaxAmount != 0 {
				maxAmount := math.NewIntFromUint64(req.MaxAmount)
				if transactionAmount.GT(maxAmount) {
					return false, nil
				}
			}

			transactions = append(transactions, transaction)
			lastKey = transactionId
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

	return &types.QueryListPaymentFilteredResponse{
		Transaction: transactions,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
