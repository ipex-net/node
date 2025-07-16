package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		BidList: []Bid{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in bid
	bidIndexMap := make(map[string]struct{})

	for _, elem := range gs.BidList {
		index := string(BidKey(elem.Index))
		if _, ok := bidIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for bid")
		}
		bidIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
