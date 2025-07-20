package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/lease/keeper"
	"ipex/x/lease/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNBid(keeper keeper.Keeper, ctx context.Context, n int) []types.Bid {
	items := make([]types.Bid, n)
	for i := range items {
		items[i].ContractId = uint64(i + 1)
		items[i].BidderOrgId = uint64(i + 100)

		keeper.SetBid(ctx, items[i])
	}
	return items
}

func TestBidGet(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBid(ctx,
			item.ContractId,
			item.BidderOrgId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestBidRemove(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	for _, item := range items {
		err := keeper.RemoveBid(ctx,
			item.ContractId,
			item.BidderOrgId,
		)
		require.NoError(t, err)
		_, found := keeper.GetBid(ctx,
			item.ContractId,
			item.BidderOrgId,
		)
		require.False(t, found)
	}
}

func TestBidGetAll(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	items := createNBid(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBids(ctx)),
	)
}

func TestBidGetByContract(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)

	// Create bids with specific contract IDs
	contractId1 := uint64(100)
	contractId2 := uint64(200)

	bid1 := types.Bid{ContractId: contractId1, BidderOrgId: 1}
	bid2 := types.Bid{ContractId: contractId1, BidderOrgId: 2}
	bid3 := types.Bid{ContractId: contractId2, BidderOrgId: 1}

	keeper.SetBid(ctx, bid1)
	keeper.SetBid(ctx, bid2)
	keeper.SetBid(ctx, bid3)

	// Get bids for contract 1
	bidsForContract1 := keeper.GetBidsByContract(ctx, contractId1)
	require.Len(t, bidsForContract1, 2)

	// Get bids for contract 2
	bidsForContract2 := keeper.GetBidsByContract(ctx, contractId2)
	require.Len(t, bidsForContract2, 1)

	// Get bids for non-existent contract
	bidsForContract3 := keeper.GetBidsByContract(ctx, 999)
	require.Len(t, bidsForContract3, 0)
}

func TestBidGetByBidder(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)

	// Create bids with specific bidder orgs
	bidderOrg1 := uint64(100)
	bidderOrg2 := uint64(200)

	bid1 := types.Bid{ContractId: 1, BidderOrgId: bidderOrg1}
	bid2 := types.Bid{ContractId: 2, BidderOrgId: bidderOrg1}
	bid3 := types.Bid{ContractId: 1, BidderOrgId: bidderOrg2}

	keeper.SetBid(ctx, bid1)
	keeper.SetBid(ctx, bid2)
	keeper.SetBid(ctx, bid3)

	// Get bids for bidder 1
	bidsForBidder1 := keeper.GetBidsByBidder(ctx, bidderOrg1)
	require.Len(t, bidsForBidder1, 2)

	// Get bids for bidder 2
	bidsForBidder2 := keeper.GetBidsByBidder(ctx, bidderOrg2)
	require.Len(t, bidsForBidder2, 1)

	// Get bids for non-existent bidder
	bidsForBidder3 := keeper.GetBidsByBidder(ctx, 999)
	require.Len(t, bidsForBidder3, 0)
}

func TestRemoveAllBidsForContract(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)

	contractId := uint64(100)

	// Create multiple bids for the same contract
	bid1 := types.Bid{ContractId: contractId, BidderOrgId: 1}
	bid2 := types.Bid{ContractId: contractId, BidderOrgId: 2}
	bid3 := types.Bid{ContractId: contractId, BidderOrgId: 3}
	bid4 := types.Bid{ContractId: 200, BidderOrgId: 1} // Different contract

	keeper.SetBid(ctx, bid1)
	keeper.SetBid(ctx, bid2)
	keeper.SetBid(ctx, bid3)
	keeper.SetBid(ctx, bid4)

	// Verify bids exist
	bidsForContract := keeper.GetBidsByContract(ctx, contractId)
	require.Len(t, bidsForContract, 3)

	// Remove all bids for the contract
	err := keeper.RemoveAllBidsForContract(ctx, contractId)
	require.NoError(t, err)

	// Verify bids are removed
	bidsForContract = keeper.GetBidsByContract(ctx, contractId)
	require.Len(t, bidsForContract, 0)

	// Verify bid for different contract still exists
	_, found := keeper.GetBid(ctx, 200, 1)
	require.True(t, found)
}
