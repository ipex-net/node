package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/subnet/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// GetSubnetCount get the total number of subnet
func (k Keeper) GetSubnetCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.SubnetCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetSubnetCount set the total number of subnet
func (k Keeper) SetSubnetCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.SubnetCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendSubnet appends a subnet in the store with a new id and update the count
func (k Keeper) AppendSubnet(
	ctx context.Context,
	subnet types.Subnet,
) uint64 {
	// Create the subnet
	count := k.GetSubnetCount(ctx)

	// Set the ID of the appended value
	subnet.Id = count

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	appendedValue := k.cdc.MustMarshal(&subnet)
	store.Set(GetSubnetIDBytes(subnet.Id), appendedValue)

	// Update subnet count
	k.SetSubnetCount(ctx, count+1)

	return count
}

// SetSubnet set a specific subnet in the store
func (k Keeper) SetSubnet(ctx context.Context, subnet types.Subnet) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	b := k.cdc.MustMarshal(&subnet)
	store.Set(GetSubnetIDBytes(subnet.Id), b)
}

// GetSubnet returns a subnet from its id
func (k Keeper) GetSubnet(ctx context.Context, id uint64) (val types.Subnet, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	b := store.Get(GetSubnetIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSubnet removes a subnet from the store
func (k Keeper) RemoveSubnet(ctx context.Context, id uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	store.Delete(GetSubnetIDBytes(id))
}

// GetAllSubnet returns all subnet
func (k Keeper) GetAllSubnet(ctx context.Context) (list []types.Subnet) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Subnet
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSubnetIDBytes returns the byte representation of the ID
func GetSubnetIDBytes(id uint64) []byte {
	bz := types.KeyPrefix(types.SubnetKey)
	bz = append(bz, []byte("/")...)
	bz = binary.BigEndian.AppendUint64(bz, id)
	return bz
}
