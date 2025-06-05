package keeper

import (
	"context"
	"fmt"
	"time"

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
func (k Keeper) UpdateSubnetReputation(ctx context.Context, subnetID uint64, reputation types.SubnetBlacklistCount) error {
	// Validate input parameters
	// if subnetID == 0 {
	// 	return fmt.Errorf("invalid subnet ID: cannot be zero")
	// }
	// Unwrap SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Get block time
	blockTime := sdkCtx.BlockTime()
	if blockTime.IsZero() {
		return fmt.Errorf("invalid block time: cannot be zero")
	}
    subnetRep := types.SubnetReputation{
        Index:       subnetID,
        Level1Count: reputation.Level1Count,
        Level2Count: reputation.Level2Count,
        Level3Count: reputation.Level3Count,
        Level4Count: reputation.Level4Count,
        LastUpdated: &blockTime,
    }
    
    k.SetSubnetReputation(ctx, subnetRep)
	return nil
}


func (k Keeper) GetSubnetsForReputationScanning(ctx context.Context) (list []uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetReputationKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()
	count := 0
	maxSubnets := 10
	for ; iterator.Valid(); iterator.Next() {
		var subnetReputation types.SubnetReputation
		k.cdc.MustUnmarshal(iterator.Value(), &subnetReputation)

		// Check if enough time has passed since last attempt
		if subnetReputation.LastUpdated != nil {
			lastAttempt := *subnetReputation.LastUpdated
			// Skip if not enough time has passed
			if now.Sub(lastAttempt) < 6 * time.Hour {
				continue
			}
		}

		// Add to the list for validation
		list = append(list, subnetReputation.Index)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			break
		}
	}

	return
}
