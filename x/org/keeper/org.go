package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/org/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetOrgCount get the total number of org
func (k Keeper) GetOrgCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.OrgCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetOrgCount set the total number of org
func (k Keeper) SetOrgCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.OrgCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendOrg appends a org in the store with a new id and update the count
func (k Keeper) AppendOrg(
	ctx context.Context,
	org types.Org,
) uint64 {
	// Create the org
	count := k.GetOrgCount(ctx)

	// Set the ID of the appended value
	org.Id = count

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	appendedValue := k.cdc.MustMarshal(&org)
	store.Set(GetOrgIDBytes(org.Id), appendedValue)

	// Update org count
	k.SetOrgCount(ctx, count+1)

	return count
}

// SetOrg set a specific org in the store
func (k Keeper) SetOrg(ctx context.Context, org types.Org) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	b := k.cdc.MustMarshal(&org)
	store.Set(GetOrgIDBytes(org.Id), b)
}

// GetOrg returns a org from its id
func (k Keeper) GetOrg(ctx context.Context, id uint64) (val types.Org, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	b := store.Get(GetOrgIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveOrg removes a org from the store
func (k Keeper) RemoveOrg(ctx context.Context, id uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	store.Delete(GetOrgIDBytes(id))
}

// GetAllOrg returns all org
func (k Keeper) GetAllOrg(ctx context.Context) (list []types.Org) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Org
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetOrgIDBytes returns the byte representation of the ID
func GetOrgIDBytes(id uint64) []byte {
	bz := types.KeyPrefix(types.OrgKey)
	bz = append(bz, []byte("/")...)
	bz = binary.BigEndian.AppendUint64(bz, id)
	return bz
}

func (k Keeper) HasOrg(ctx context.Context, id uint64) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	return store.Has(GetOrgIDBytes(id))
}

func (k Keeper) VerifyOrgAdmin(ctx context.Context, orgid uint64, admin string) bool {

	// Validate the provided admin address
	_, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return false
	}
    // Get the org
    org, found := k.GetOrg(ctx, orgid)
    if !found {
        return false
    }

    return org.Admin == admin
}