package keeper

import (
	"context"

	"ipex/x/trust/types"
)

func (k Keeper) SetLeaseReputation(ctx context.Context, leaseReputation types.LeaseReputation) {
	k.LeaseReputations.Set(ctx, leaseReputation.Index, leaseReputation)
}

func (k Keeper) GetLeaseReputation(ctx context.Context, index uint64) (val types.LeaseReputation, found bool) {
	val, err := k.LeaseReputations.Get(ctx, index)
	if err != nil {
		return val, false
	}
	return val, true
}

func (k Keeper) RemoveLeaseReputation(ctx context.Context, index uint64) {
	k.LeaseReputations.Remove(ctx, index)
}

func (k Keeper) GetAllLeaseReputation(ctx context.Context) (list []types.LeaseReputation) {
	err := k.LeaseReputations.Walk(ctx, nil, func(key uint64, value types.LeaseReputation) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitializeLeaseReputation initializes a new lease reputation entry when the lease is created
func (k Keeper) InitializeLeaseReputation(ctx context.Context, leaseId uint64) {
	// Create new lease reputation with initial values
	// Even if subnet has some blacklist entries, we initialize all lease reputation blacklist counts to zero
	// This reputation only tracks changes related to lease, not the absolute blacklist counts per subnet
	leaseReputation := types.LeaseReputation{
		Index:       leaseId,
		Level1Count: 0,
		Level2Count: 0,
		Level3Count: 0,
		Level4Count: 0,
		LastUpdated: nil,
	}

	// Set the subnet reputation in the store
	k.SetLeaseReputation(ctx, leaseReputation)

}

// GetShortLeaseReputation retrieves lease reputation data including L1 listing status, L2 count, failed abuse count, and existence flag
func (k Keeper) GetShortLeaseReputation(ctx context.Context, index uint64) (l1Listed bool, l2Count int32, failedAbuse uint64, found bool) {
	reputation, err := k.LeaseReputations.Get(ctx, index)
	if err != nil {
		return false, 0, 0, false
	}
	l1Listed = false
	if reputation.Level1Count != 0 {
		l1Listed = true
	}
	return l1Listed, reputation.Level2Count, reputation.AbuseFailed, true
}