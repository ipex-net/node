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

func createNLessorReputation(keeper keeper.Keeper, ctx context.Context, n int) []types.LessorReputation {
	items := make([]types.LessorReputation, n)
	for i := range items {
		items[i].Index = uint64(i)

		keeper.SetLessorReputation(ctx, items[i])
	}
	return items
}

func TestLessorReputationGet(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLessorReputation(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestLessorReputationRemove(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveLessorReputation(ctx,
			item.Index,
		)
		_, found := keeper.GetLessorReputation(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestLessorReputationGetAll(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	items := createNLessorReputation(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllLessorReputation(ctx)),
	)
}

// Business Logic Tests for Lessor Reputation

func TestInitializeLessorReputation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(123)

	// Initialize reputation
	keeper.InitializeLessorReputation(ctx, lessorId)

	// Verify it was created with correct initial values
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, lessorId, rep.Index)
	require.Equal(t, uint64(0), rep.IpMonthsTotal)
	require.Equal(t, uint64(0), rep.IpsLeasedNow)
	require.Equal(t, uint64(0), rep.ContractsActive)
	require.Equal(t, uint64(0), rep.ContractsTerminated)
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated)
}

func TestLogLessorContractActivation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(100)
	ipCount := uint64(256)

	// Initialize reputation first
	keeper.InitializeLessorReputation(ctx, lessorId)

	// Log contract activation
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, ipCount, rep.IpMonthsTotal, "IP months should be added")
	require.Equal(t, ipCount, rep.IpsLeasedNow, "IPs leased now should be set")
	require.Equal(t, uint64(1), rep.ContractsActive, "Active contracts should increment")
	require.Equal(t, uint64(0), rep.ContractsTerminated, "Terminated contracts should remain 0")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated contracts should remain 0")
}

func TestLogLessorContractActivationMultiple(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(100)
	ipCount1 := uint64(256)
	ipCount2 := uint64(512)

	// Initialize reputation
	keeper.InitializeLessorReputation(ctx, lessorId)

	// Log first contract activation
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount1)
	require.NoError(t, err)

	// Log second contract activation
	err = keeper.LogLessorContractActivation(ctx, lessorId, ipCount2)
	require.NoError(t, err)

	// Verify cumulative updates
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, ipCount1+ipCount2, rep.IpMonthsTotal, "IP months should accumulate")
	require.Equal(t, ipCount1+ipCount2, rep.IpsLeasedNow, "IPs leased now should accumulate")
	require.Equal(t, uint64(2), rep.ContractsActive, "Active contracts should be 2")
}

func TestLogLessorContractRenewal(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(200)
	ipCount := uint64(128)

	// Initialize and activate a contract first
	keeper.InitializeLessorReputation(ctx, lessorId)
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Get initial state
	repBefore, _ := keeper.GetLessorReputation(ctx, lessorId)
	initialIpMonths := repBefore.IpMonthsTotal

	// Log contract renewal
	err = keeper.LogLessorContractRenewal(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Verify only IP months increased
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, initialIpMonths+ipCount, rep.IpMonthsTotal, "IP months should increase")
	require.Equal(t, ipCount, rep.IpsLeasedNow, "IPs leased now should stay same")
	require.Equal(t, uint64(1), rep.ContractsActive, "Active contracts should stay same")
	require.Equal(t, uint64(0), rep.ContractsTerminated, "Terminated contracts should stay same")
}

func TestLogLessorContractTerminationNormal(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(300)
	ipCount := uint64(64)

	// Initialize and activate a contract first
	keeper.InitializeLessorReputation(ctx, lessorId)
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Log normal termination (no violation)
	err = keeper.LogLessorContractTermination(ctx, lessorId, false, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, ipCount, rep.IpMonthsTotal, "IP months should stay same")
	require.Equal(t, uint64(0), rep.IpsLeasedNow, "IPs leased should decrease")
	require.Equal(t, uint64(0), rep.ContractsActive, "Active contracts should decrease")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated contracts should increment")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated should stay 0")
}

func TestLogLessorContractTerminationWithViolation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(400)
	ipCount := uint64(32)

	// Initialize and activate a contract first
	keeper.InitializeLessorReputation(ctx, lessorId)
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	// Log termination with violation
	err = keeper.LogLessorContractTermination(ctx, lessorId, true, ipCount)
	require.NoError(t, err)

	// Verify reputation was updated correctly
	rep, found := keeper.GetLessorReputation(ctx, lessorId)
	require.True(t, found)
	require.Equal(t, uint64(0), rep.ContractsActive, "Active contracts should decrease")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated contracts should increment")
	require.Equal(t, uint64(1), rep.ContractsTerminatedViolated, "Violated should increment")
}

// Edge Cases and Validation Tests for Lessor

func TestLogLessorContractActivationZeroLessorId(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)

	err := keeper.LogLessorContractActivation(ctx, 0, 256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lessor id can't be nil")
}

func TestLogLessorContractActivationZeroIpCount(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(500)

	keeper.InitializeLessorReputation(ctx, lessorId)

	err := keeper.LogLessorContractActivation(ctx, lessorId, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "incorrect ip count")
}

func TestLogLessorContractActivationMissingReputation(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(600)

	// Don't initialize reputation first
	err := keeper.LogLessorContractActivation(ctx, lessorId, 256)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can not find lessor")
}

func TestLogLessorReputationFullLifecycle(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	lessorId := uint64(700)
	ipCount := uint64(1024)

	// Initialize
	keeper.InitializeLessorReputation(ctx, lessorId)

	// Activate contract
	err := keeper.LogLessorContractActivation(ctx, lessorId, ipCount)
	require.NoError(t, err)

	rep, _ := keeper.GetLessorReputation(ctx, lessorId)
	require.Equal(t, uint64(1), rep.ContractsActive)
	require.Equal(t, ipCount, rep.IpsLeasedNow)
	require.Equal(t, ipCount, rep.IpMonthsTotal)

	// Renew contract twice
	err = keeper.LogLessorContractRenewal(ctx, lessorId, ipCount)
	require.NoError(t, err)
	err = keeper.LogLessorContractRenewal(ctx, lessorId, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLessorReputation(ctx, lessorId)
	require.Equal(t, uint64(1), rep.ContractsActive, "Active should stay 1")
	require.Equal(t, ipCount, rep.IpsLeasedNow, "IPs leased should stay same")
	require.Equal(t, ipCount*3, rep.IpMonthsTotal, "IP months should accumulate")

	// Terminate contract normally
	err = keeper.LogLessorContractTermination(ctx, lessorId, false, ipCount)
	require.NoError(t, err)

	rep, _ = keeper.GetLessorReputation(ctx, lessorId)
	require.Equal(t, uint64(0), rep.ContractsActive, "Active should be 0")
	require.Equal(t, uint64(0), rep.IpsLeasedNow, "IPs leased should be 0")
	require.Equal(t, ipCount*3, rep.IpMonthsTotal, "IP months should stay accumulated")
	require.Equal(t, uint64(1), rep.ContractsTerminated, "Terminated should be 1")
	require.Equal(t, uint64(0), rep.ContractsTerminatedViolated, "Violated should be 0")
}
