package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"ipex/x/subnet/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// Verify subnet user
func (k Keeper) VerifySubnetUser(ctx context.Context, subnetId uint64, user string) bool {

	// Validate the provided admin address
	_, err := sdk.AccAddressFromBech32(user)
	if err != nil {
		return false
	}

	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return false
	}

	//Checking that the user is admin of user organisation
	return k.orgKeeper.VerifyOrgAdmin(ctx, subnet.UserOrgId, user)
}

// Verify subnet owner
func (k Keeper) VerifySubnetOwner(ctx context.Context, subnetId uint64, owner string) bool {

	// Validate the provided admin address
	_, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return false
	}

	//Fetching the subnet
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return false
	}

	//Checking that the user is admin of user organisation
	return k.orgKeeper.VerifyOrgAdmin(ctx, subnet.OwnerOrgId, owner)
}

// Verify subnet owner
func (k Keeper) CheckSubnetLeaseAvailability(ctx context.Context, subnetId uint64) bool {

	//Fetching the subnet
	_, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return false
	}
	//TODO: Turn on when subnet status management is ready
	//if subnet.Status != types.SubnetStatus_SUBNET_STATUS_READY {
	//	return false
	//}
	return true
}

// Set subnet user
func (k Keeper) SetSubnetUserOrg(goCtx context.Context, subnetId uint64, userOrgId uint64) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// checking that user org exists
	if userOrgId == 0 || !k.orgKeeper.HasOrg(ctx, userOrgId) {
		return
	}
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return
	}
	subnet.UserOrgId = userOrgId
	subnet.UpdatedAt = ctx.BlockTime()
	k.SetSubnet(ctx, subnet)
}

func (k Keeper) RemoveSubnetUserOrg(goCtx context.Context, subnetId uint64) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return
	}
	subnet.OriginAsn = 0
	subnet.UserOrgId = 0
	subnet.UpdatedAt = ctx.BlockTime()
	subnet.OriginTransferRequestedAt = nil
	subnet.RpkiValidityChangedAt = nil
	subnet.RpkiOriginValid = false
	subnet.Status = types.SubnetStatus_SUBNET_STATUS_READY

	//REBUILD THIS ONCE subnet structure is ready
	k.SetSubnet(ctx, subnet)
}

func (k Keeper) GetSubnetsForRirSync(ctx context.Context) (list []types.Subnet) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	now := time.Now()
	count := 0
	maxSubnets := 5
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.Subnet
		k.cdc.MustUnmarshal(iterator.Value(), &subnet)

		// Check if enough time has passed since last attempt
		if subnet.RirSyncedAt != nil {
			lastAttempt := *subnet.RirSyncedAt

			// Determine required interval based on org age and validity status
			var requiredInterval time.Duration

			switch {
			
			case subnet.Status == types.SubnetStatus_SUBNET_STATUS_READY:
				requiredInterval = 72 * time.Hour //If subnet was succesfully synced, rerun the sync after x hours
			case subnet.Status == types.SubnetStatus_SUBNET_STATUS_LEASED:
				requiredInterval = 72 * time.Hour //If subnet was succesfully synced, rerun the sync after x hours
			case subnet.Status == types.SubnetStatus_SUBNET_STATUS_REJECTED:
				continue //If synced and was rejected
			case subnet.Status == types.SubnetStatus_SUBNET_STATUS_INITIAL:
				requiredInterval = time.Second
			default: // 1 minute
				requiredInterval = time.Minute
			}

			// Skip if not enough time has passed
			if now.Sub(lastAttempt) < requiredInterval {
				continue
			}
		}

		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			break 
		}
	}

	return
}


// UpdateOrgValidationStatus updates the domain validation status for an organization
func (k Keeper) UpdateSubnetRirDetails(ctx context.Context, subnetID uint64, info types.SubnetInfo) error {
	// Get the subnet
	subnet, found := k.GetSubnet(ctx, subnetID)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetID)
	}

	// If subnet is in initial status and rir record is invalid - we remove it
	if subnet.Status== types.SubnetStatus_SUBNET_STATUS_INITIAL && !info.Valid{
		k.RemoveSubnet(ctx, subnetID)
		return fmt.Errorf("Subnet with address %d failed RIR validation and was removed", subnet.NetworkAddress)
	}

	// Update subnet details
	subnet.RirRecordValid = info.Valid
	subnet.RirAbuseEmail = info.AbuseEmail
	subnet.RirOwnerOrgId = info.OrgID
	subnet.RirOwnerOrgName = info.OrgName
	subnet.RirAllocationType = info.AllocationType
	subnet.Rir = info.Rir

	// Change the status for initial subnet
	if subnet.Status== types.SubnetStatus_SUBNET_STATUS_INITIAL{
		subnet.Status = types.SubnetStatus_SUBNET_STATUS_READY
	}

	// Update timestamp
	nowTime := time.Now()
	subnet.RirSyncedAt = &nowTime
	subnet.UpdatedAt = nowTime
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}

