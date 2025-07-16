package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/lease/keeper"
	"ipex/x/lease/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNBid(keeper keeper.Keeper, ctx context.Context, n int) []types.Bid {
	items := make([]types.Bid, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetBid(ctx, items[i])
	}
	return items
}

func TestBidGet(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBid(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestBidRemove(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBid(ctx,
			item.Index,
		)
		_, found := keeper.GetBid(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestBidGetAll(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBid(ctx)),
	)
}
