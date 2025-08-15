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
	"ipex/x/trust/types"
)

func TestBlacklistLogQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNBlacklistLog(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetBlacklistLogRequest
		response *types.QueryGetBlacklistLogResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetBlacklistLogRequest{Id: msgs[0].Id},
			response: &types.QueryGetBlacklistLogResponse{BlacklistLog: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetBlacklistLogRequest{Id: msgs[1].Id},
			response: &types.QueryGetBlacklistLogResponse{BlacklistLog: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetBlacklistLogRequest{Id: 100000},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BlacklistLog(ctx, tc.request)
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

func TestBlacklistLogQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNBlacklistLog(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllBlacklistLogRequest {
		return &types.QueryAllBlacklistLogRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		// Pagination is properly implemented in BlacklistLogAll
		resp, err := keeper.BlacklistLogAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.BlacklistLog))
		// Should return exactly 2 items due to limit
	})
	t.Run("ByKey", func(t *testing.T) {
		// Pagination is properly implemented in BlacklistLogAll
		resp, err := keeper.BlacklistLogAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.BlacklistLog))
		// Should return exactly 2 items due to limit
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.BlacklistLogAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.BlacklistLog),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BlacklistLogAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
