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

func TestSubnetReputationQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNSubnetReputation(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetSubnetReputationRequest
		response *types.QueryGetSubnetReputationResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetSubnetReputationRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetSubnetReputationResponse{SubnetReputation: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetSubnetReputationRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetSubnetReputationResponse{SubnetReputation: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetSubnetReputationRequest{
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
			response, err := keeper.SubnetReputation(ctx, tc.request)
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

func TestSubnetReputationQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	msgs := createNSubnetReputation(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllSubnetReputationRequest {
		return &types.QueryAllSubnetReputationRequest{
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
			resp, err := keeper.SubnetReputationAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SubnetReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.SubnetReputation),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.SubnetReputationAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.SubnetReputation), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.SubnetReputation),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.SubnetReputationAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.SubnetReputation),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.SubnetReputationAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
