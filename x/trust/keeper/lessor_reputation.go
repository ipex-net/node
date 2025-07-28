package keeper

import (
	"context"

	"ipex/x/trust/types"
)

func (k Keeper) SetLessorReputation(ctx context.Context, lessorReputation types.LessorReputation) {
	k.LessorReputations.Set(ctx, lessorReputation.Index, lessorReputation)
}

func (k Keeper) GetLessorReputation(ctx context.Context, index uint64) (val types.LessorReputation, found bool) {
	val, err := k.LessorReputations.Get(ctx, index)
	if err != nil {
		return val, false
	}
	return val, true
}

func (k Keeper) RemoveLessorReputation(ctx context.Context, index uint64) {
	k.LessorReputations.Remove(ctx, index)
}

func (k Keeper) GetAllLessorReputation(ctx context.Context) (list []types.LessorReputation) {
	err := k.LessorReputations.Walk(ctx, nil, func(key uint64, value types.LessorReputation) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitializeLessorReputation initializes a new lessor reputation entry when org is created
func (k Keeper) InitializeLessorReputation(ctx context.Context, lessorId uint64) {
	// Create new lessor reputation with initial values
	lessorReputation := types.LessorReputation{
		Index:                       lessorId,
		IpMonthsTotal:               0,
		IpsLeasedNow:                0,
		ContractsActive:             0,
		ContractsTerminated:         0,
		ContractsTerminatedViolated: 0,
	}

	// Set the lessor reputation in the store
	k.SetLessorReputation(ctx, lessorReputation)
}
