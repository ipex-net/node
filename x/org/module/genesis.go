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

	// Set org sequence to match count from genesis
	if genState.OrgCount > 0 {
		err := k.OrgSequence.Set(ctx, genState.OrgCount)
		if err != nil {
			panic(err)
		}
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
