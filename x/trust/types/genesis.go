package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SubnetReputationList: []SubnetReputation{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in subnetReputation
	subnetReputationIndexMap := make(map[string]struct{})

	for _, elem := range gs.SubnetReputationList {
		index := string(SubnetReputationKey(elem.Index))
		if _, ok := subnetReputationIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for subnetReputation")
		}
		subnetReputationIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
