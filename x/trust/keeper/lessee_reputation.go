package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetLesseeReputation set a specific lesseeReputation in the store from its index
func (k Keeper) SetLesseeReputation(ctx context.Context, lesseeReputation types.LesseeReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LesseeReputationKeyPrefix))
	b := k.cdc.MustMarshal(&lesseeReputation)
	store.Set(types.LesseeReputationKey(
		lesseeReputation.Index,
	), b)
}

// GetLesseeReputation returns a lesseeReputation from its index
func (k Keeper) GetLesseeReputation(
	ctx context.Context,
	index uint64,

) (val types.LesseeReputation, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LesseeReputationKeyPrefix))

	b := store.Get(types.LesseeReputationKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLesseeReputation removes a lesseeReputation from the store
func (k Keeper) RemoveLesseeReputation(
	ctx context.Context,
	index uint64,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LesseeReputationKeyPrefix))
	store.Delete(types.LesseeReputationKey(
		index,
	))
}

// GetAllLesseeReputation returns all lesseeReputation
func (k Keeper) GetAllLesseeReputation(ctx context.Context) (list []types.LesseeReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LesseeReputationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.LesseeReputation
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
