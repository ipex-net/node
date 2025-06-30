package subnet

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/subnet/keeper"
	"ipex/x/subnet/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the subnet
	for _, elem := range genState.SubnetList {
		k.SetSubnet(ctx, elem)
	}

	// Set subnet sequence to match count from genesis
	if genState.SubnetCount > 0 {
		err := k.SubnetSequence.Set(ctx, genState.SubnetCount)
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

	genesis.SubnetList = k.GetAllSubnet(ctx)
	
	// Get current sequence value for subnet count
	count, err := k.SubnetSequence.Peek(ctx)
	if err != nil {
		count = 0
	}
	genesis.SubnetCount = count
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
