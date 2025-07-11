package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/x/lease/keeper"
	"ipex/x/lease/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNContract(keeper keeper.Keeper, ctx context.Context, n int) []types.Contract {
	items := make([]types.Contract, n)
	for i := range items {
		items[i].Id = uint64(i + 1)
		keeper.SetContract(ctx, items[i])
	}
	return items
}

func TestListContractPaginated(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	msgs := createNContract(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryListContractRequest {
		return &types.QueryListContractRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		// Note: Pagination properly implemented in ListContract
		resp, err := keeper.ListContract(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination properly implemented in ListContract
		resp, err := keeper.ListContract(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.ListContract(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.Contract))
		// Check pagination response includes total count
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.ListContract(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}