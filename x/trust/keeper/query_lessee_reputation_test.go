package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/trust/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestLesseeReputationQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNLesseeReputation(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetLesseeReputationRequest
		response *types.QueryGetLesseeReputationResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLesseeReputationRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetLesseeReputationResponse{LesseeReputation: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetLesseeReputationRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetLesseeReputationResponse{LesseeReputation: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLesseeReputationRequest{
				Index: 100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LesseeReputation(ctx, tc.request)
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

func TestLesseeReputationQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNLesseeReputation(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLesseeReputationRequest {
		return &types.QueryAllLesseeReputationRequest{
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
			resp, err := keeper.LesseeReputationAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LesseeReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LesseeReputation),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LesseeReputationAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LesseeReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LesseeReputation),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LesseeReputationAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.LesseeReputation),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LesseeReputationAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
