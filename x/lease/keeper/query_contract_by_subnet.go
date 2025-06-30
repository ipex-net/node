package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"ipex/x/lease/types"
)

func (k Keeper) GetContractsBySubnet(ctx context.Context, subnetId uint64) []types.Contract {
	var contracts []types.Contract

	// Query all contracts for this subnet using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](subnetId)
	err := k.ContractsBySubnet.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], _ bool) (bool, error) {
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