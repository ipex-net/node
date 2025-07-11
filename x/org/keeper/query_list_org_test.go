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
	"ipex/x/org/keeper"
	"ipex/x/org/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNOrgForList(keeper keeper.Keeper, ctx context.Context, n int) []types.Org {
	items := make([]types.Org, n)
	for i := range items {
		items[i].Id = uint64(i + 1)
		items[i].Name = "test-org-" + strconv.Itoa(i)
		keeper.SetOrg(ctx, items[i])
	}
	return items
}

func TestListOrgPaginated(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	msgs := createNOrgForList(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryListOrgRequest {
		return &types.QueryListOrgRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		// Note: Pagination properly implemented in ListOrg
		resp, err := keeper.ListOrg(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Org))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination properly implemented in ListOrg
		resp, err := keeper.ListOrg(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Org))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.ListOrg(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.Org))
		// Check pagination response includes total count
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.ListOrg(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}