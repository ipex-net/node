package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) AppendContract(ctx context.Context, contract types.Contract) uint64 {
	count := k.GetContractCount(ctx)
	// Set the ID of the appended value (start from 1, not 0)
	contract.Id = count + 1
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	appendedValue := k.cdc.MustMarshal(&contract)
	store.Set(GetContractIDBytes(contract.Id), appendedValue)
	k.SetContractCount(ctx, count+1)
	return contract.Id
}

func (k Keeper) GetContractCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ContractCountKey)
	bz := store.Get(byteKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func GetContractIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

func (k Keeper) SetContractCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ContractCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

func (k Keeper) GetContract(ctx context.Context, id uint64) (val types.Contract, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	b := store.Get(GetContractIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) SetContract(ctx context.Context, contract types.Contract) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	b := k.cdc.MustMarshal(&contract)
	store.Set(GetContractIDBytes(contract.Id), b)
}

func (k Keeper) RemoveContract(ctx context.Context, id uint64) error {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	key := GetContractIDBytes(id)
	// Check if contract exists
	if !store.Has(key) {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "contract with id %d not found", id)
	}
	store.Delete(key)
	return nil
}

func (k Keeper) GetAllContract(ctx context.Context) (list []types.Contract) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var val types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetFirstLeasedContractBySubnet returns the first contract ID with LEASED status for the given subnet ID
// Returns the contract ID and true if found, or 0 and false if not found
func (k Keeper) GetFirstLeasedContractBySubnet(ctx context.Context, subnetId uint64) (uint64, bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var contract types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)

		// Check if this contract matches our criteria
		if contract.SubnetId == subnetId && contract.Status == types.ContractStatus_CONTRACT_STATUS_LEASED {
			return contract.Id, true
		}
	}

	return 0, false
}
