package keeper

import (
	"context"

	"ipex/x/trust/types"
)

func (k Keeper) SetLesseeReputation(ctx context.Context, lesseeReputation types.LesseeReputation) {
	k.LesseeReputations.Set(ctx, lesseeReputation.Index, lesseeReputation)
}

func (k Keeper) GetLesseeReputation(ctx context.Context, index uint64) (val types.LesseeReputation, found bool) {
	val, err := k.LesseeReputations.Get(ctx, index)
	if err != nil {
		return val, false
	}
	return val, true
}

func (k Keeper) RemoveLesseeReputation(ctx context.Context, index uint64) {
	k.LesseeReputations.Remove(ctx, index)
}

func (k Keeper) GetAllLesseeReputation(ctx context.Context) (list []types.LesseeReputation) {
	err := k.LesseeReputations.Walk(ctx, nil, func(key uint64, value types.LesseeReputation) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitializeLesseeReputation initializes a new lessee reputation entry when org is created
func (k Keeper) InitializeLesseeReputation(ctx context.Context, lesseeId uint64) {

	// Create new lesee reputation with initial values
	lesseeReputation := types.LesseeReputation{
		Index:       lesseeId,
		Level1Count: 0, // Initialize all counts to zero
		Level2Count: 0,
		Level3Count: 0,
		Level4Count: 0,
		LastUpdated: nil,
	}

	// Set the subnet reputation in the store
	k.SetLesseeReputation(ctx, lesseeReputation)

}
