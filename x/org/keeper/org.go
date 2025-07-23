package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/org/types"
)

// updateOrgIndexes updates all secondary indexes for an org
func (k Keeper) updateOrgIndexes(ctx context.Context, org types.Org) {
	// Index by admin
	err := k.OrgsByAdmin.Set(ctx, org.Admin, org.Id)
	if err != nil {
		panic(err)
	}

	// Index by creator
	err = k.OrgsByCreator.Set(ctx, org.Creator, org.Id)
	if err != nil {
		panic(err)
	}
}

// removeOrgFromIndexes removes org from all secondary indexes
func (k Keeper) removeOrgFromIndexes(ctx context.Context, org types.Org) {
	// Remove from admin index
	err := k.OrgsByAdmin.Remove(ctx, org.Admin)
	if err != nil {
		panic(err)
	}

	// Remove from creator index
	err = k.OrgsByCreator.Remove(ctx, org.Creator)
	if err != nil {
		panic(err)
	}
}

// AppendOrg appends a org in the store with a new id and update the count
func (k Keeper) AppendOrg(
	ctx context.Context,
	org types.Org,
) uint64 {
	// Get next ID from sequence
	id, err := k.OrgSequence.Next(ctx)
	if err != nil {
		panic(err)
	}
	org.Id = id

	// Store the org
	err = k.Orgs.Set(ctx, org.Id, org)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateOrgIndexes(ctx, org)

	// Initialize lessee reputation for new org
	k.trustKeeper.InitializeLesseeReputation(ctx, org.Id)
	return org.Id
}

// SetOrg set a specific org in the store
func (k Keeper) SetOrg(ctx context.Context, org types.Org) {
	// Get old org to handle index updates
	oldOrg, found := k.GetOrg(ctx, org.Id)
	if found {
		// Remove old indexes
		k.removeOrgFromIndexes(ctx, oldOrg)
	}

	// Store the org
	err := k.Orgs.Set(ctx, org.Id, org)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateOrgIndexes(ctx, org)
}

// GetOrg returns a org from its id
func (k Keeper) GetOrg(ctx context.Context, id uint64) (val types.Org, found bool) {
	org, err := k.Orgs.Get(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return val, false
		}
		panic(err)
	}
	return org, true
}

// RemoveOrg removes a org from the store
func (k Keeper) RemoveOrg(ctx context.Context, id uint64) {
	// Get the org first to access its indexed fields
	org, found := k.GetOrg(ctx, id)
	if !found {
		return
	}

	// Remove from main collection
	err := k.Orgs.Remove(ctx, id)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeOrgFromIndexes(ctx, org)

	// Remove lessee reputation
	k.trustKeeper.RemoveLesseeReputation(ctx, id)
}

// GetAllOrg returns all org
func (k Keeper) GetAllOrg(ctx context.Context) (list []types.Org) {
	err := k.Orgs.Walk(ctx, nil, func(key uint64, value types.Org) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

func (k Keeper) HasOrg(ctx context.Context, id uint64) bool {
	_, found := k.GetOrg(ctx, id)
	return found
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
	if !found || org.Admin == "" {
		return nil, false
	}

	adminAddr, err := sdk.AccAddressFromBech32(org.Admin)
	if err != nil {
		return nil, false
	}

	return adminAddr, true

}

func (k Keeper) GetOrgsForDomainValidation(ctx context.Context, referenceTime time.Time) (list []types.Org) {
	err := k.Orgs.Walk(ctx, nil, func(key uint64, org types.Org) (bool, error) {
		// Skip orgs without websites
		if org.Website == "" {
			return false, nil
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
				return false, nil
			}
		}

		// Add to the list for validation
		list = append(list, org)
		return false, nil
	})

	if err != nil {
		panic(err)
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

// IsOrgVerified returns org domain verification status
func (k Keeper) IsOrgVerified(ctx context.Context, id uint64) bool {
	org, _ := k.GetOrg(ctx, id)
	return org.DomainVerified
}
