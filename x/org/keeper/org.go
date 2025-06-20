package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

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

	// Set the ID of the appended value (start from 1, not 0)
	org.Id = count + 1

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	appendedValue := k.cdc.MustMarshal(&org)
	store.Set(GetOrgIDBytes(org.Id), appendedValue)

	// Update org count
	k.SetOrgCount(ctx, count+1)

	// Initialize lessee reputation for new org
	k.trustKeeper.InitializeLesseeReputation(ctx, org.Id)
	return org.Id
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
	
	// Remove lessee reputation
	k.trustKeeper.RemoveLesseeReputation(ctx, id)
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

func (k Keeper) GetOrgTreasury(ctx context.Context, orgid uint64) (sdk.AccAddress, bool) {
	org, found := k.GetOrg(ctx, orgid)
	if !found || org.Treasury == "" {
		return nil, false
	}

	treasuryAddr, err := sdk.AccAddressFromBech32(org.Treasury)
	if err != nil {
		return nil, false
	}

	return treasuryAddr, true

}

func (k Keeper) GetOrgAdmin(ctx context.Context, orgid uint64) (sdk.AccAddress, bool) {
	org, found := k.GetOrg(ctx, orgid)
	if !found || org.Treasury == "" {
		return nil, false
	}

	treasuryAddr, err := sdk.AccAddressFromBech32(org.Admin)
	if err != nil {
		return nil, false
	}

	return treasuryAddr, true

}

func (k Keeper) GetOrgsForDomainValidation(ctx context.Context, referenceTime time.Time) (list []types.Org) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.OrgKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var org types.Org
		k.cdc.MustUnmarshal(iterator.Value(), &org)

		// Skip orgs without websites
		if org.Website == "" {
			continue
		}

		// Check if enough time has passed since last attempt
		if org.DomainVerifAttemptedAt != nil {
			lastAttempt := *org.DomainVerifAttemptedAt
			orgAge := referenceTime.Sub(org.CreatedAt)

			// Determine required interval based on org age and validity status
			var requiredInterval time.Duration

			switch {

			case org.DomainVerified:
				requiredInterval = 72 * time.Hour //If org is verified we run re-verification every 3 days
			case orgAge < 24*time.Hour: // Org age < 1 day
				requiredInterval = 1 * time.Minute
			case orgAge < 72*time.Hour: // < 3 days
				requiredInterval = 3 * time.Minute
			case orgAge < 7*24*time.Hour: // < 7 days
				requiredInterval = 5 * time.Minute
			case orgAge < 14*24*time.Hour: // < 14 days
				requiredInterval = 10 * time.Minute
			case orgAge < 30*24*time.Hour: // < 30 days
				requiredInterval = 30 * time.Minute
			default: // > 30 days
				requiredInterval = 4 * time.Hour
			}

			// Skip if not enough time has passed
			if referenceTime.Sub(lastAttempt) < requiredInterval {
				continue
			}
		}

		// Add to the list for validation
		list = append(list, org)
	}

	return
}

// UpdateOrgValidationStatus updates the domain validation status for an organization
func (k Keeper) UpdateOrgDomainValidationStatus(ctx context.Context, orgID uint64, isValid bool) error {
	// Get the organization
	org, found := k.GetOrg(ctx, orgID)
	if !found {
		return fmt.Errorf("org with ID %d not found", orgID)
	}

	// Update validation status
	org.DomainVerified = isValid

	// Update timestamp
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()
	if isValid {
		org.DomainVerifiedAt = &now
	}

	//Updating verification attempt time
	org.DomainVerifAttemptedAt = &now

	// Save the updated organization
	k.SetOrg(ctx, org)

	return nil
}
