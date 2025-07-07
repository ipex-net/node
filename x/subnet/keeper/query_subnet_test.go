package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/subnet/types"
)

func TestSubnetQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	msgs := createNSubnet(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetSubnetRequest
		response *types.QueryGetSubnetResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetSubnetRequest{Id: msgs[0].Id},
			response: &types.QueryGetSubnetResponse{Subnet: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetSubnetRequest{Id: msgs[1].Id},
			response: &types.QueryGetSubnetResponse{Subnet: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetSubnetRequest{Id: 100000},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Subnet(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestSubnetQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	msgs := createNSubnet(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllSubnetRequest {
		return &types.QueryAllSubnetRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.SubnetAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Subnet), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Subnet),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.SubnetAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Subnet), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Subnet),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SubnetAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Subnet),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SubnetAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
