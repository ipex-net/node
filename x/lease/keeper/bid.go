package keeper

import (
	"context"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"ipex/x/lease/types"
)

// updateBidIndexes updates all secondary indexes for a bid
func (k Keeper) updateBidIndexes(ctx context.Context, bid types.Bid) {
	bidKey := collections.Join(bid.ContractId, bid.BidderOrg)
	
	// Index by contract
	err := k.BidsByContract.Set(ctx, collections.Join(bid.ContractId, bidKey), true)
	if err != nil {
		panic(err)
	}

	// Index by bidder
	err = k.BidsByBidder.Set(ctx, collections.Join(bid.BidderOrg, bidKey), true)
	if err != nil {
		panic(err)
	}
}

// removeBidFromIndexes removes bid from all secondary indexes
func (k Keeper) removeBidFromIndexes(ctx context.Context, bid types.Bid) {
	bidKey := collections.Join(bid.ContractId, bid.BidderOrg)
	
	// Remove from contract index
	err := k.BidsByContract.Remove(ctx, collections.Join(bid.ContractId, bidKey))
	if err != nil {
		panic(err)
	}

	// Remove from bidder index
	err = k.BidsByBidder.Remove(ctx, collections.Join(bid.BidderOrg, bidKey))
	if err != nil {
		panic(err)
	}
}

// SetBid set a specific bid in the store
func (k Keeper) SetBid(ctx context.Context, bid types.Bid) {
	bidKey := collections.Join(bid.ContractId, bid.BidderOrg)
	
	// Get old bid to handle index updates
	oldBid, found := k.GetBid(ctx, bid.ContractId, bid.BidderOrg)
	if found {
		// Remove old indexes
		k.removeBidFromIndexes(ctx, oldBid)
	}

	// Store the bid
	err := k.Bids.Set(ctx, bidKey, bid)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateBidIndexes(ctx, bid)
}

// GetBid returns a bid from contract ID and bidder org
func (k Keeper) GetBid(
	ctx context.Context,
	contractId uint64,
	bidderOrg uint64,
) (val types.Bid, found bool) {
	bidKey := collections.Join(contractId, bidderOrg)
	bid, err := k.Bids.Get(ctx, bidKey)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return val, false
		}
		panic(err)
	}
	return bid, true
}

// RemoveBid removes a bid from the store
func (k Keeper) RemoveBid(
	ctx context.Context,
	contractId uint64,
	bidderOrg uint64,
) error {
	// Get the bid first to access its indexed fields
	bid, found := k.GetBid(ctx, contractId, bidderOrg)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "bid with contract ID %d and bidder org %d not found", contractId, bidderOrg)
	}

	bidKey := collections.Join(contractId, bidderOrg)
	
	// Remove from main collection
	err := k.Bids.Remove(ctx, bidKey)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeBidFromIndexes(ctx, bid)

	return nil
}

// GetAllBid returns all bid
func (k Keeper) GetAllBid(ctx context.Context) (list []types.Bid) {
	err := k.Bids.Walk(ctx, nil, func(key collections.Pair[uint64, uint64], value types.Bid) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// GetBidsByContract returns all bids for a specific contract
func (k Keeper) GetBidsByContract(ctx context.Context, contractId uint64) (list []types.Bid) {
	// Get all bids for this contract
	rng := collections.NewPrefixedPairRange[uint64, collections.Pair[uint64, uint64]](contractId)

	err := k.BidsByContract.Walk(ctx, rng, func(key collections.Pair[uint64, collections.Pair[uint64, uint64]], _ bool) (bool, error) {
		bidKey := key.K2()
		bid, found := k.GetBid(ctx, bidKey.K1(), bidKey.K2())
		if found {
			list = append(list, bid)
		}
		return false, nil
	})

	if err != nil {
		panic(err)
	}

	return
}

// GetBidsByBidder returns all bids for a specific bidder organization
func (k Keeper) GetBidsByBidder(ctx context.Context, bidderOrg uint64) (list []types.Bid) {
	// Get all bids for this bidder
	rng := collections.NewPrefixedPairRange[uint64, collections.Pair[uint64, uint64]](bidderOrg)

	err := k.BidsByBidder.Walk(ctx, rng, func(key collections.Pair[uint64, collections.Pair[uint64, uint64]], _ bool) (bool, error) {
		bidKey := key.K2()
		bid, found := k.GetBid(ctx, bidKey.K1(), bidKey.K2())
		if found {
			list = append(list, bid)
		}
		return false, nil
	})

	if err != nil {
		panic(err)
	}

	return
}

// RemoveAllBidsForContract removes all bids for a specific contract
func (k Keeper) RemoveAllBidsForContract(ctx context.Context, contractId uint64) error {
	bids := k.GetBidsByContract(ctx, contractId)
	for _, bid := range bids {
		err := k.RemoveBid(ctx, bid.ContractId, bid.BidderOrg)
		if err != nil {
			return err
		}
	}
	return nil
}

// // CancelExpiredBid cancels the bids that are expired 
// func (k Keeper) CancelExpiredBid(ctx context.Context) {
// 	//bids := k.GetAllBid(ctx cp)
// }