package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SubnetList: []Subnet{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated ID in subnet
	subnetIdMap := make(map[uint64]bool)
	subnetCount := gs.GetSubnetCount()
	for _, elem := range gs.SubnetList {
		if _, ok := subnetIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for subnet")
		}
		if elem.Id >= subnetCount {
			return fmt.Errorf("subnet id should be lower or equal than the last id")
		}
		subnetIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
