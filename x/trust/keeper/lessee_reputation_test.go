package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/trust/keeper"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNLesseeReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.LesseeReputation {
	items := make([]types.LesseeReputation, n)
	for i := range items {
		items[i].Index = uint64(i + 1) // Start from 1 instead of 0

		keeper.SetLesseeReputation(ctx, items[i])
	}
	return items
}

func TestLesseeReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLesseeReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLesseeReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLesseeReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetLesseeReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestLesseeReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLesseeReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLesseeReputation(ctx)),
	)
}

// Business Logic Tests for Lessee Reputation

func TestInitializeLesseeReputation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(123)

	// Initialize reputation
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Verify it was created with correct initial values
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, lesseeId, rep.Index)
	require.Equal(t, uint64(0), rep.IpMonthsTotal)
	require.Equal(t, uint64(0), rep.IpsRentedNow)
	require.Equal(t, uint64(0), rep.ContractsActive)
	require.Equal(t, uint64(0), rep.ContractsTerminated)
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated)
}

func TestLogLesseeContractActivation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(100)
	ipCount := uint64(256)

	// Initialize reputation first
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Log contract activation
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, ipCount, rep.IpMonthsTotal, "IP months should be added")
	require.Equal(t, ipCount, rep.IpsRentedNow, "IPs rented now should be set")
	require.Equal(t, uint64(1), rep.ContractsActive, "Active contracts should increment")
	require.Equal(t, uint64(0), rep.ContractsTerminated, "Terminated contracts should remain 0")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated contracts should remain 0")
}

func TestLogLesseeContractActivationMultiple(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(100)
	ipCount1 := uint64(256)
	ipCount2 := uint64(512)

	// Initialize reputation
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Log first contract activation
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount1)
	require.NoError(t, err)

	// Log second contract activation
	err = keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount2)
	require.NoError(t, err)

	// Verify cumulative updates
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, ipCount1+ipCount2, rep.IpMonthsTotal, "IP months should accumulate")
	require.Equal(t, ipCount1+ipCount2, rep.IpsRentedNow, "IPs rented now should accumulate")
	require.Equal(t, uint64(2), rep.ContractsActive, "Active contracts should be 2")
}

func TestLogLesseeContractRenewal(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(200)
	ipCount := uint64(128)

	// Initialize and activate a contract first
	keeper.InitializeLesseeReputation(ctx, lesseeId)
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	// Get initial state
	repBefore, _ := keeper.GetLesseeReputation(ctx, lesseeId)
	initialIpMonths := repBefore.IpMonthsTotal

	// Log contract renewal
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	// Verify only IP months increased
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, initialIpMonths+ipCount, rep.IpMonthsTotal, "IP months should increase")
	require.Equal(t, ipCount, rep.IpsRentedNow, "IPs rented now should stay same")
	require.Equal(t, uint64(1), rep.ContractsActive, "Active contracts should stay same")
	require.Equal(t, uint64(0), rep.ContractsTerminated, "Terminated contracts should stay same")
}

func TestLogLesseeContractTerminationNormal(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(300)
	ipCount := uint64(64)

	// Initialize and activate a contract first
	keeper.InitializeLesseeReputation(ctx, lesseeId)
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	// Log normal termination (no violation)
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, ipCount, rep.IpMonthsTotal, "IP months should stay same")
	require.Equal(t, uint64(0), rep.IpsRentedNow, "IPs rented should decrease")
	require.Equal(t, uint64(0), rep.ContractsActive, "Active contracts should decrease")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated contracts should increment")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated should stay 0")
}

func TestLogLesseeContractTerminationWithViolation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(400)
	ipCount := uint64(32)

	// Initialize and activate a contract first
	keeper.InitializeLesseeReputation(ctx, lesseeId)
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	// Log termination with violation
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, true, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLesseeReputation(ctx, lesseeId)
	require.True(t, found)
	require.Equal(t, uint64(0), rep.ContractsActive, "Active contracts should decrease")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated contracts should increment")
	require.Equal(t, uint64(1), rep.ContractsTerminatedViolated, "Violated should increment")
}

// Edge Cases and Validation Tests

func TestLogLesseeContractActivationZeroLesseeId(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)

	err := keeper.LogLesseeContractActivation(ctx, 0, 256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lessee id can't be nil")
}

func TestLogLesseeContractActivationZeroIpCount(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(500)

	keeper.InitializeLesseeReputation(ctx, lesseeId)

	err := keeper.LogLesseeContractActivation(ctx, lesseeId, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "incorrect ip count")
}

func TestLogLesseeContractActivationMissingReputation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(600)

	// Don't initialize reputation first
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, 256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can not find lessee")
}

func TestLogLesseeReputationFullLifecycle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lesseeId := uint64(700)
	ipCount := uint64(1024)

	// Initialize
	keeper.InitializeLesseeReputation(ctx, lesseeId)

	// Activate contract
	err := keeper.LogLesseeContractActivation(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	rep, _ := keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(1), rep.ContractsActive)
	require.Equal(t, ipCount, rep.IpsRentedNow)
	require.Equal(t, ipCount, rep.IpMonthsTotal)

	// Renew contract twice
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount)
	require.NoError(t, err)
	err = keeper.LogLesseeContractRenewal(ctx, lesseeId, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(1), rep.ContractsActive, "Active should stay 1")
	require.Equal(t, ipCount, rep.IpsRentedNow, "IPs rented should stay same")
	require.Equal(t, ipCount*3, rep.IpMonthsTotal, "IP months should accumulate")

	// Terminate contract normally
	err = keeper.LogLesseeContractTermination(ctx, lesseeId, false, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLesseeReputation(ctx, lesseeId)
	require.Equal(t, uint64(0), rep.ContractsActive, "Active should be 0")
	require.Equal(t, uint64(0), rep.IpsRentedNow, "IPs rented should be 0")
	require.Equal(t, ipCount*3, rep.IpMonthsTotal, "IP months should stay accumulated")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated should be 1")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated should be 0")
}
