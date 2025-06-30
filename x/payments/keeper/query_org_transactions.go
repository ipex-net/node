package keeper

import (
	"context"

	"ipex/x/payments/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// GetTransactionsByOrg returns all transactions where the org is involved (either as from or to)
func (k Keeper) GetTransactionsByOrg(ctx context.Context, orgId uint64) ([]types.Transaction, error) {
	var transactions []types.Transaction

	// Use org index to efficiently get transactions for this organization
	rng := collections.NewPrefixedPairRange[uint64, uint64](orgId)
	err := k.TransactionsByOrg.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// Get the transaction ID from the key
		transactionId := key.K2()
		
		// Fetch the full transaction
		transaction, found := k.GetTransaction(ctx, transactionId)
		if !found {
			// Transaction might have been deleted, skip
			return false, nil
		}
		
		transactions = append(transactions, transaction)
		return false, nil
	})

	return transactions, err
}

// GetTransactionsByOrgPaginated returns paginated transactions where the org is involved
func (k Keeper) GetTransactionsByOrgPaginated(ctx context.Context, orgId uint64, pagination *query.PageRequest) ([]types.Transaction, *query.PageResponse, error) {
	var transactions []types.Transaction
	var count uint64
	var skipped uint64

	// Set up pagination parameters
	limit := uint64(100) // Default limit
	offset := uint64(0)
	if pagination != nil {
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
		if pagination.Offset > 0 {
			offset = pagination.Offset
		}
	}

	// Use org index to efficiently get transactions for this organization
	rng := collections.NewPrefixedPairRange[uint64, uint64](orgId)
	err := k.TransactionsByOrg.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		count++
		
		// Skip entries until we reach the offset
		if skipped < offset {
			skipped++
			return false, nil
		}
		
		// Stop if we've collected enough items
		if uint64(len(transactions)) >= limit {
			return true, nil
		}
		
		// Get the transaction ID from the key
		transactionId := key.K2()
		
		// Fetch the full transaction
		transaction, found := k.GetTransaction(ctx, transactionId)
		if !found {
			// Transaction might have been deleted, skip
			return false, nil
		}
		
		transactions = append(transactions, transaction)
		return false, nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Calculate next key for pagination
	var nextKey []byte
	if uint64(len(transactions)) == limit && count > offset+limit {
		// There are more items available
		nextKey = []byte("has_more")
	}

	pageResponse := &query.PageResponse{
		NextKey: nextKey,
		Total:   count,
	}

	return transactions, pageResponse, nil
}