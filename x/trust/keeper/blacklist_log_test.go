package keeper_test

import (
	"context"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/trust/keeper"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

func createNBlacklistLog(keeper keeper.Keeper, ctx context.Context, n int) []types.BlacklistLog {
	items := make([]types.BlacklistLog, n)
	for i := range items {
		items[i].Id = keeper.AppendBlacklistLog(ctx, items[i])
	}
	return items
}

func TestBlacklistLogGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNBlacklistLog(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetBlacklistLog(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestBlacklistLogRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNBlacklistLog(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBlacklistLog(ctx, item.Id)
		// Check that the blacklist log was removed by checking if it exists in GetAllBlacklistLog
		allLogs := keeper.GetAllBlacklistLog(ctx)
		found := false
		for _, log := range allLogs {
			if log.Id == item.Id {
				found = true
				break
			}
		}
		require.False(t, found)
	}
}

func TestBlacklistLogGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNBlacklistLog(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBlacklistLog(ctx)),
	)
}

func TestBlacklistLogCount(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNBlacklistLog(keeper, ctx, 10)
	allLogs := keeper.GetAllBlacklistLog(ctx)
	require.Equal(t, uint64(len(items)), uint64(len(allLogs)))
}
