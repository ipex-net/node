package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/subnet/types"
)

// updateSubnetIndexes updates all secondary indexes for a subnet
func (k Keeper) updateSubnetIndexes(ctx context.Context, subnet types.Subnet) {
	// Index by owner org (only if owner org is set)
	if subnet.OwnerOrgId != 0 {
		err := k.SubnetsByOwnerOrg.Set(ctx, collections.Join(subnet.OwnerOrgId, subnet.Id), true)
		if err != nil {
			panic(err)
		}
	}

	// Index by user org (only if user org is set)
	if subnet.UserOrgId != 0 {
		err := k.SubnetsByUserOrg.Set(ctx, collections.Join(subnet.UserOrgId, subnet.Id), true)
		if err != nil {
			panic(err)
		}
	}

	// Index by network address
	err := k.SubnetsByNetworkAddress.Set(ctx, collections.Join(subnet.NetworkAddress, subnet.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by origin check pending status (only if pending is true)
	if subnet.OriginCheckPending {
		err = k.SubnetOriginCheckPending.Set(ctx, subnet.Id, true)
		if err != nil {
			panic(err)
		}
	}
}

// removeSubnetFromIndexes removes subnet from all secondary indexes
func (k Keeper) removeSubnetFromIndexes(ctx context.Context, subnet types.Subnet) {
	// Remove from owner org index (only if owner org was set)
	if subnet.OwnerOrgId != 0 {
		err := k.SubnetsByOwnerOrg.Remove(ctx, collections.Join(subnet.OwnerOrgId, subnet.Id))
		if err != nil {
			panic(err)
		}
	}

	// Remove from user org index (only if user org was set)
	if subnet.UserOrgId != 0 {
		err := k.SubnetsByUserOrg.Remove(ctx, collections.Join(subnet.UserOrgId, subnet.Id))
		if err != nil {
			panic(err)
		}
	}

	// Remove from network address index
	err := k.SubnetsByNetworkAddress.Remove(ctx, collections.Join(subnet.NetworkAddress, subnet.Id))
	if err != nil {
		panic(err)
	}

	// Remove from origin check pending index (always attempt removal)
	err = k.SubnetOriginCheckPending.Remove(ctx, subnet.Id)
	if err != nil && !strings.Contains(err.Error(), "collections: not found") {
		panic(err)
	}
}

// AppendSubnet appends a subnet in the store with a new id and update the count
func (k Keeper) AppendSubnet(
	ctx context.Context,
	subnet types.Subnet,
) uint64 {
	// Get next ID from sequence
	id, err := k.SubnetSequence.Next(ctx)
	if err != nil {
		panic(err)
	}
	subnet.Id = id

	// Store the subnet
	err = k.Subnets.Set(ctx, subnet.Id, subnet)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateSubnetIndexes(ctx, subnet)

	// Initialize subnet reputation
	k.trustKeeper.InitializeSubnetReputation(ctx, subnet.Id)

	return subnet.Id
}

// SetSubnet set a specific subnet in the store
func (k Keeper) SetSubnet(ctx context.Context, subnet types.Subnet) {
	// Get old subnet to handle index updates
	oldSubnet, found := k.GetSubnet(ctx, subnet.Id)
	if found {
		// Handle origin check pending index changes specifically
		if oldSubnet.OriginCheckPending != subnet.OriginCheckPending {
			if oldSubnet.OriginCheckPending {
				// Remove from index if was pending
				err := k.SubnetOriginCheckPending.Remove(ctx, subnet.Id)
				if err != nil && !strings.Contains(err.Error(), "collections: not found") {
					panic(err)
				}
			}
		}
		
		// Remove old indexes (except origin check pending which we handled above)
		k.removeSubnetFromIndexes(ctx, oldSubnet)
	}

	// Store the subnet
	err := k.Subnets.Set(ctx, subnet.Id, subnet)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateSubnetIndexes(ctx, subnet)
}

// GetSubnet returns a subnet from its id
func (k Keeper) GetSubnet(ctx context.Context, id uint64) (val types.Subnet, found bool) {
	subnet, err := k.Subnets.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return val, false
		}
		panic(err)
	}
	return subnet, true
}

