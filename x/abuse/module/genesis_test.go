package abuse_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	abuse "ipex/x/abuse/module"
	"ipex/x/abuse/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AbuseKeeper(t)
	abuse.InitGenesis(ctx, k, genesisState)
	got := abuse.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
