package keeper

import (
	"context"
	"fmt"

	"ipex/x/trust/types"
)

func (k Keeper) SetLessorReputation(ctx context.Context, lessorReputation types.LessorReputation) {
	k.LessorReputations.Set(ctx, lessorReputation.Index, lessorReputation)
}

func (k Keeper) GetLessorReputation(ctx context.Context, index uint64) (val types.LessorReputation, found bool) {
	val, err := k.LessorReputations.Get(ctx, index)
	if err != nil {
		return val, false
	}
	return val, true
}

func (k Keeper) RemoveLessorReputation(ctx context.Context, index uint64) {
	k.LessorReputations.Remove(ctx, index)
}

func (k Keeper) GetAllLessorReputation(ctx context.Context) (list []types.LessorReputation) {
	err := k.LessorReputations.Walk(ctx, nil, func(key uint64, value types.LessorReputation) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitializeLessorReputation initializes a new lessor reputation entry when org is created
func (k Keeper) InitializeLessorReputation(ctx context.Context, lessorId uint64) {
	// Create new lessor reputation with initial values
	lessorReputation := types.LessorReputation{
		Index:                       lessorId,
		IpMonthsTotal:               0,
		IpsLeasedNow:                0,
		ContractsActive:             0,
		ContractsTerminated:         0,
		ContractsTerminatedViolated: 0,
	}

	// Set the lessor reputation in the store
	k.SetLessorReputation(ctx, lessorReputation)
}

// LogLessorContractTermination updates usage stats on contract termination
func (k Keeper) LogLessorContractTermination(ctx context.Context, lessorId uint64, lessorContractViolation bool, ipCount uint64) error {
	if lessorId == 0 {
		return fmt.Errorf("lessor id can't be nil")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lessorRep, found := k.GetLessorReputation(ctx, lessorId)
	if !found {
		return fmt.Errorf("can not find lessor")
	}
	lessorRep.ContractsTerminated++
	if lessorContractViolation {
		lessorRep.ContractsTerminatedViolated++
	}
	// Prevent underflow for ContractsActive
	if lessorRep.ContractsActive > 0 {
		lessorRep.ContractsActive--
	}
	// Prevent underflow for IpsLeasedNow
	if lessorRep.IpsLeasedNow >= ipCount {
		lessorRep.IpsLeasedNow = lessorRep.IpsLeasedNow - ipCount
	} else {
		lessorRep.IpsLeasedNow = 0
	}
	k.SetLessorReputation(ctx, lessorRep)
	return nil
}

// LogLessorContractRenewal updates usage stats on contract renewal
func (k Keeper) LogLessorContractRenewal(ctx context.Context, lessorId uint64, ipCount uint64) error {
	if lessorId == 0 {
		return fmt.Errorf("lessor id can't be nil")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lessorRep, found := k.GetLessorReputation(ctx, lessorId)
	if !found {
		return fmt.Errorf("can not find lessor")
	}
	lessorRep.IpMonthsTotal = lessorRep.IpMonthsTotal + ipCount
	k.SetLessorReputation(ctx, lessorRep)
	return nil
}

// LogLessorContractActivation updates usage stats on contract activation
func (k Keeper) LogLessorContractActivation(ctx context.Context, lessorId uint64, ipCount uint64) error {
	if lessorId == 0 {
		return fmt.Errorf("lessor id can not be nil")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lessorRep, found := k.GetLessorReputation(ctx, lessorId)
	if !found {
		return fmt.Errorf("can not find lessor")
	}
	lessorRep.IpMonthsTotal = lessorRep.IpMonthsTotal + ipCount
	lessorRep.ContractsActive++
	lessorRep.IpsLeasedNow = lessorRep.IpsLeasedNow + ipCount
	k.SetLessorReputation(ctx, lessorRep)
	return nil
}

// LessorHasActiveContracts returns true if lessor has any active contracts
func (k Keeper) LessorHasActiveContracts(ctx context.Context, lessorId uint64) (bool, error) {
	if lessorId == 0 {
		return false, fmt.Errorf("lessor id can not be nil")
	}
	lessorRep, found := k.GetLessorReputation(ctx, lessorId)
	if !found {
		return false, fmt.Errorf("can not find lessor")
	}
	if lessorRep.ContractsActive > 0 {
		return true, nil
	}
	return false, nil
}