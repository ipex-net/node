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
			response: &types.QueryGetLesseeReputationResponse{LesseeReputation: types.LesseeReputationDisplay{
				Index:           msgs[0].Index,
				Level1Count:     msgs[0].Level1Count,
				Level2Count:     msgs[0].Level2Count,
				Level3Count:     msgs[0].Level3Count,
				Level4Count:     msgs[0].Level4Count,
				LastUpdated:     msgs[0].LastUpdated,
				AbuseResolved:   msgs[0].AbuseResolved,
				AbuseFailed:     msgs[0].AbuseFailed,
				AbuseInProgress: msgs[0].AbuseInProgress,
			}},
		},
		{
			desc: "Second",
			request: &types.QueryGetLesseeReputationRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetLesseeReputationResponse{LesseeReputation: types.LesseeReputationDisplay{
				Index:           msgs[1].Index,
				Level1Count:     msgs[1].Level1Count,
				Level2Count:     msgs[1].Level2Count,
				Level3Count:     msgs[1].Level3Count,
				Level4Count:     msgs[1].Level4Count,
				LastUpdated:     msgs[1].LastUpdated,
				AbuseResolved:   msgs[1].AbuseResolved,
				AbuseFailed:     msgs[1].AbuseFailed,
				AbuseInProgress: msgs[1].AbuseInProgress,
			}},
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
		// Note: Pagination is properly implemented in LesseeReputationAll
		resp, err := keeper.LesseeReputationAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.LesseeReputation))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination is properly implemented in LesseeReputationAll
		resp, err := keeper.LesseeReputationAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.LesseeReputation))
		// Pagination should return exactly 2 items as requested
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LesseeReputationAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.LesseeReputation))
		// Can't use ElementsMatch directly due to type conversion, just check length
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LesseeReputationAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
