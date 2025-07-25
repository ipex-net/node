package keeper

import (
	"context"

	"ipex/x/lease/types"

	"cosmossdk.io/collections"
)

func (k Keeper) GetContractsByStatus(ctx context.Context, status types.ContractStatus) ([]types.Contract, error) {
	var contracts []types.Contract

	// Query all contracts with this status using prefix iteration
	rng := collections.NewPrefixedPairRange[int32, uint64](int32(status))
	err := k.ContractsByStatus.Walk(ctx, rng, func(key collections.Pair[int32, uint64], _ bool) (bool, error) {
		// key.K2() is the contract ID
		contractId := key.K2()
		contract, found := k.GetContract(ctx, contractId)
		if found {
			contracts = append(contracts, contract)
		}
		return false, nil
	})

	if err != nil {
		return nil, err
	}

	return contracts, nil
}
