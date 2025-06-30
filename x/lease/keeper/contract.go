package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"ipex/x/lease/types"
)

// updateContractIndexes updates all secondary indexes for a contract
func (k Keeper) updateContractIndexes(ctx context.Context, contract types.Contract) {
	// Index by status
	err := k.ContractsByStatus.Set(ctx, collections.Join(int32(contract.Status), contract.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by subnet
	err = k.ContractsBySubnet.Set(ctx, collections.Join(contract.SubnetId, contract.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by owner
	err = k.ContractsByOwner.Set(ctx, collections.Join(contract.OwnerOrg, contract.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by lessee (only if lessee is set)
	if contract.LesseeOrg != 0 {
		err = k.ContractsByLessee.Set(ctx, collections.Join(contract.LesseeOrg, contract.Id), true)
		if err != nil {
			panic(err)
		}
	}
}

// removeContractFromIndexes removes contract from all secondary indexes
func (k Keeper) removeContractFromIndexes(ctx context.Context, contract types.Contract) {
	// Remove from status index
	err := k.ContractsByStatus.Remove(ctx, collections.Join(int32(contract.Status), contract.Id))
	if err != nil {
		panic(err)
	}

	// Remove from subnet index
	err = k.ContractsBySubnet.Remove(ctx, collections.Join(contract.SubnetId, contract.Id))
	if err != nil {
		panic(err)
	}

	// Remove from owner index
	err = k.ContractsByOwner.Remove(ctx, collections.Join(contract.OwnerOrg, contract.Id))
	if err != nil {
		panic(err)
	}

	// Remove from lessee index (only if lessee was set)
	if contract.LesseeOrg != 0 {
		err = k.ContractsByLessee.Remove(ctx, collections.Join(contract.LesseeOrg, contract.Id))
		if err != nil {
			panic(err)
		}
	}
}

func (k Keeper) AppendContract(ctx context.Context, contract types.Contract) uint64 {
	// Get next ID from sequence (add 1 to start from ID 1 instead of 0)
	id, err := k.ContractSequence.Next(ctx)
	if err != nil {
		panic(err)
	}
	contract.Id = id + 1

	// Store the contract
	err = k.Contracts.Set(ctx, contract.Id, contract)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateContractIndexes(ctx, contract)

	return contract.Id
}


func (k Keeper) GetContract(ctx context.Context, id uint64) (val types.Contract, found bool) {
	contract, err := k.Contracts.Get(ctx, id)
	if err != nil {
		if err.Error() == "collections: not found" {
			return val, false
		}
		panic(err)
	}
	return contract, true
}

func (k Keeper) SetContract(ctx context.Context, contract types.Contract) {
	// Get old contract to handle index updates
	oldContract, found := k.GetContract(ctx, contract.Id)
	if found {
		// Remove old indexes
		k.removeContractFromIndexes(ctx, oldContract)
	}

	// Store the contract
	err := k.Contracts.Set(ctx, contract.Id, contract)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateContractIndexes(ctx, contract)
}

func (k Keeper) RemoveContract(ctx context.Context, id uint64) error {
	// Get the contract first to access its indexed fields
	contract, found := k.GetContract(ctx, id)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "contract with id %d not found", id)
	}

	// Remove from main collection
	err := k.Contracts.Remove(ctx, id)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeContractFromIndexes(ctx, contract)

	return nil
}

func (k Keeper) GetAllContract(ctx context.Context) (list []types.Contract) {
	err := k.Contracts.Walk(ctx, nil, func(key uint64, value types.Contract) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// GetFirstLeasedContractBySubnet returns the first contract ID with LEASED status for the given subnet ID
// Returns the contract ID and true if found, or 0 and false if not found
func (k Keeper) GetFirstLeasedContractBySubnet(ctx context.Context, subnetId uint64) (uint64, bool) {
	// Get all contracts for this subnet
	rng := collections.NewPrefixedPairRange[uint64, uint64](subnetId)
	var contractId uint64
	found := false
	
	err := k.ContractsBySubnet.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], _ bool) (bool, error) {
		id := key.K2()
		contract, exists := k.GetContract(ctx, id)
		if exists && contract.Status == types.ContractStatus_CONTRACT_STATUS_LEASED {
			contractId = id
			found = true
			return true, nil // stop iteration
		}
		return false, nil // continue
	})
	
	if err != nil {
		panic(err)
	}
	
	return contractId, found
}
