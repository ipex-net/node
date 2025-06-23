package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/abuse/keeper"
	"ipex/x/abuse/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNComplaint(keeper keeper.Keeper, ctx context.Context, n int) []types.Complaint {
	items := make([]types.Complaint, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetComplaint(ctx, items[i])
	}
	return items
}

func TestComplaintGet(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetComplaint(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestComplaintRemove(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveComplaint(ctx,
			item.Index,
		)
		_, found := keeper.GetComplaint(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestComplaintGetAll(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllComplaint(ctx)),
	)
}
