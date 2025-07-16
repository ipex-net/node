package abuse

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/abuse/keeper"
	"ipex/x/abuse/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the complaint
	for _, elem := range genState.ComplaintList {
		k.SetComplaint(ctx, elem)
	}

	// Initialize complaint sequence to start from 1
	if len(genState.ComplaintList) == 0 {
		err := k.ComplaintSequence.Set(ctx, 1)
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

	genesis.ComplaintList = k.GetAllComplaint(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
