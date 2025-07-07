package keeper_test

import (
	"context"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/payments/keeper"
	"ipex/x/payments/types"

	"github.com/stretchr/testify/require"
)

func createNTransaction(keeper keeper.Keeper, ctx context.Context, n int) []types.Transaction {
	items := make([]types.Transaction, n)
	for i := range items {
		items[i].Id = keeper.AppendTransaction(ctx, items[i])
	}
	return items
}

func TestTransactionGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	items := createNTransaction(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetTransaction(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestTransactionRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	items := createNTransaction(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveTransaction(ctx, item.Id)
		// Check that the transaction was removed by checking if it exists in GetAllTransaction
		allTransactions := keeper.GetAllTransaction(ctx)
		found := false
		for _, transaction := range allTransactions {
			if transaction.Id == item.Id {
				found = true
				break
			}
		}
		require.False(t, found)
	}
}

func TestTransactionGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	items := createNTransaction(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllTransaction(ctx)),
	)
}

func TestTransactionCount(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	items := createNTransaction(keeper, ctx, 10)
	count := uint64(len(items))
	allTransactions := keeper.GetAllTransaction(ctx)
	require.Equal(t, count, uint64(len(allTransactions)))
}
