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

func createNSubnetReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.SubnetReputation {
	items := make([]types.SubnetReputation, n)
	for i := range items {
		items[i].Index = uint64(i + 1)

		keeper.SetSubnetReputation(ctx, items[i])
	}
	return items
}

func TestSubnetReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNSubnetReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetSubnetReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestSubnetReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNSubnetReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSubnetReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetSubnetReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestSubnetReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNSubnetReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllSubnetReputation(ctx)),
	)
}