// RemoveSubnet removes a subnet from the store
func (k Keeper) RemoveSubnet(ctx context.Context, id uint64) {
	// Get the subnet first to access its indexed fields
	subnet, found := k.GetSubnet(ctx, id)
	if !found {
		return
	}

	// Remove from main collection
	err := k.Subnets.Remove(ctx, id)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeSubnetFromIndexes(ctx, subnet)

	// Remove subnet reputation
	k.trustKeeper.RemoveSubnetReputation(ctx, id)
}

// GetAllSubnet returns all subnet
func (k Keeper) GetAllSubnet(ctx context.Context) (list []types.Subnet) {
	err := k.Subnets.Walk(ctx, nil, func(key uint64, value types.Subnet) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
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

// CheckSubnetLeaseAvailability verifies subnet availability
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

// SetSubnetUserOrg sets subnet user and params
func (k Keeper) SetSubnetUserOrg(goCtx context.Context, subnetId uint64, userOrgId uint64, allowUserGeoChange bool, allowUserASNDROPOriginTransfer bool ) {
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
	subnet.AllowUserASNDROPOriginTransfer = allowUserASNDROPOriginTransfer
	subnet.AllowUserGeoChange = allowUserGeoChange
	subnet.UpdatedAt = ctx.BlockTime()
	k.SetSubnet(ctx, subnet)
}

// RemoveSubnetUserOrg removes the user from subnet and resets user params
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
	subnet.OriginCheckPending = false
	subnet.UserRequestedASNDROPOriginTransfer = false
	subnet.Status = types.SubnetStatus_SUBNET_STATUS_READY

	//REBUILD THIS ONCE subnet structure is ready
	k.SetSubnet(ctx, subnet)
}

// GetSubnetsForRirSync retrieves subnets for RIR sync
func (k Keeper) GetSubnetsForRirSync(ctx context.Context, referenceTime time.Time) (list []types.Subnet) {
	count := 0
	maxSubnets := 5

	err := k.Subnets.Walk(ctx, nil, func(key uint64, subnet types.Subnet) (bool, error) {
		// Check if enough time has passed since last attempt
		if subnet.RirSyncedAt != nil {
			lastAttempt := *subnet.RirSyncedAt

			// Determine required interval based on org age and validity status
			var requiredInterval time.Duration

			switch subnet.Status {

			case types.SubnetStatus_SUBNET_STATUS_READY:
				requiredInterval = 72 * time.Hour //If subnet was succesfully synced, rerun the sync after x hours
			case types.SubnetStatus_SUBNET_STATUS_LEASED:
				requiredInterval = 72 * time.Hour //If subnet was succesfully synced, rerun the sync after x hours
			case types.SubnetStatus_SUBNET_STATUS_REJECTED:
				return false, nil //If synced and was rejected
			case types.SubnetStatus_SUBNET_STATUS_INITIAL:
				requiredInterval = time.Second
			default: // 1 minute
				requiredInterval = time.Minute
			}

			// Skip if not enough time has passed
			if referenceTime.Sub(lastAttempt) < requiredInterval {
				return false, nil
			}
		}

		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			return true, nil // stop iteration
		}

		return false, nil
	})

	if err != nil {
		panic(err)
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
	if subnet.Status == types.SubnetStatus_SUBNET_STATUS_INITIAL && !info.Valid {
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
	if subnet.Status == types.SubnetStatus_SUBNET_STATUS_INITIAL {
		subnet.Status = types.SubnetStatus_SUBNET_STATUS_READY
	}

	// Update timestamp
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()
	subnet.RirSyncedAt = &now
	subnet.UpdatedAt = now
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}

// GetSubnetsForRPKIValidation retrieves subnets for rpki validation
func (k Keeper) GetSubnetsForRPKIValidation(ctx context.Context, referenceTime time.Time) (list []types.Subnet) {
	count := 0
	maxSubnets := 5

	err := k.Subnets.Walk(ctx, nil, func(key uint64, subnet types.Subnet) (bool, error) {
		// If no origin set - don't validate
		if subnet.OriginAsn == 0 {
			return false, nil
		}

		// If no origin transfer requested timestamp - skip
		if subnet.OriginTransferRequestedAt == nil {
			return false, nil
		}

		// If origin check is pending - skip
		if subnet.OriginCheckPending {
			return false, nil
		}

		originTransferAge := referenceTime.Sub(*subnet.OriginTransferRequestedAt)

		// Determine required interval
		var requiredInterval time.Duration

		switch {
		case subnet.RpkiOriginValid:
			requiredInterval = 6 * time.Hour
		case originTransferAge < 6*time.Hour: // Org age < 1 day
			requiredInterval = 1 * time.Minute
		case originTransferAge < 24*time.Hour: // Org age < 1 day
			requiredInterval = 5 * time.Minute
		default:
			requiredInterval = time.Hour
		}

		// Skip if not enough time has passed
		if subnet.RpkiLastSyncedAt != nil && referenceTime.Sub(*subnet.RpkiLastSyncedAt) < requiredInterval {
			return false, nil
		}

		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			return true, nil // stop iteration
		}

		return false, nil
	})

	if err != nil {
		panic(err)
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// If status changed - reflecting changes
	if subnet.RpkiOriginValid != isValid {
		subnet.RpkiOriginValid = isValid
		subnet.RpkiValidityChangedAt = &now
	}
	// In any case we update timestamps
	subnet.UpdatedAt = now
	subnet.RpkiLastSyncedAt = &now
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}

// GetSubnetsForGeoValidation retrieves subnets for geo validation
func (k Keeper) GetSubnetsForGeoValidation(ctx context.Context, referenceTime time.Time) (list []types.Subnet) {
	count := 0
	maxSubnets := 5

	err := k.Subnets.Walk(ctx, nil, func(key uint64, subnet types.Subnet) (bool, error) {
		// If no country set - don't validate
		if subnet.Country == "" {
			return false, nil
		}

		// Skip if no geo details change requested timestamp present
		if subnet.GeoDetailsChangedAt == nil {
			return false, nil
		}

		// Skip if subnet already passed verification
		if subnet.TraceGeoVerified {
			return false, nil
		}

		geoChangeAge := referenceTime.Sub(*subnet.GeoDetailsChangedAt)

		// Determine required interval
		var requiredInterval time.Duration

		switch {
		// If geo change is recent - attempt the validation every hours
		case geoChangeAge < 72*time.Hour:
			requiredInterval = 1 * time.Hour
		case geoChangeAge < 240*time.Hour:
			requiredInterval = 6 * time.Hour
		default:
			requiredInterval = 72 * time.Hour
		}

		// Skip if not enough time has passed
		if subnet.GeoVerifAttemptedAt != nil {
			timeSinceLastAttempt := referenceTime.Sub(*subnet.GeoVerifAttemptedAt)
			if timeSinceLastAttempt < requiredInterval {
				return false, nil
			}
		}

		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			return true, nil // stop iteration
		}

		return false, nil
	})

	if err != nil {
		panic(err)
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// If status changed - reflecting changes
	if subnet.TraceGeoVerified != isValid {
		subnet.TraceGeoVerified = isValid
	}
	// In any case we update timestamps
	subnet.UpdatedAt = now
	subnet.GeoVerifAttemptedAt = &now
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}

// GetSubnetsForOwnershipVerification retrieves subnets for ownership verification
func (k Keeper) GetSubnetsForOwnershipVerification(ctx context.Context, referenceTime time.Time) (list []types.Subnet) {
	count := 0
	maxSubnets := 5

	err := k.Subnets.Walk(ctx, nil, func(key uint64, subnet types.Subnet) (bool, error) {
		var lastAttempt time.Time
		if subnet.OwnershipVerifLastAttemptedAt != nil {
			lastAttempt = *subnet.OwnershipVerifLastAttemptedAt
		} else {
			lastAttempt = time.Time{}
		}
		subnetAge := referenceTime.Sub(subnet.CreatedAt)

		// Determine required interval based on subnet age and validity status
		var requiredInterval time.Duration

		// While the subnet is in active lease we don't want to perform verification
		if subnet.OwnershipVerified && subnet.Status == types.SubnetStatus_SUBNET_STATUS_LEASED {
			return false, nil
		}

		switch {
		// If ownership was verified previously - we re-check every 3 days
		case subnet.OwnershipVerified:
			requiredInterval = 72 * time.Hour //If subnet is verified we run re-verification every 3 days
		// We want to gradually decrease the verification intervals for increased subnet ages
		case subnetAge < 24*time.Hour: // Org age < 1 day
			requiredInterval = 1 * time.Minute
		case subnetAge < 7*24*time.Hour: // < 7 days
			requiredInterval = 15 * time.Minute
		case subnetAge < 14*24*time.Hour: // < 14 days
			requiredInterval = 30 * time.Minute
		case subnetAge < 30*24*time.Hour: // < 30 days
			requiredInterval = 1 * time.Hour
		default: // > 30 days
			requiredInterval = 72 * time.Hour
		}

		// Skip if not enough time has passed
		if referenceTime.Sub(lastAttempt) < requiredInterval {
			return false, nil
		}

		// Add to the list for validation
		list = append(list, subnet)
		// Increment count and check if we've reached the maximum
		count++
		if count >= maxSubnets {
			return true, nil // stop iteration
		}

		return false, nil
	})

	if err != nil {
		panic(err)
	}

	return
}

// UpdateOrgValidationStatus updates the domain validation status for an organization
func (k Keeper) UpdateSubnetOwnershipVerification(ctx context.Context, subnetId uint64, isValid bool) error {
	// Get the subnet
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetId)
	}
	// Get the timestamp
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime()

	// If status changed - reflecting changes
	if subnet.OwnershipVerified != isValid {
		subnet.OwnershipVerified = isValid
		subnet.OwnershipVerifChangedAt = &now
	}
	// In any case we update timestamps
	subnet.UpdatedAt = now
	subnet.OwnershipVerifLastAttemptedAt = &now
	// Save the updated organization
	k.SetSubnet(ctx, subnet)

	return nil
}

