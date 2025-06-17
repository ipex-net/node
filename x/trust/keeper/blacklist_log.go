package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// GetBlacklistLogCount get the total number of blacklistLog
func (k Keeper) GetBlacklistLogCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.BlacklistLogCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetBlacklistLogCount set the total number of blacklistLog
func (k Keeper) SetBlacklistLogCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.BlacklistLogCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendBlacklistLog appends a blacklistLog in the store with a new id and update the count
func (k Keeper) AppendBlacklistLog(
	ctx context.Context,
	blacklistLog types.BlacklistLog,
) uint64 {
	// Create the blacklistLog
	count := k.GetBlacklistLogCount(ctx)

	// Set the ID of the appended value
	blacklistLog.Id = count

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BlacklistLogKey))
	appendedValue := k.cdc.MustMarshal(&blacklistLog)
	store.Set(GetBlacklistLogIDBytes(blacklistLog.Id), appendedValue)

	// Update blacklistLog count
	k.SetBlacklistLogCount(ctx, count+1)

	return count
}

// SetBlacklistLog set a specific blacklistLog in the store
func (k Keeper) SetBlacklistLog(ctx context.Context, blacklistLog types.BlacklistLog) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BlacklistLogKey))
	b := k.cdc.MustMarshal(&blacklistLog)
	store.Set(GetBlacklistLogIDBytes(blacklistLog.Id), b)
}

// GetBlacklistLog returns a blacklistLog from its id
func (k Keeper) GetBlacklistLog(ctx context.Context, id uint64) (val types.BlacklistLog, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BlacklistLogKey))
	b := store.Get(GetBlacklistLogIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlacklistLog removes a blacklistLog from the store
func (k Keeper) RemoveBlacklistLog(ctx context.Context, id uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BlacklistLogKey))
	store.Delete(GetBlacklistLogIDBytes(id))
}

// GetAllBlacklistLog returns all blacklistLog
func (k Keeper) GetAllBlacklistLog(ctx context.Context) (list []types.BlacklistLog) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BlacklistLogKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.BlacklistLog
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetBlacklistLogIDBytes returns the byte representation of the ID
func GetBlacklistLogIDBytes(id uint64) []byte {
	bz := types.KeyPrefix(types.BlacklistLogKey)
	bz = append(bz, []byte("/")...)
	bz = binary.BigEndian.AppendUint64(bz, id)
	return bz
}
