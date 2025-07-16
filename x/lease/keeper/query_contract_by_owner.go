package keeper

import (
	"context"

	"ipex/x/lease/types"

	"cosmossdk.io/collections"
)

func (k Keeper) GetContractsByOwner(ctx context.Context, ownerOrgId uint64) []types.Contract {
	var contracts []types.Contract

	// Query all contracts for this owner org using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](ownerOrgId)
	err := k.ContractsByOwner.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], _ bool) (bool, error) {
		// key.K2() is the contract ID
		contractId := key.K2()
		contract, found := k.GetContract(ctx, contractId)
		if found {
			contracts = append(contracts, contract)
		}
		return false, nil
	})

	if err != nil {
		panic(err)
	}

	return contracts
}
