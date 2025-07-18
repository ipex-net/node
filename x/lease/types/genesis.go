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
	// Check for duplicated bid keys (contract-id + bidder-org combination)
	bidKeyMap := make(map[string]struct{})

	for _, elem := range gs.BidList {
		// Create key from contract ID and bidder org
		key := fmt.Sprintf("%d-%d", elem.ContractId, elem.BidderOrgId)
		if _, ok := bidKeyMap[key]; ok {
			return fmt.Errorf("duplicated bid for contract %d and bidder org %d", elem.ContractId, elem.BidderOrgId)
		}
		bidKeyMap[key] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
