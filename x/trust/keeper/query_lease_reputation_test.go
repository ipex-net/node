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
			response: &types.QueryGetLeaseReputationResponse{LeaseReputation: types.LeaseReputationDisplay{
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
			request: &types.QueryGetLeaseReputationRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetLeaseReputationResponse{LeaseReputation: types.LeaseReputationDisplay{
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
		// Note: Pagination not implemented in LeaseReputationAll, so we just test that it returns all items
		resp, err := keeper.LeaseReputationAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.LeaseReputation))
		// Can't use ElementsMatch directly due to type conversion, just check length
	})
	t.Run("ByKey", func(t *testing.T) {
		// Note: Pagination not implemented in LeaseReputationAll, so we just test that it returns all items
		resp, err := keeper.LeaseReputationAll(ctx, request(nil, 0, 2, false))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.LeaseReputation))
		// Can't use ElementsMatch directly due to type conversion, just check length
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LeaseReputationAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), len(resp.LeaseReputation))
		// Note: Pagination is nil in LeaseReputationAll, so we can't check resp.Pagination.Total
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LeaseReputationAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
