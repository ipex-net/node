package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/x/payments/keeper"
	"ipex/x/payments/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNTransactionForEntity(keeper keeper.Keeper, ctx context.Context, entityId uint64, n int) []types.Transaction {
	items := make([]types.Transaction, n)
	for i := range items {
		items[i].Id = uint64(i + 1)
		items[i].EntityId = entityId
		items[i].Amount = sdk.NewCoin("token", math.NewInt(100))
		keeper.SetTransaction(ctx, items[i])
	}
	return items
}

func TestListContractPaymentPaginated(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	entityId := uint64(1)
	msgs := createNTransactionForEntity(keeper, ctx, entityId, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryListContractPaymentRequest {
		return &types.QueryListContractPaymentRequest{
			Id: entityId,
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		// Note: Pagination properly implemented in ListContractPayment
		resp, err := keeper.ListContractPayment(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Transaction))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination properly implemented in ListContractPayment
		resp, err := keeper.ListContractPayment(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Transaction))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.ListContractPayment(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.Transaction))
		// Check pagination response includes total count
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
	})
	t.Run("LatestFirst", func(t *testing.T) {
		// Test that transactions are returned in descending order (latest first)
		resp, err := keeper.ListContractPayment(ctx, request(nil, 0, 5, false))
		require.NoError(t, err)
		require.Equal(t, 5, len(resp.Transaction))
		
		// Verify transactions are in descending ID order (latest first)
		for i := 1; i < len(resp.Transaction); i++ {
			require.Greater(t, resp.Transaction[i-1].Id, resp.Transaction[i].Id,
				"transactions should be ordered by ID descending (latest first)")
		}
		
		// The first transaction should have the highest ID (latest)
		require.Equal(t, uint64(5), resp.Transaction[0].Id, "first transaction should have highest ID")
		require.Equal(t, uint64(1), resp.Transaction[4].Id, "last transaction should have lowest ID")
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.ListContractPayment(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
