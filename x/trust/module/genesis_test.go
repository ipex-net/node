package trust_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	trust "ipex/x/trust/module"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SubnetReputationList: []types.SubnetReputation{
			{
				Index: 0,
			},
			{
				Index: 1,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.TrustKeeper(t)
	trust.InitGenesis(ctx, k, genesisState)
	got := trust.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.SubnetReputationList, got.SubnetReputationList)
	// this line is used by starport scaffolding # genesis/test/assert
}
