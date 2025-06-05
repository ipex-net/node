package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetSubnetReputation set a specific subnetReputation in the store from its index
func (k Keeper) SetSubnetReputation(ctx context.Context, subnetReputation types.SubnetReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetReputationKeyPrefix))
	b := k.cdc.MustMarshal(&subnetReputation)
	store.Set(types.SubnetReputationKey(
		subnetReputation.Index,
	), b)
}

// GetSubnetReputation returns a subnetReputation from its index
func (k Keeper) GetSubnetReputation(
	ctx context.Context,
	index uint64,

) (val types.SubnetReputation, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetReputationKeyPrefix))

	b := store.Get(types.SubnetReputationKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSubnetReputation removes a subnetReputation from the store
func (k Keeper) RemoveSubnetReputation(
	ctx context.Context,
	index uint64,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetReputationKeyPrefix))
	store.Delete(types.SubnetReputationKey(
		index,
	))
}

// GetAllSubnetReputation returns all subnetReputation
func (k Keeper) GetAllSubnetReputation(ctx context.Context) (list []types.SubnetReputation) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetReputationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.SubnetReputation
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}


// UpdateSubnetReputation is a helper method 
func (k Keeper) UpdateSubnetReputation(ctx context.Context, subnetID uint64, l1, l2, l3, l4 uint64) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
    subnetRep := types.SubnetReputation{
        Index:       subnetID,
        Level1Count: l1,
        Level2Count: l2,
        Level3Count: l3,
        Level4Count: l4,
        LastUpdated: &blockTime,
    }
    
    k.SetSubnetReputation(ctx, subnetRep)
}