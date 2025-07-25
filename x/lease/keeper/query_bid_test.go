package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/lease/keeper"
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

func createBidsForTesting(keeper keeper.Keeper, ctx context.Context) []types.Bid {
	bids := []types.Bid{
		// Contract 1 bids
		{ContractId: 1, BidderOrgId: 101, OwnerOrgId: 201},
		{ContractId: 1, BidderOrgId: 102, OwnerOrgId: 201},
		// Contract 2 bids
		{ContractId: 2, BidderOrgId: 103, OwnerOrgId: 202},
		{ContractId: 2, BidderOrgId: 104, OwnerOrgId: 202},
		// Contract 3 bids (same owner as contract 1)
		{ContractId: 3, BidderOrgId: 105, OwnerOrgId: 201},
		// Contract 4 bid (bidder 101 bidding on different contract)
		{ContractId: 4, BidderOrgId: 101, OwnerOrgId: 203},
	}

	for _, bid := range bids {
		keeper.SetBid(ctx, bid)
	}
	return bids
}

func TestBidByContractQuery(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	bids := createBidsForTesting(keeper, ctx)

	tests := []struct {
		desc         string
		request      *types.QueryBidByContractRequest
		expectedBids []types.Bid
		err          error
	}{
		{
			desc: "Contract 1 bids",
			request: &types.QueryBidByContractRequest{
				ContractId: 1,
			},
			expectedBids: []types.Bid{bids[0], bids[1]},
		},
		{
			desc: "Contract 2 bids",
			request: &types.QueryBidByContractRequest{
				ContractId: 2,
			},
			expectedBids: []types.Bid{bids[2], bids[3]},
		},
		{
			desc: "Contract with no bids",
			request: &types.QueryBidByContractRequest{
				ContractId: 999,
			},
			expectedBids: []types.Bid{},
		},
		{
			desc: "Zero contract ID",
			request: &types.QueryBidByContractRequest{
				ContractId: 0,
			},
			err: status.Error(codes.InvalidArgument, "contract ID cannot be zero"),
		},
		{
			desc: "Invalid request",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BidByContract(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t,
					nullify.Fill(tc.expectedBids),
					nullify.Fill(response.Bid),
				)
			}
		})
	}
}

func TestBidByBidderQuery(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	bids := createBidsForTesting(keeper, ctx)

	tests := []struct {
		desc         string
		request      *types.QueryBidByBidderRequest
		expectedBids []types.Bid
		err          error
	}{
		{
			desc: "Bidder 101 bids",
			request: &types.QueryBidByBidderRequest{
				BidderOrgId: 101,
			},
			expectedBids: []types.Bid{bids[0], bids[5]}, // bidder 101 has bids on contracts 1 and 4
		},
		{
			desc: "Bidder 102 bids",
			request: &types.QueryBidByBidderRequest{
				BidderOrgId: 102,
			},
			expectedBids: []types.Bid{bids[1]},
		},
		{
			desc: "Bidder with no bids",
			request: &types.QueryBidByBidderRequest{
				BidderOrgId: 999,
			},
			expectedBids: []types.Bid{},
		},
		{
			desc: "Zero bidder org ID",
			request: &types.QueryBidByBidderRequest{
				BidderOrgId: 0,
			},
			err: status.Error(codes.InvalidArgument, "bidder org ID cannot be zero"),
		},
		{
			desc: "Invalid request",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BidByBidder(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t,
					nullify.Fill(tc.expectedBids),
					nullify.Fill(response.Bid),
				)
			}
		})
	}
}

func TestBidByLessorQuery(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	bids := createBidsForTesting(keeper, ctx)

	tests := []struct {
		desc         string
		request      *types.QueryBidByLessorRequest
		expectedBids []types.Bid
		err          error
	}{
		{
			desc: "Lessor 201 bids",
			request: &types.QueryBidByLessorRequest{
				LessorOrgId: 201,
			},
			expectedBids: []types.Bid{bids[0], bids[1], bids[4]}, // lessor 201 owns contracts 1 and 3
		},
		{
			desc: "Lessor 202 bids",
			request: &types.QueryBidByLessorRequest{
				LessorOrgId: 202,
			},
			expectedBids: []types.Bid{bids[2], bids[3]}, // lessor 202 owns contract 2
		},
		{
			desc: "Lessor 203 bids",
			request: &types.QueryBidByLessorRequest{
				LessorOrgId: 203,
			},
			expectedBids: []types.Bid{bids[5]}, // lessor 203 owns contract 4
		},
		{
			desc: "Lessor with no bids",
			request: &types.QueryBidByLessorRequest{
				LessorOrgId: 999,
			},
			expectedBids: []types.Bid{},
		},
		{
			desc: "Zero lessor org ID",
			request: &types.QueryBidByLessorRequest{
				LessorOrgId: 0,
			},
			err: status.Error(codes.InvalidArgument, "lessor org ID cannot be zero"),
		},
		{
			desc: "Invalid request",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BidByLessor(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t,
					nullify.Fill(tc.expectedBids),
					nullify.Fill(response.Bid),
				)
			}
		})
	}
}
