package payments_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	payments "ipex/x/payments/module"
	"ipex/x/payments/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.PaymentsKeeper(t)
	payments.InitGenesis(ctx, k, genesisState)
	got := payments.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
