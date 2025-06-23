package keeper

import (
	"context"

	"ipex/x/abuse/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetComplaint set a specific complaint in the store from its index
func (k Keeper) SetComplaint(ctx context.Context, complaint types.Complaint) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	b := k.cdc.MustMarshal(&complaint)
	store.Set(types.ComplaintKey(
		complaint.Index,
	), b)
}

// GetComplaint returns a complaint from its index
func (k Keeper) GetComplaint(
	ctx context.Context,
	index string,

) (val types.Complaint, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))

	b := store.Get(types.ComplaintKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveComplaint removes a complaint from the store
func (k Keeper) RemoveComplaint(
	ctx context.Context,
	index string,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ComplaintKeyPrefix))
	store.Delete(types.ComplaintKey(
		index,
	))
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
