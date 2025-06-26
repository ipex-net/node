package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/abuse/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// AppendComplaint appends a complaint to the store with an auto-incrementing id
func (k Keeper) AppendComplaint(ctx context.Context, complaint types.Complaint) uint64 {
	count := k.GetComplaintCount(ctx)
	// Set the ID of the appended value (start from 1, not 0)
	complaint.Id = count + 1
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	appendedValue := k.cdc.MustMarshal(&complaint)
	store.Set(types.GetComplaintIDBytes(complaint.Id), appendedValue)
	k.SetComplaintCount(ctx, count+1)
	return complaint.Id
}

// GetComplaintCount returns the total number of complaints
func (k Keeper) GetComplaintCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ComplaintCountKey)
	bz := store.Get(byteKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// SetComplaintCount sets the total number of complaints
func (k Keeper) SetComplaintCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ComplaintCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// SetComplaint set a specific complaint in the store from its id
func (k Keeper) SetComplaint(ctx context.Context, complaint types.Complaint) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	b := k.cdc.MustMarshal(&complaint)
	store.Set(types.GetComplaintIDBytes(complaint.Id), b)
}

// GetComplaint returns a complaint from its id
func (k Keeper) GetComplaint(
	ctx context.Context,
	id uint64,
) (val types.Complaint, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))

	b := store.Get(types.GetComplaintIDBytes(id))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveComplaint removes a complaint from the store
func (k Keeper) RemoveComplaint(
	ctx context.Context,
	id uint64,
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	store.Delete(types.GetComplaintIDBytes(id))
}

// GetAllComplaint returns all complaint
func (k Keeper) GetAllComplaint(ctx context.Context) (list []types.Complaint) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Complaint
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}


