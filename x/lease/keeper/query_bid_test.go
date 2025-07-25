package keeper_test

import (
	"strconv"
	"testing"

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
				ContractId:  msgs[0].ContractId,
				BidderOrgId: msgs[0].BidderOrgId,
			},
			response: &types.QueryGetBidResponse{Bid: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetBidRequest{
				ContractId:  msgs[1].ContractId,
				BidderOrgId: msgs[1].BidderOrgId,
			},
			response: &types.QueryGetBidResponse{Bid: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetBidRequest{
				ContractId:  100000,
				BidderOrgId: 100000,
			},
			err: status.Error(codes.NotFound, "bid not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			desc: "ZeroContractId",
			request: &types.QueryGetBidRequest{
				ContractId:  0,
				BidderOrgId: 1,
			},
			err: status.Error(codes.InvalidArgument, "contract ID cannot be zero"),
		},
		{
			desc: "ZeroBidderOrgId",
			request: &types.QueryGetBidRequest{
				ContractId:  1,
				BidderOrgId: 0,
			},
			err: status.Error(codes.InvalidArgument, "bidder org ID cannot be zero"),
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