// IsRirRecordValid returns the RIR record validity
func (k Keeper) IsRirRecordValid(ctx context.Context, subnetId uint64) bool {
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return false
	}
	return subnet.RirRecordValid
}

// SetSubnetOriginCheckPending sets a subnet's origin check pending status to true
func (k Keeper) SetSubnetOriginCheckPending(ctx context.Context, subnetId uint64) error {
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetId)
	}

	if subnet.OriginCheckPending {
		return nil
	}

	// Update the subnet
	subnet.OriginCheckPending = true
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnet.UpdatedAt = sdkCtx.BlockTime()
	
	// Store updated subnet
	k.SetSubnet(ctx, subnet)

	return nil
}

// ClearSubnetOriginCheckPending sets a subnet's origin check pending status to false
func (k Keeper) ClearSubnetOriginCheckPending(ctx context.Context, subnetId uint64) error {
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetId)
	}

	if !subnet.OriginCheckPending {
		return nil
	}

	// Update the subnet
	subnet.OriginCheckPending = false
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnet.UpdatedAt = sdkCtx.BlockTime()
	
	// Store updated subnet
	k.SetSubnet(ctx, subnet)

	return nil
}

// GetOriginCheckPendingSubnets returns subnet IDs that have origin check pending
func (k Keeper) GetOriginCheckPendingSubnets(ctx context.Context, limit int) ([]uint64) {
	var subnetIDs []uint64
	count := 0
	
	err := k.SubnetOriginCheckPending.Walk(ctx, nil, func(subnetId uint64, _ bool) (bool, error) {
		subnetIDs = append(subnetIDs, subnetId)
		count++
		
		// Stop if we've reached the limit
		if limit > 0 && count >= limit {
			return true, nil
		}
		
		return false, nil
	})
	
	if err != nil {
		return nil
	}
	
	return subnetIDs
}

// FlagUserRequestedASNDROPOrigin sets up an indicator that user requested an origin transfer to origin from ASNDROP list. This flag stays on until the end of the lease.
func (k Keeper) FlagUserRequestedASNDROPOrigin(ctx context.Context, subnetId uint64) error {
	subnet, found := k.GetSubnet(ctx, subnetId)
	if !found {
		return fmt.Errorf("subnet with ID %d not found", subnetId)
	}

	if subnet.UserRequestedASNDROPOriginTransfer {
		return nil
	}
	// Reset origin
	subnet.OriginAsn = 0
	// Reset verification tag
	subnet.OriginCheckPending = false
	// Set ASNDROP flag
	subnet.UserRequestedASNDROPOriginTransfer = true
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	subnet.UpdatedAt = sdkCtx.BlockTime()
	
	// Store updated subnet
	k.SetSubnet(ctx, subnet)

	return nil
}