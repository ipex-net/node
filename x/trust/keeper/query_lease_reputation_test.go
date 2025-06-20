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

func TestLeaseReputationQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNLeaseReputation(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetLeaseReputationRequest
		response *types.QueryGetLeaseReputationResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetLeaseReputationRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetLeaseReputationResponse{LeaseReputation: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetLeaseReputationRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetLeaseReputationResponse{LeaseReputation: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetLeaseReputationRequest{
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
			response, err := keeper.LeaseReputation(ctx, tc.request)
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

func TestLeaseReputationQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNLeaseReputation(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLeaseReputationRequest {
		return &types.QueryAllLeaseReputationRequest{
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
			resp, err := keeper.LeaseReputationAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LeaseReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LeaseReputation),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LeaseReputationAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LeaseReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LeaseReputation),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LeaseReputationAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.LeaseReputation),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LeaseReputationAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
