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

func createNLessorReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.LessorReputation {
	items := make([]types.LessorReputation, n)
	for i := range items {
		items[i].Index = int32(i)

		keeper.SetLessorReputation(ctx, items[i])
	}
	return items
}

func TestLessorReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLessorReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLessorReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLessorReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetLessorReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestLessorReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLessorReputation(ctx)),
	)
}
