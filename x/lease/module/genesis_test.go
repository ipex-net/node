package lease_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	lease "ipex/x/lease/module"
	"ipex/x/lease/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		BidList: []types.Bid{
			{
				ContractId:  1,
				BidderOrgId: 100,
			},
			{
				ContractId:  2,
				BidderOrgId: 101,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.LeaseKeeper(t)
	lease.InitGenesis(ctx, k, genesisState)
	got := lease.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.BidList, got.BidList)
	// this line is used by starport scaffolding # genesis/test/assert
}
