package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/trust/keeper"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNLeaseReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.LeaseReputation {
	items := make([]types.LeaseReputation, n)
	for i := range items {
		items[i].Index = uint64(i)

		keeper.SetLeaseReputation(ctx, items[i])
	}
	return items
}

func TestLeaseReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLeaseReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLeaseReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLeaseReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLeaseReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLeaseReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetLeaseReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestLeaseReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLeaseReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLeaseReputation(ctx)),
	)
}
