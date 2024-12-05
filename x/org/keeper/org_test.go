package keeper_test

import (
	"context"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/org/keeper"
	"ipex/x/org/types"

	"github.com/stretchr/testify/require"
)

func createNOrg(keeper keeper.Keeper, ctx context.Context, n int) []types.Org {
	items := make([]types.Org, n)
	for i := range items {
		items[i].Id = keeper.AppendOrg(ctx, items[i])
	}
	return items
}

func TestOrgGet(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	items := createNOrg(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetOrg(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestOrgRemove(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	items := createNOrg(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveOrg(ctx, item.Id)
		_, found := keeper.GetOrg(ctx, item.Id)
		require.False(t, found)
	}
}

func TestOrgGetAll(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	items := createNOrg(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllOrg(ctx)),
	)
}

func TestOrgCount(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	items := createNOrg(keeper, ctx, 10)
	count := uint64(len(items))
	require.Equal(t, count, keeper.GetOrgCount(ctx))
}
