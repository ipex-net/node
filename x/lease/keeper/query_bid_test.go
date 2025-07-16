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
	"ipex/x/lease/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize


func TestBidQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	msgs := createNBid(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetBidRequest
		response *types.QueryGetBidResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetBidRequest{
				Index: strconv.FormatUint(msgs[0].Index, 10),
			},
			response: &types.QueryGetBidResponse{Bid: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetBidRequest{
				Index: strconv.FormatUint(msgs[1].Index, 10),
			},
			response: &types.QueryGetBidResponse{Bid: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetBidRequest{
				Index: "100000",
			},
			err: status.Error(codes.NotFound, "bid not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Bid(ctx, tc.request)
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

func TestBidQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	msgs := createNBid(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllBidRequest {
		return &types.QueryAllBidRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		resp, err := keeper.BidAll(ctx, request(nil, 0, 0, false))
		require.NoError(t, err)
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Bid),
		)
	})
	t.Run("ByKey", func(t *testing.T) {
		resp, err := keeper.BidAll(ctx, request(nil, 0, 0, false))
		require.NoError(t, err)
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Bid),
		)
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.BidAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Bid),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BidAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
