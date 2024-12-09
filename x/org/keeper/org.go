package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/org/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
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

// Permissions
// GetUserPermissionCount get the total number of permissions
func (k Keeper) GetUserPermissionCount(ctx context.Context) uint64 {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, []byte{})
    byteKey := types.KeyPrefix(types.PermissionCountKey)
    bz := store.Get(byteKey)

    // Count doesn't exist: no element
    if bz == nil {
        return 0
    }

    // Parse bytes
    return binary.BigEndian.Uint64(bz)
}

// SetUserPermissionCount set the total number of permissions
func (k Keeper) SetUserPermissionCount(ctx context.Context, count uint64) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, []byte{})
    byteKey := types.KeyPrefix(types.PermissionCountKey)
    bz := make([]byte, 8)
    binary.BigEndian.PutUint64(bz, count)
    store.Set(byteKey, bz)
}

// SetUserPermission set a specific permission in the store
func (k Keeper) SetUserPermission(ctx context.Context, permission types.UserPermission) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PermissionKey))
    key := GetPermissionKey(permission.OrgId, permission.UserAddress)
    b := k.cdc.MustMarshal(&permission)
    store.Set(key, b)
}

// GetUserPermission returns a permission from orgId and userAddress
func (k Keeper) GetUserPermission(ctx context.Context, orgID uint64, userAddress string) (val types.UserPermission, found bool) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PermissionKey))
    key := GetPermissionKey(orgID, userAddress)
    b := store.Get(key)
    if b == nil {
        return val, false
    }
    k.cdc.MustUnmarshal(b, &val)
    return val, true
}

// RemoveUserPermission removes a permission from the store
func (k Keeper) RemoveUserPermission(ctx context.Context, orgID uint64, userAddress string) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PermissionKey))
    key := GetPermissionKey(orgID, userAddress)
    store.Delete(key)
}

// GetAllUserPermissions returns all permissions
func (k Keeper) GetAllUserPermissions(ctx context.Context) (list []types.UserPermission) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PermissionKey))
    iterator := storetypes.KVStorePrefixIterator(store, []byte{})

    defer iterator.Close()

    for ; iterator.Valid(); iterator.Next() {
        var val types.UserPermission
        k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
    }

    return
}

// GetOrgUserPermissions returns all permissions for a specific organization
func (k Keeper) GetOrgUserPermissions(ctx context.Context, orgID uint64) (list []types.UserPermission) {
    storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
    store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.PermissionKey))
    orgPrefix := GetOrgPermissionPrefix(orgID)
    iterator := storetypes.KVStorePrefixIterator(store, orgPrefix)

    defer iterator.Close()

    for ; iterator.Valid(); iterator.Next() {
        var val types.UserPermission
        k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
    }

    return
}

// HasPermission checks if a user has a specific permission in an organization
func (k Keeper) HasPermission(ctx context.Context, orgID uint64, userAddress string, permission uint64) bool {
    userPerm, found := k.GetUserPermission(ctx, orgID, userAddress)
    if !found {
        return false
    }
    return (userPerm.Permissions & permission) == permission
}

// UpdateUserPermission updates or creates a permission entry
func (k Keeper) UpdateUserPermission(
    ctx context.Context,
    orgID uint64,
    userAddress string,
    newPermissions uint64,
) error {
    // Verify org exists
    _, found := k.GetOrg(ctx, orgID)
    if !found {
        return types.ErrOrgNotFound
    }

    permission := types.UserPermission{
        OrgId:       orgID,
        UserAddress: userAddress,
        Permissions: newPermissions,
    }

    k.SetUserPermission(ctx, permission)
    return nil
}


// GetPermissionKey gets the store key for a permission entry
func GetPermissionKey(orgID uint64, userAddress string) []byte {
    orgBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(orgBytes, orgID)
    return append(orgBytes, []byte(userAddress)...)
}

// GetOrgPermissionPrefix gets the store prefix for all permissions in an org
func GetOrgPermissionPrefix(orgID uint64) []byte {
    orgBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(orgBytes, orgID)
    return orgBytes
}