package keeper

import (
	"context"
	"fmt"

	"ipex/x/trust/types"
)

func (k Keeper) SetLesseeReputation(ctx context.Context, lesseeReputation types.LesseeReputation) {
	k.LesseeReputations.Set(ctx, lesseeReputation.Index, lesseeReputation)
}

func (k Keeper) GetLesseeReputation(ctx context.Context, index uint64) (val types.LesseeReputation, found bool) {
	val, err := k.LesseeReputations.Get(ctx, index)
	if err != nil {
		return val, false
	}
	return val, true
}

func (k Keeper) RemoveLesseeReputation(ctx context.Context, index uint64) {
	k.LesseeReputations.Remove(ctx, index)
}

func (k Keeper) GetAllLesseeReputation(ctx context.Context) (list []types.LesseeReputation) {
	err := k.LesseeReputations.Walk(ctx, nil, func(key uint64, value types.LesseeReputation) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// InitializeLesseeReputation initializes a new lessee reputation entry when org is created
func (k Keeper) InitializeLesseeReputation(ctx context.Context, lesseeId uint64) {

	// Create new lesee reputation with initial values
	lesseeReputation := types.LesseeReputation{
		Index:       lesseeId,
		Level1Count: 0, // Initialize all counts to zero
		Level2Count: 0,
		Level3Count: 0,
		Level4Count: 0,
		LastUpdated: nil,
		IpMonthsTotal: 0,
		IpsRentedNow: 0,
		ContractsActive: 0,
		ContractsTerminated: 0,
		ContractsTerminatedViolated: 0,
	}

	// Set the subnet reputation in the store
	k.SetLesseeReputation(ctx, lesseeReputation)

}

// LogLesseeContractTermination updates usage stats on contract termination
func (k Keeper) LogLesseeContractTermination(ctx context.Context, lesseeId uint64, violated bool, subnetMask uint64) error {
	ipCount := IPCountFromMask(subnetMask)
	if lesseeId == 0 {
		return fmt.Errorf("lessee id can't be nil")
	}
	if subnetMask > 32 || subnetMask < 8 {
		return fmt.Errorf("incorrect subnet mask")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lesseeRep, found := k.GetLesseeReputation(ctx, lesseeId)
	if !found {
		return fmt.Errorf("can not find lessee")
	}
	lesseeRep.ContractsTerminated++
	if violated {
		lesseeRep.ContractsTerminatedViolated++
	}
	lesseeRep.ContractsActive--
	lesseeRep.IpsRentedNow = lesseeRep.IpsRentedNow - ipCount
	k.SetLesseeReputation(ctx, lesseeRep)
	return nil
}

// LogLesseeContractRenewal updates usage stats on contract renewal
func (k Keeper) LogLesseeContractRenewal(ctx context.Context, lesseeId uint64, subnetMask uint64) error {
	ipCount := IPCountFromMask(subnetMask)
	if lesseeId == 0 {
		return fmt.Errorf("lessee id can't be nil")
	}
	if subnetMask > 32 || subnetMask < 8 {
		return fmt.Errorf("incorrect subnet mask")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lesseeRep, found := k.GetLesseeReputation(ctx, lesseeId)
	if !found {
		return fmt.Errorf("can not find lessee")
	}
	lesseeRep.IpMonthsTotal = lesseeRep.IpMonthsTotal + ipCount
	k.SetLesseeReputation(ctx, lesseeRep)
	return nil
}

// LogLesseeContractActivation updates usage stats on contract activation
func (k Keeper) LogLesseeContractActivation(ctx context.Context, lesseeId uint64, subnetMask uint64) error {
	ipCount := IPCountFromMask(subnetMask)
	if lesseeId == 0 {
		return fmt.Errorf("lessee id can't be nil")
	}
	if subnetMask > 32 || subnetMask < 8 {
		return fmt.Errorf("incorrect subnet mask")
	}
	if ipCount == 0 {
		return fmt.Errorf("incorrect ip count")
	}
	lesseeRep, found := k.GetLesseeReputation(ctx, lesseeId)
	if !found {
		return fmt.Errorf("can not find lessee")
	}
	lesseeRep.IpMonthsTotal = lesseeRep.IpMonthsTotal + ipCount
	lesseeRep.ContractsActive++
	lesseeRep.IpsRentedNow = lesseeRep.IpsRentedNow + ipCount
	k.SetLesseeReputation(ctx, lesseeRep)
	return nil
}

// IPCountFromMask returns the count of IPs per mask
func IPCountFromMask(mask uint64) uint64 {
   if mask > 32 {
       return 0
   }
   
   hostBits := 32 - mask
   return 1 << hostBits 
}