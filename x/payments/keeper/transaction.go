package keeper

import (
	"context"

	"ipex/x/payments/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


// AppendTransaction appends a transaction in the store with a new id
func (k Keeper) AppendTransaction(
	ctx context.Context,
	transaction types.Transaction,
) uint64 {
	// Get next ID from sequence
	id, err := k.TransactionSequence.Next(ctx)
	if err != nil {
		panic(err)
	}
	transaction.Id = id

	// Store the transaction
	err = k.Transactions.Set(ctx, id, transaction)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateTransactionIndexes(ctx, transaction)

	return id
}

// SetTransaction set a specific transaction in the store
func (k Keeper) SetTransaction(ctx context.Context, transaction types.Transaction) {
	// Store the transaction
	err := k.Transactions.Set(ctx, transaction.Id, transaction)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes (in case any indexed fields changed)
	k.updateTransactionIndexes(ctx, transaction)
}

// GetTransaction returns a transaction from its id
func (k Keeper) GetTransaction(ctx context.Context, id uint64) (val types.Transaction, found bool) {
	transaction, err := k.Transactions.Get(ctx, id)
	if err != nil {
		if err.Error() == "collections: not found" {
			return val, false
		}
		panic(err)
	}
	return transaction, true
}

// RemoveTransaction removes a transaction from the store
func (k Keeper) RemoveTransaction(ctx context.Context, id uint64) {
	// Get the transaction first to access its indexed fields
	transaction, found := k.GetTransaction(ctx, id)
	if !found {
		return
	}

	// Remove from main collection
	err := k.Transactions.Remove(ctx, id)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeTransactionFromIndexes(ctx, transaction)
}

// GetAllTransaction returns all transaction
func (k Keeper) GetAllTransaction(ctx context.Context) (list []types.Transaction) {
	err := k.Transactions.Walk(ctx, nil, func(key uint64, value types.Transaction) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// updateTransactionIndexes updates all secondary indexes for a transaction
func (k Keeper) updateTransactionIndexes(ctx context.Context, transaction types.Transaction) {
	// Index by entity
	if transaction.EntityId != 0 {
		err := k.TransactionsByEntity.Set(ctx, collections.Join(transaction.EntityId, transaction.Id), true)
		if err != nil {
			panic(err)
		}
	}

	// Index by org - add entry for FromOrg
	if transaction.FromOrgId != 0 {
		err := k.TransactionsByOrg.Set(ctx, collections.Join(transaction.FromOrgId, transaction.Id), true)
		if err != nil {
			panic(err)
		}
	}

	// Index by org - add entry for ToOrg (if different from FromOrg)
	if transaction.ToOrgId != 0 && transaction.ToOrgId != transaction.FromOrgId {
		err := k.TransactionsByOrg.Set(ctx, collections.Join(transaction.ToOrgId, transaction.Id), true)
		if err != nil {
			panic(err)
		}
	}
}

// removeTransactionFromIndexes removes transaction from all secondary indexes
func (k Keeper) removeTransactionFromIndexes(ctx context.Context, transaction types.Transaction) {
	// Remove from entity index
	if transaction.EntityId != 0 {
		k.TransactionsByEntity.Remove(ctx, collections.Join(transaction.EntityId, transaction.Id))
	}

	// Remove from org index - FromOrg entry
	if transaction.FromOrgId != 0 {
		k.TransactionsByOrg.Remove(ctx, collections.Join(transaction.FromOrgId, transaction.Id))
	}

	// Remove from org index - ToOrg entry (if different)
	if transaction.ToOrgId != 0 && transaction.ToOrgId != transaction.FromOrgId {
		k.TransactionsByOrg.Remove(ctx, collections.Join(transaction.ToOrgId, transaction.Id))
	}
}


func (k Keeper) RecordTransaction(
	goCtx context.Context,
	transactionType types.TransactionType,
	relatedModuleId types.ModuleID,
	entityId uint64,
	amount sdk.Coin,
	fromOrgId uint64,
	toOrgId uint64,
	fromAddress string,
	toAddress string,
) uint64 {
	// Unwrap the SDK context for methods that require it
	ctx := sdk.UnwrapSDKContext(goCtx)
	blockTime := ctx.BlockTime()
	// Create the transaction
	transaction := types.Transaction{
		TransactionType: transactionType,
		RelatedModuleId: relatedModuleId,
		EntityId:        entityId,
		Amount:          amount,
		FromOrgId:       fromOrgId,
		ToOrgId:         toOrgId,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		Timestamp:       &blockTime,
		BlockHeight:     ctx.BlockHeight(),
	}

	//Append the transaction
	id := k.AppendTransaction(ctx, transaction)

	return id
}
