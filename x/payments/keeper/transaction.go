package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	"ipex/x/payments/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTransactionCount get the total number of transaction
func (k Keeper) GetTransactionCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.TransactionCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetTransactionCount set the total number of transaction
func (k Keeper) SetTransactionCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.TransactionCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendTransaction appends a transaction in the store with a new id and update the count
func (k Keeper) AppendTransaction(
	ctx context.Context,
	transaction types.Transaction,
) uint64 {
	// Create the transaction
	count := k.GetTransactionCount(ctx)

	// Set the ID of the appended value
	transaction.Id = count

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TransactionKey))
	appendedValue := k.cdc.MustMarshal(&transaction)
	store.Set(GetTransactionIDBytes(transaction.Id), appendedValue)

	// Update transaction count
	k.SetTransactionCount(ctx, count+1)

	return count
}

// SetTransaction set a specific transaction in the store
func (k Keeper) SetTransaction(ctx context.Context, transaction types.Transaction) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TransactionKey))
	b := k.cdc.MustMarshal(&transaction)
	store.Set(GetTransactionIDBytes(transaction.Id), b)
}

// GetTransaction returns a transaction from its id
func (k Keeper) GetTransaction(ctx context.Context, id uint64) (val types.Transaction, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TransactionKey))
	b := store.Get(GetTransactionIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveTransaction removes a transaction from the store
func (k Keeper) RemoveTransaction(ctx context.Context, id uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TransactionKey))
	store.Delete(GetTransactionIDBytes(id))
}

// GetAllTransaction returns all transaction
func (k Keeper) GetAllTransaction(ctx context.Context) (list []types.Transaction) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.TransactionKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Transaction
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetTransactionIDBytes returns the byte representation of the ID
func GetTransactionIDBytes(id uint64) []byte {
	bz := types.KeyPrefix(types.TransactionKey)
	bz = append(bz, []byte("/")...)
	bz = binary.BigEndian.AppendUint64(bz, id)
	return bz
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
    referenceId uint64,
    memo string,
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
        ReferenceId:     referenceId,
        Memo:            memo,
        Timestamp:       &blockTime,
		BlockHeight: 	 ctx.BlockHeight(),
    }
    
	//Append the transaction
    id :=k.AppendTransaction(ctx, transaction)

    // Emit event with all relevant transaction details
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeTransactionRecorded,
            sdk.NewAttribute(types.AttributeKeyTransactionID, fmt.Sprintf("%d", id)),
            sdk.NewAttribute(types.AttributeKeyTransactionType, transaction.TransactionType.String()),
            sdk.NewAttribute(types.AttributeKeyModuleID, transaction.RelatedModuleId.String()),
            sdk.NewAttribute(types.AttributeKeyEntityID, fmt.Sprintf("%d", transaction.EntityId)),
            sdk.NewAttribute(types.AttributeKeyAmount, transaction.Amount.String()),
            sdk.NewAttribute(types.AttributeKeyFromOrgID, fmt.Sprintf("%d", transaction.FromOrgId)),
            sdk.NewAttribute(types.AttributeKeyToOrgID, fmt.Sprintf("%d", transaction.ToOrgId)),
            sdk.NewAttribute(types.AttributeKeyFromAddress, transaction.FromAddress),
            sdk.NewAttribute(types.AttributeKeyToAddress, transaction.ToAddress),
            sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", transaction.BlockHeight)),
            sdk.NewAttribute(types.AttributeKeyReferenceID, fmt.Sprintf("%d", transaction.ReferenceId)),
        ),
    )
    return id
}