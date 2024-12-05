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
	"ipex/x/org/types"
)

func TestOrgQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	msgs := createNOrg(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetOrgRequest
		response *types.QueryGetOrgResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetOrgRequest{Id: msgs[0].Id},
			response: &types.QueryGetOrgResponse{Org: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetOrgRequest{Id: msgs[1].Id},
			response: &types.QueryGetOrgResponse{Org: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetOrgRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Org(ctx, tc.request)
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

func TestOrgQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.OrgKeeper(t)
	msgs := createNOrg(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllOrgRequest {
		return &types.QueryAllOrgRequest{
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
			resp, err := keeper.OrgAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Org), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Org),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.OrgAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Org), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Org),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.OrgAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Org),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.OrgAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
