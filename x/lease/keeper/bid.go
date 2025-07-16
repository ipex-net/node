package keeper

import (
	"context"

	"ipex/x/lease/types"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetBid set a specific bid in the store from its index
func (k Keeper) SetBid(ctx context.Context, bid types.Bid) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BidKeyPrefix))
	b := k.cdc.MustMarshal(&bid)
	store.Set(types.BidKey(
		bid.Index,
	), b)
}

// GetBid returns a bid from its index
func (k Keeper) GetBid(
	ctx context.Context,
	index string,

) (val types.Bid, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BidKeyPrefix))

	b := store.Get(types.BidKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBid removes a bid from the store
func (k Keeper) RemoveBid(
	ctx context.Context,
	index string,

) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BidKeyPrefix))
	store.Delete(types.BidKey(
		index,
	))
}

// GetAllBid returns all bid
func (k Keeper) GetAllBid(ctx context.Context) (list []types.Bid) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.BidKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Bid
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
