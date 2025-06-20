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

func createNLesseeReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.LesseeReputation {
	items := make([]types.LesseeReputation, n)
	for i := range items {
		items[i].Index = uint64(i)

		keeper.SetLesseeReputation(ctx, items[i])
	}
	return items
}

func TestLesseeReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLesseeReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLesseeReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLesseeReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetLesseeReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestLesseeReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLesseeReputation(ctx)),
	)
}
