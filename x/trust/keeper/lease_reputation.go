package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetLeaseReputation set a specific leaseReputation in the store from its index
func (k Keeper) SetLeaseReputation(ctx context.Context, leaseReputation types.LeaseReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseReputationKeyPrefix))
	b := k.cdc.MustMarshal(&leaseReputation)
	store.Set(types.LeaseReputationKey(
		leaseReputation.Index,
	), b)
}

// GetLeaseReputation returns a leaseReputation from its index
func (k Keeper) GetLeaseReputation(
	ctx context.Context,
	index uint64,

) (val types.LeaseReputation, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseReputationKeyPrefix))

	b := store.Get(types.LeaseReputationKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveLeaseReputation removes a leaseReputation from the store
func (k Keeper) RemoveLeaseReputation(
	ctx context.Context,
	index uint64,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseReputationKeyPrefix))
	store.Delete(types.LeaseReputationKey(
		index,
	))
}

// GetAllLeaseReputation returns all leaseReputation
func (k Keeper) GetAllLeaseReputation(ctx context.Context) (list []types.LeaseReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.LeaseReputationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.LeaseReputation
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// InitializeLeaseReputation initializes a new lease reputation entry when the lease is created
func (k Keeper) InitializeLeaseReputation(ctx context.Context, leaseId uint64) {
	// Create new lease reputation with initial values
	// Even if subnet has some blacklist entries, we initialize all lease reputation blacklist counts to zero
	// This reputation only tracks changes related to lease, not the absolute blacklist counts per subnet
	leaseReputation := types.LeaseReputation{
		Index:       leaseId,
		Level1Count: 0,
		Level2Count: 0,
		Level3Count: 0,
		Level4Count: 0,
		LastUpdated: nil,
	}

	// Set the subnet reputation in the store
	k.SetLeaseReputation(ctx, leaseReputation)

}
