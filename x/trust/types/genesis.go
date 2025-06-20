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
		BlacklistLogList:     []BlacklistLog{},
		LeaseReputationList:  []LeaseReputation{},
		LesseeReputationList: []LesseeReputation{},
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
	// Check for duplicated ID in blacklistLog
	blacklistLogIdMap := make(map[uint64]bool)
	blacklistLogCount := gs.GetBlacklistLogCount()
	for _, elem := range gs.BlacklistLogList {
		if _, ok := blacklistLogIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for blacklistLog")
		}
		if elem.Id >= blacklistLogCount {
			return fmt.Errorf("blacklistLog id should be lower or equal than the last id")
		}
		blacklistLogIdMap[elem.Id] = true
	}
	// Check for duplicated index in leaseReputation
	leaseReputationIndexMap := make(map[string]struct{})

	for _, elem := range gs.LeaseReputationList {
		index := string(LeaseReputationKey(elem.Index))
		if _, ok := leaseReputationIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for leaseReputation")
		}
		leaseReputationIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in lesseeReputation
	lesseeReputationIndexMap := make(map[string]struct{})

	for _, elem := range gs.LesseeReputationList {
		index := string(LesseeReputationKey(elem.Index))
		if _, ok := lesseeReputationIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for lesseeReputation")
		}
		lesseeReputationIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
