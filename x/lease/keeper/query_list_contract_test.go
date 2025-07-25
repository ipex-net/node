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
	
	// Create multiple contracts for pagination testing
	contracts := []types.Contract{
		{Id: 1, SubnetId: 101, OwnerOrg: 201, LesseeOrg: 301},
		{Id: 2, SubnetId: 102, OwnerOrg: 202, LesseeOrg: 302},
		{Id: 3, SubnetId: 103, OwnerOrg: 203, LesseeOrg: 303},
		{Id: 4, SubnetId: 104, OwnerOrg: 204, LesseeOrg: 304},
		{Id: 5, SubnetId: 105, OwnerOrg: 205, LesseeOrg: 305},
		{Id: 6, SubnetId: 106, OwnerOrg: 206, LesseeOrg: 306},
	}

	for _, contract := range contracts {
		keeper.SetContract(ctx, contract)
	}

	request := func(next []byte, limit uint64) *types.QueryListContractRequest {
		return &types.QueryListContractRequest{
			Pagination: &query.PageRequest{
				Key:   next,
				Limit: limit,
			},
		}
	}

	t.Run("WithLimit", func(t *testing.T) {
		resp, err := keeper.ListContract(ctx, request(nil, 3))
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Contract))
		require.NotNil(t, resp.Pagination.NextKey)
	})

	t.Run("WithPagination", func(t *testing.T) {
		// First page
		resp1, err := keeper.ListContract(ctx, request(nil, 2))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp1.Contract))
		require.NotNil(t, resp1.Pagination.NextKey)

		// Second page
		resp2, err := keeper.ListContract(ctx, request(resp1.Pagination.NextKey, 2))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp2.Contract))
		require.NotNil(t, resp2.Pagination.NextKey)

		// Third page
		resp3, err := keeper.ListContract(ctx, request(resp2.Pagination.NextKey, 2))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp3.Contract))
		// Note: ListContract implementation sets nextKey when collectedCount == limit,
		// even on the last page. This is different from bid query implementation.

		// Verify that the next request returns no more items
		resp4, err := keeper.ListContract(ctx, request(resp3.Pagination.NextKey, 2))
		require.NoError(t, err)
		require.Equal(t, 0, len(resp4.Contract))
		require.Nil(t, resp4.Pagination.NextKey)
	})

	t.Run("AllResults", func(t *testing.T) {
		resp, err := keeper.ListContract(ctx, request(nil, 0)) // No limit
		require.NoError(t, err)
		require.Equal(t, len(contracts), len(resp.Contract))
		require.Nil(t, resp.Pagination.NextKey)
	})

	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.ListContract(ctx, &types.QueryListContractRequest{
			Pagination: &query.PageRequest{
				CountTotal: true,
			},
		})
		require.NoError(t, err)
		require.Equal(t, len(contracts), len(resp.Contract))
		require.Equal(t, uint64(len(contracts)), resp.Pagination.Total)
	})

	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.ListContract(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
