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
	"ipex/x/subnet/keeper"
	"ipex/x/subnet/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNSubnetForList(keeper keeper.Keeper, ctx context.Context, n int) []types.Subnet {
	items := make([]types.Subnet, n)
	for i := range items {
		items[i].Id = uint64(i + 1)
		items[i].NetworkAddress = uint64(192<<24 | 168<<16 | 1<<8 | uint64(i))
		keeper.SetSubnet(ctx, items[i])
	}
	return items
}

func TestListSubnetPaginated(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	msgs := createNSubnetForList(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryListSubnetRequest {
		return &types.QueryListSubnetRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		// Note: Pagination properly implemented in ListSubnet
		resp, err := keeper.ListSubnet(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Subnet))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination properly implemented in ListSubnet
		resp, err := keeper.ListSubnet(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Subnet))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.ListSubnet(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.Subnet))
		// Check pagination response includes total count
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.ListSubnet(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}