// GetSubnetsForRPKIValidation retrieves subnets for rpki validation
func (k Keeper) GetSubnetsForRPKIValidation(ctx context.Context) (list []types.Subnet) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	now := time.Now()
	count := 0
	maxSubnets := 5
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.Subnet
		k.cdc.MustUnmarshal(iterator.Value(), &subnet)

		// If no origin set - don't validate
		if subnet.OriginAsn == 0 {
			continue
		}

		// If no origin transfer requested timestamp - skip
		if subnet.OriginTransferRequestedAt == nil {
			continue
		}

		originTransferAge := now.Sub(*subnet.OriginTransferRequestedAt)

		// Determine required interval
		var requiredInterval time.Duration

		switch {
			case subnet.RpkiOriginValid:
				requiredInterval = 6 * time.Hour
			case originTransferAge < 6 * time.Hour: // Org age < 1 day
				requiredInterval = 1 * time.Minute
			case originTransferAge < 24 * time.Hour: // Org age < 1 day
				requiredInterval = 5 * time.Minute
			default: 
				requiredInterval = time.Hour
		}

		// Skip if not enough time has passed
		if subnet.RpkiLastSyncedAt != nil && now.Sub(*subnet.RpkiLastSyncedAt) < requiredInterval {
			continue
		}
		
		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			break 
		}
	}

	return
}


// UpdateOrgValidationStatus updates the domain validation status for an organization
func (k Keeper) UpdateSubnetRpkiValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error {
	// Get the subnet
	subnet, found := k.GetSubnet(ctx, subnetID)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetID)
	}

	// Get the timestamp
	nowTime := time.Now()

	// If status changed - reflecting changes
	if subnet.RpkiOriginValid != isValid{
		subnet.RpkiOriginValid = isValid
		subnet.RpkiValidityChangedAt = &nowTime
	}
	// In any case we update timestamps
	subnet.UpdatedAt = nowTime
	subnet.RpkiLastSyncedAt = &nowTime
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}


// GetSubnetsForRPKIValidation retrieves subnets for rpki validation
func (k Keeper) GetSubnetsForGeoValidation(ctx context.Context) (list []types.Subnet) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	now := time.Now()
	count := 0
	maxSubnets := 5
	for ; iterator.Valid(); iterator.Next() {
		var subnet types.Subnet
		k.cdc.MustUnmarshal(iterator.Value(), &subnet)

		// If no country set - don't validate
		if subnet.Country == "" {
			continue
		}

		// Skip if no geo details change requested timestamp present
		if subnet.GeoDetailsChangedAt == nil {
			continue
		}

		// Skip if subnet already passed verification
		if subnet.TraceGeoVerified {
			continue
		}

		geoChangeAge := now.Sub(*subnet.GeoDetailsChangedAt)

		// Determine required interval
		var requiredInterval time.Duration

		switch {
			// If no verification attempted yet
			case subnet.GeoVerifAttemptedAt == nil:
				requiredInterval = 1 * time.Second
			// If geo change is recent - attempt the validation every hours
			case geoChangeAge < 72 * time.Hour: // Org age < 1 day
				requiredInterval = 1 * time.Hour
			case geoChangeAge < 240 * time.Hour: // Org age < 1 day
				requiredInterval = 6 * time.Hour
			default: 
				requiredInterval = 72 * time.Hour
		}

		// Skip if not enough time has passed
		if subnet.GeoVerifAttemptedAt != nil && now.Sub(*subnet.GeoVerifAttemptedAt) < requiredInterval {
			continue
		}
		
		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			break 
		}
	}

	return
}


// UpdateOrgValidationStatus updates the domain validation status for an organization
func (k Keeper) UpdateSubnetGeoValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error {
	// Get the subnet
	subnet, found := k.GetSubnet(ctx, subnetID)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetID)
	}

	// Get the timestamp
	nowTime := time.Now()

	// If status changed - reflecting changes
	if subnet.TraceGeoVerified != isValid{
		subnet.TraceGeoVerified = isValid
	}
	// In any case we update timestamps
	subnet.UpdatedAt = nowTime
	subnet.GeoVerifAttemptedAt = &nowTime
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}