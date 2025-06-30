package payments

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/payments/keeper"
	"ipex/x/payments/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the transaction
	for _, elem := range genState.TransactionList {
		k.SetTransaction(ctx, elem)
	}

	// Initialize transaction sequence with the count from genesis
	// Ensure sequence starts from 1 if no transactions exist
	sequenceValue := genState.TransactionCount
	if sequenceValue == 0 {
		sequenceValue = 1
	}
	err := k.TransactionSequence.Set(ctx, sequenceValue)
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

	genesis.TransactionList = k.GetAllTransaction(ctx)
	// Export current transaction sequence value
	sequenceValue, err := k.TransactionSequence.Peek(ctx)
	if err != nil {
		// If no sequence exists yet, start at 0
		genesis.TransactionCount = 0
	} else {
		genesis.TransactionCount = sequenceValue
	}
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
