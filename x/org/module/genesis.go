package org

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/org/keeper"
	"ipex/x/org/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the org
	for _, elem := range genState.OrgList {
		k.SetOrg(ctx, elem)
	}

	// Initialize org sequence with the count from genesis
	// Ensure sequence starts from 1 if no orgs exist
	sequenceValue := genState.OrgCount
	if sequenceValue == 0 {
		sequenceValue = 1
	}
	err := k.OrgSequence.Set(ctx, sequenceValue)
	if err != nil {
		panic(err)
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

	genesis.OrgList = k.GetAllOrg(ctx)

	// Get current sequence value for org count
	count, err := k.OrgSequence.Peek(ctx)
	if err != nil {
		count = 0
	}
	genesis.OrgCount = count
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
