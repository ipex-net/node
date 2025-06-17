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
func (k Keeper) UpdateSubnetReputation(ctx context.Context, subnetId uint64, leaseId uint64, lessorOrgId uint64, lesseeOrgId uint64, reputation types.SubnetBlacklistCount) error {
	// Unwrap SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Get block time
	blockTime := sdkCtx.BlockTime()
	if blockTime.IsZero() {
		return fmt.Errorf("invalid block time: cannot be zero")
	}
	// Get the current subnet reputation
	currentRep, found := k.GetSubnetReputation(ctx, subnetId)
	// Calculating the reputation change
	var level1Change, level2Change, level3Change, level4Change int32
	if found {
		level1Change = int32(reputation.Level1Count) - int32(currentRep.Level1Count)
		level2Change = int32(reputation.Level2Count) - int32(currentRep.Level2Count)
		level3Change = int32(reputation.Level3Count) - int32(currentRep.Level3Count)
		level4Change = int32(reputation.Level4Count) - int32(currentRep.Level4Count)
	} else {
		level1Change, level2Change, level3Change, level4Change = int32(reputation.Level1Count), int32(reputation.Level2Count), int32(reputation.Level3Count), int32(reputation.Level4Count)
	}

	// Logging only if there was a reputation change
	if level1Change != 0 || level2Change != 0 || level3Change != 0 || level4Change != 0 {
		logEntry := types.BlacklistLog{
			SubnetId:     subnetId,
			LessorOrgId:  lessorOrgId,
			LeaseId:      leaseId,
			LesseeOrgId:  lesseeOrgId,
			Level1Change: level1Change,
			Level2Change: level2Change,
			Level3Change: level3Change,
			Level4Change: level4Change,
			Timestamp:    &blockTime,
		}
		k.AppendBlacklistLog(ctx, logEntry)
	}

	subnetRep := types.SubnetReputation{
		Index:       subnetId,
		Level1Count: reputation.Level1Count,
		Level2Count: reputation.Level2Count,
		Level3Count: reputation.Level3Count,
		Level4Count: reputation.Level4Count,
		LastUpdated: &blockTime,
	}
	// TODO: log the subnet reputation change
	// TODO: update lessee reputation
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
			if now.Sub(lastAttempt) < 6*time.Hour {
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

// InitializeSubnetReputation initializes a new subnet reputation entry when a subnet is created
func (k Keeper) InitializeSubnetReputation(ctx context.Context, subnetID uint64) {

	// Create new subnet reputation with initial values
	subnetReputation := types.SubnetReputation{
		Index:       subnetID,
		Level1Count: 0, // Initialize all counts to zero
		Level2Count: 0,
		Level3Count: 0,
		Level4Count: 0,
		LastUpdated: nil,
	}

	// Set the subnet reputation in the store
	k.SetSubnetReputation(ctx, subnetReputation)

}
