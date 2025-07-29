package keeper_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"

	"github.com/stretchr/testify/require"
)

// Integration tests for reputation system covering complex scenarios

func TestReputationBothSidesOfContract(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(100)
	lessorId := uint64(200)
	ipCount := uint64(512)

	// Initialize both reputations
	keeper.InitializeLesseeReputation(ctx, lesseeId)
	keeper.InitializeLessorReputation(ctx, lessorId)

	// Both sides activate contract
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)
	err = keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Verify both sides updated correctly
	lesseeRep, _ := keeper.GetLesseeReputation(ctx, lesseeId)
	lessorRep, _ := keeper.GetLessorReputation(ctx, lessorId)

	require.Equal(t, uint64(1), lesseeRep.ContractsActive)
	require.Equal(t, uint64(1), lessorRep.ContractsActive)
	require.Equal(t, ipCount, lesseeRep.IpsRentedNow)
	require.Equal(t, ipCount, lessorRep.IpsLeasedNow)

	// Both sides renew
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount)
	require.NoError(t, err)
	err = keeper.LogLessorContractRenewal(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Verify IP months accumulated for both
	lesseeRep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	lessorRep, _ = keeper.GetLessorReputation(ctx, lessorId)
	require.Equal(t, ipCount*2, lesseeRep.IpMonthsTotal)
	require.Equal(t, ipCount*2, lessorRep.IpMonthsTotal)

	// Normal termination for both sides
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err)
	err = keeper.LogLessorContractTermination(ctx, lessorId, false, ipCount)
	require.NoError(t, err)

	// Verify final state
	lesseeRep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	lessorRep, _ = keeper.GetLessorReputation(ctx, lessorId)

	require.Equal(t, uint64(0), lesseeRep.ContractsActive)
	require.Equal(t, uint64(0), lessorRep.ContractsActive)
	require.Equal(t, uint64(1), lesseeRep.ContractsTerminated)
	require.Equal(t, uint64(1), lessorRep.ContractsTerminated)
	require.Equal(t, uint64(0), lesseeRep.ContractsTerminatedViolated)
	require.Equal(t, uint64(0), lessorRep.ContractsTerminatedViolated)
}

func TestReputationMultipleContractsPerUser(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(300)
	ipCount1 := uint64(256)
	ipCount2 := uint64(128)
	ipCount3 := uint64(512)

	// Initialize reputation
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Activate 3 different contracts
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount1)
	require.NoError(t, err)
	err = keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount2)
	require.NoError(t, err)
	err = keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount3)
	require.NoError(t, err)

	// Verify state after activations
	rep, _ := keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(3), rep.ContractsActive)
	require.Equal(t, ipCount1+ipCount2+ipCount3, rep.IpsRentedNow)
	require.Equal(t, ipCount1+ipCount2+ipCount3, rep.IpMonthsTotal)

	// Renew all contracts
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount1)
	require.NoError(t, err)
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount2)
	require.NoError(t, err)
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount3)
	require.NoError(t, err)

	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	expectedIpMonths := (ipCount1 + ipCount2 + ipCount3) * 2
	require.Equal(t, expectedIpMonths, rep.IpMonthsTotal)

	// Terminate with mixed violation results
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount1) // Normal
	require.NoError(t, err)
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, true, ipCount2) // Violation
	require.NoError(t, err)
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount3) // Normal
	require.NoError(t, err)

	// Verify final counts
	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(0), rep.ContractsActive)
	require.Equal(t, uint64(0), rep.IpsRentedNow)
	require.Equal(t, uint64(3), rep.ContractsTerminated)
	require.Equal(t, uint64(1), rep.ContractsTerminatedViolated, "Only one violation")
}

func TestReputationEdgeCasesContractCounts(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(400)
	ipCount := uint64(64)

	// Initialize reputation
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Try to terminate without any active contracts - current implementation allows this
	err := keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err, "Current implementation allows termination without active contracts")

	// Check state after terminating without active contracts
	rep, _ := keeper.GetLesseeReputation(ctx, lesseeId)
	// With underflow protection, ContractsActive should stay at 0
	require.Equal(t, uint64(0), rep.ContractsActive, "ContractsActive should not underflow")

	// Activate a contract  
	err = keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	// After activation, it should properly be 1
	require.Equal(t, uint64(1), rep.ContractsActive)

	// Terminate it
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(0), rep.ContractsActive)

	// Try to terminate again - current implementation allows this
	// Note: This depends on your business logic - you might want to prevent this or allow it
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err, "Current implementation allows multiple terminations")
}

func TestReputationIPCountAccuracy(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(500)

	// Test with various IP counts to ensure math is correct
	testCases := []struct {
		name    string
		ipCount uint64
	}{
		{"Small network /30", 4},
		{"Medium network /24", 256},
		{"Large network /16", 65536},
		{"Single IP /32", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use unique lessee ID for each test
			currentLesseeId := lesseeId + uint64(len(tc.name))
			
			keeper.InitializeLesseeReputation(ctx, currentLesseeId)

			// Activate
			err := keeper.LogLesseeContractActivation(ctx, currentLesseeId, tc.ipCount)
			require.NoError(t, err)

			rep, _ := keeper.GetLesseeReputation(ctx, currentLesseeId)
			require.Equal(t, tc.ipCount, rep.IpsRentedNow)
			require.Equal(t, tc.ipCount, rep.IpMonthsTotal)

			// Renew multiple times
			for i := 0; i < 5; i++ {
				err = keeper.LogLesseeContractRenewal(ctx, currentLesseeId, tc.ipCount)
				require.NoError(t, err)
			}

			rep, _ = keeper.GetLesseeReputation(ctx, currentLesseeId)
			require.Equal(t, tc.ipCount, rep.IpsRentedNow, "IPs rented should stay same")
			require.Equal(t, tc.ipCount*6, rep.IpMonthsTotal, "IP months should accumulate (1 activation + 5 renewals)")

			// Terminate
			err = keeper.LogLesseeContractTermination(ctx, currentLesseeId, false, tc.ipCount)
			require.NoError(t, err)

			rep, _ = keeper.GetLesseeReputation(ctx, currentLesseeId)
			require.Equal(t, uint64(0), rep.IpsRentedNow, "IPs rented should be 0")
			require.Equal(t, tc.ipCount*6, rep.IpMonthsTotal, "IP months should remain accumulated")
		})
	}
}

// Test that validates cross-validation between functions
func TestReputationFunctionParameterValidation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)

	// Test all functions with zero parameters
	testFunctions := []struct {
		name string
		testFunc func() error
	}{
		{
			"LogLesseeContractActivation zero lessee ID",
			func() error { return keeper.LogLesseeContractActivation(ctx, 0, 256) },
		},
		{
			"LogLesseeContractActivation zero IP count",
			func() error { 
				keeper.InitializeLesseeReputation(ctx, 999)
				return keeper.LogLesseeContractActivation(ctx, 999, 0) 
			},
		},
		{
			"LogLessorContractActivation zero lessor ID",
			func() error { return keeper.LogLessorContractActivation(ctx, 0, 256) },
		},
		{
			"LogLessorContractActivation zero IP count",
			func() error { 
				keeper.InitializeLessorReputation(ctx, 998)
				return keeper.LogLessorContractActivation(ctx, 998, 0) 
			},
		},
	}

	for _, tc := range testFunctions {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			require.Error(t, err, "Function should return error for invalid parameters")
		})
	}
}