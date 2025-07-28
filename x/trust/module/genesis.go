package trust

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/trust/keeper"
	"ipex/x/trust/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the subnetReputation
	for _, elem := range genState.SubnetReputationList {
		k.SetSubnetReputation(ctx, elem)
	}
	// Set all the blacklistLog
	for _, elem := range genState.BlacklistLogList {
		k.SetBlacklistLog(ctx, elem)
	}

	// Initialize blacklist log sequence with the count from genesis
	// Ensure sequence starts from 1 if no blacklist logs exist
	sequenceValue := genState.BlacklistLogCount
	if sequenceValue == 0 {
		sequenceValue = 1
	}
	err := k.BlacklistLogSequence.Set(ctx, sequenceValue)
	if err != nil {
		panic(err)
	}
	// Set all the leaseReputation
	for _, elem := range genState.LeaseReputationList {
		k.SetLeaseReputation(ctx, elem)
	}
	// Set all the lesseeReputation
	for _, elem := range genState.LesseeReputationList {
		k.SetLesseeReputation(ctx, elem)
	}
	// Set all the lessorReputation
	for _, elem := range genState.LessorReputationList {
		k.SetLessorReputation(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.SubnetReputationList = k.GetAllSubnetReputation(ctx)
	genesis.BlacklistLogList = k.GetAllBlacklistLog(ctx)
	// Export current blacklist log sequence value
	sequenceValue, err := k.BlacklistLogSequence.Peek(ctx)
	if err != nil {
		// If no sequence exists yet, start at 0
		genesis.BlacklistLogCount = 0
	} else {
		genesis.BlacklistLogCount = sequenceValue
	}
	genesis.LeaseReputationList = k.GetAllLeaseReputation(ctx)
	genesis.LesseeReputationList = k.GetAllLesseeReputation(ctx)
	genesis.LessorReputationList = k.GetAllLessorReputation(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
