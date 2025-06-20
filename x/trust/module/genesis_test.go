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
		BlacklistLogList: []types.BlacklistLog{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		BlacklistLogCount: 2,
		LeaseReputationList: []types.LeaseReputation{
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
	require.ElementsMatch(t, genesisState.BlacklistLogList, got.BlacklistLogList)
	require.Equal(t, genesisState.BlacklistLogCount, got.BlacklistLogCount)
	require.ElementsMatch(t, genesisState.LeaseReputationList, got.LeaseReputationList)
	// this line is used by starport scaffolding # genesis/test/assert
}
