package org_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	org "ipex/x/org/module"
	"ipex/x/org/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		OrgList: []types.Org{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		OrgCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.OrgKeeper(t)
	org.InitGenesis(ctx, k, genesisState)
	got := org.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.OrgList, got.OrgList)
	require.Equal(t, genesisState.OrgCount, got.OrgCount)
	// this line is used by starport scaffolding # genesis/test/assert
}
