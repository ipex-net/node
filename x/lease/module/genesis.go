package lease

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/lease/keeper"
	"ipex/x/lease/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Initialize contract sequence to start from 1
	err := k.ContractSequence.Set(ctx, 1)
	if err != nil {
		panic(err)
	}

	// Set all the bid
	for _, elem := range genState.BidList {
		k.SetBid(ctx, elem)
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

	genesis.BidList = k.GetAllBid(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
