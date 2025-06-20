package keeper

import (
	"context"

	"ipex/x/trust/types"

	"cosmossdk.io/collections"
)

// GetBlacklistLogCount get the total number of blacklistLog
func (k Keeper) GetBlacklistLogCount(ctx context.Context) uint64 {
	seq, err := k.BlacklistLogSequence.Peek(ctx)
	if err != nil {
		return 0
	}
	return seq
}

// SetBlacklistLogCount set the total number of blacklistLog (deprecated - use sequence)
func (k Keeper) SetBlacklistLogCount(ctx context.Context, count uint64) {
	// This is now handled by the sequence, no need to set manually
}

// AppendBlacklistLog appends a blacklistLog in the store with a new id and update the count
func (k Keeper) AppendBlacklistLog(
	ctx context.Context,
	blacklistLog types.BlacklistLog,
) uint64 {
	// Get next ID from sequence
	id, err := k.BlacklistLogSequence.Next(ctx)
	if err != nil {
		panic(err)
	}

	// Set the ID of the appended value
	blacklistLog.Id = id

	// Store the blacklist log
	if err := k.BlacklistLogs.Set(ctx, id, blacklistLog); err != nil {
		panic(err)
	}

	// Add to indexes
	k.addToIndexes(ctx, id, blacklistLog)

	return id
}

// SetBlacklistLog set a specific blacklistLog in the store
func (k Keeper) SetBlacklistLog(ctx context.Context, blacklistLog types.BlacklistLog) {
	// Get existing log to remove from old indexes if needed
	existingLog, found := k.GetBlacklistLog(ctx, blacklistLog.Id)
	if found {
		k.removeFromIndexes(ctx, blacklistLog.Id, existingLog)
	}

	// Store the blacklist log
	if err := k.BlacklistLogs.Set(ctx, blacklistLog.Id, blacklistLog); err != nil {
		panic(err)
	}

	// Add to indexes
	k.addToIndexes(ctx, blacklistLog.Id, blacklistLog)
}

// GetBlacklistLog returns a blacklistLog from its id
func (k Keeper) GetBlacklistLog(ctx context.Context, id uint64) (val types.BlacklistLog, found bool) {
	val, err := k.BlacklistLogs.Get(ctx, id)
	if err != nil {
		if err.Error() == "collections: not found" {
			return val, false
		}
		panic(err)
	}
	return val, true
}

// RemoveBlacklistLog removes a blacklistLog from the store
func (k Keeper) RemoveBlacklistLog(ctx context.Context, id uint64) {
	// Get existing log to remove from indexes
	existingLog, found := k.GetBlacklistLog(ctx, id)
	if found {
		k.removeFromIndexes(ctx, id, existingLog)
	}

	// Remove from main store
	if err := k.BlacklistLogs.Remove(ctx, id); err != nil {
		panic(err)
	}
}

// GetAllBlacklistLog returns all blacklistLog
func (k Keeper) GetAllBlacklistLog(ctx context.Context) (list []types.BlacklistLog) {
	err := k.BlacklistLogs.Walk(ctx, nil, func(key uint64, value types.BlacklistLog) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return list
}

// addToIndexes adds a blacklist log to all relevant indexes
func (k Keeper) addToIndexes(ctx context.Context, id uint64, log types.BlacklistLog) {
	// Add to subnet index
	subnetKey := collections.Join(log.SubnetId, id)
	if err := k.BlacklistLogsBySubnet.Set(ctx, subnetKey, true); err != nil {
		panic(err)
	}

	// Add to lessor index
	lessorKey := collections.Join(log.LessorOrgId, id)
	if err := k.BlacklistLogsByLessor.Set(ctx, lessorKey, true); err != nil {
		panic(err)
	}

	// Add to lessee index
	lesseeKey := collections.Join(log.LesseeOrgId, id)
	if err := k.BlacklistLogsByLessee.Set(ctx, lesseeKey, true); err != nil {
		panic(err)
	}

	// Add to lease index
	leaseKey := collections.Join(log.LeaseId, id)
	if err := k.BlacklistLogsByLease.Set(ctx, leaseKey, true); err != nil {
		panic(err)
	}
}

// removeFromIndexes removes a blacklist log from all relevant indexes
func (k Keeper) removeFromIndexes(ctx context.Context, id uint64, log types.BlacklistLog) {
	// Remove from subnet index
	subnetKey := collections.Join(log.SubnetId, id)
	k.BlacklistLogsBySubnet.Remove(ctx, subnetKey)

	// Remove from lessor index
	lessorKey := collections.Join(log.LessorOrgId, id)
	k.BlacklistLogsByLessor.Remove(ctx, lessorKey)

	// Remove from lessee index
	lesseeKey := collections.Join(log.LesseeOrgId, id)
	k.BlacklistLogsByLessee.Remove(ctx, lesseeKey)

	// Remove from lease index
	leaseKey := collections.Join(log.LeaseId, id)
	k.BlacklistLogsByLease.Remove(ctx, leaseKey)
}
