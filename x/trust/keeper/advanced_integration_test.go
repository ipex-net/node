package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

// TestBlacklistReputationLevels tests the 4-level blacklist reputation system
func TestBlacklistReputationLevels(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	subnetId := uint64(1001)
	lessorOrgId := uint64(2001)
	leaseId := uint64(3001)
	lesseeOrgId := uint64(4001)

	t.Run("Level 1 Blacklist Changes", func(t *testing.T) {
		// Create blacklist log with level 1 changes
		blacklistLog := types.BlacklistLog{
			SubnetId:     subnetId,
			LessorOrgId:  lessorOrgId,
			LeaseId:      leaseId,
			LesseeOrgId:  lesseeOrgId,
			Level1Change: 5,  // Significant level 1 increase
			Level2Change: 0,
			Level3Change: 0,
			Level4Change: 0,
		}

		logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
		require.Greater(t, logId, uint64(0), "should create blacklist log")

		// Verify the log was created with correct level changes
		retrievedLog, found := keeper.GetBlacklistLog(ctx, logId)
		require.True(t, found, "should find blacklist log")
		require.Equal(t, int32(5), retrievedLog.Level1Change, "should have level 1 change")
		require.Equal(t, int32(0), retrievedLog.Level2Change, "should have no level 2 change")
		require.Equal(t, int32(0), retrievedLog.Level3Change, "should have no level 3 change")
		require.Equal(t, int32(0), retrievedLog.Level4Change, "should have no level 4 change")
	})

	t.Run("Multiple Level Changes", func(t *testing.T) {
		// Create blacklist log with changes across all levels
		blacklistLog := types.BlacklistLog{
			SubnetId:     subnetId,
			LessorOrgId:  lessorOrgId,
			LeaseId:      leaseId + 1, // Different lease
			LesseeOrgId:  lesseeOrgId,
			Level1Change: 2,
			Level2Change: 3,
			Level3Change: 1,
			Level4Change: 1,
		}

		logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
		require.Greater(t, logId, uint64(0), "should create blacklist log")

		// Verify all level changes
		retrievedLog, found := keeper.GetBlacklistLog(ctx, logId)
		require.True(t, found, "should find blacklist log")
		require.Equal(t, int32(2), retrievedLog.Level1Change, "should have level 1 change")
		require.Equal(t, int32(3), retrievedLog.Level2Change, "should have level 2 change")
		require.Equal(t, int32(1), retrievedLog.Level3Change, "should have level 3 change")
		require.Equal(t, int32(1), retrievedLog.Level4Change, "should have level 4 change")
	})

	t.Run("Negative Level Changes", func(t *testing.T) {
		// Test reduction in blacklist levels (reputation improvement)
		blacklistLog := types.BlacklistLog{
			SubnetId:     subnetId,
			LessorOrgId:  lessorOrgId,
			LeaseId:      leaseId + 2, // Different lease
			LesseeOrgId:  lesseeOrgId,
			Level1Change: -1, // Reputation improvement
			Level2Change: -2,
			Level3Change: 0,
			Level4Change: 0,
		}

		logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
		require.Greater(t, logId, uint64(0), "should create blacklist log")

		// Verify negative changes (reputation improvements)
		retrievedLog, found := keeper.GetBlacklistLog(ctx, logId)
		require.True(t, found, "should find blacklist log")
		require.Equal(t, int32(-1), retrievedLog.Level1Change, "should have negative level 1 change")
		require.Equal(t, int32(-2), retrievedLog.Level2Change, "should have negative level 2 change")
	})
}

// TestSubnetReputationIntegration tests subnet reputation tracking and updates
func TestSubnetReputationIntegration(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	subnetId := uint64(1001)

	t.Run("Initialize Subnet Reputation", func(t *testing.T) {
		// Initialize subnet reputation
		keeper.InitializeSubnetReputation(ctx, subnetId)

		// Verify initialization
		subnetRep, found := keeper.GetSubnetReputation(ctx, subnetId)
		require.True(t, found, "should find subnet reputation")
		require.Equal(t, subnetId, subnetRep.Index, "should have correct subnet ID")
		require.Equal(t, uint32(0), subnetRep.Level1Count, "should start with zero level 1 count")
		require.Equal(t, uint32(0), subnetRep.Level2Count, "should start with zero level 2 count")
		require.Equal(t, uint32(0), subnetRep.Level3Count, "should start with zero level 3 count")
		require.Equal(t, uint32(0), subnetRep.Level4Count, "should start with zero level 4 count")
		require.Nil(t, subnetRep.LastUpdated, "should start with nil last updated timestamp")
	})

	t.Run("Update Subnet Reputation", func(t *testing.T) {
		// Get current reputation
		subnetRep, found := keeper.GetSubnetReputation(ctx, subnetId)
		require.True(t, found, "should find subnet reputation")

		// Update reputation levels
		subnetRep.Level1Count = 5
		subnetRep.Level2Count = 3
		subnetRep.Level3Count = 1
		subnetRep.Level4Count = 0

		// Save updated reputation
		keeper.SetSubnetReputation(ctx, subnetRep)

		// Verify updates
		updatedRep, found := keeper.GetSubnetReputation(ctx, subnetId)
		require.True(t, found, "should find updated subnet reputation")
		require.Equal(t, uint32(5), updatedRep.Level1Count, "should have updated level 1 count")
		require.Equal(t, uint32(3), updatedRep.Level2Count, "should have updated level 2 count")
		require.Equal(t, uint32(1), updatedRep.Level3Count, "should have updated level 3 count")
		require.Equal(t, uint32(0), updatedRep.Level4Count, "should have updated level 4 count")
	})

	t.Run("Multiple Subnet Reputations", func(t *testing.T) {
		// Create multiple subnet reputations
		subnetIds := []uint64{1002, 1003, 1004}
		
		for _, id := range subnetIds {
			keeper.InitializeSubnetReputation(ctx, id)
		}

		// Verify all subnets were created
		allSubnets := keeper.GetAllSubnetReputation(ctx)
		require.GreaterOrEqual(t, len(allSubnets), 4, "should have at least 4 subnet reputations") // Original + 3 new
		
		// Verify each subnet has unique index
		subnetMap := make(map[uint64]bool)
		for _, subnet := range allSubnets {
			require.False(t, subnetMap[subnet.Index], "should not have duplicate subnet indices")
			subnetMap[subnet.Index] = true
		}
	})
}

// TestEdgeCasesAndBoundaries tests edge cases and boundary conditions
func TestEdgeCasesAndBoundaries(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	t.Run("Zero Values", func(t *testing.T) {
		// Test handling of zero values
		leaseId := uint64(0)
		lesseeOrgId := uint64(0)
		abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)

		// Should handle zero IDs gracefully
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should handle zero lease and lessee IDs")

		// Verify reputation was created for zero IDs
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should create lease reputation for zero ID")
		require.Equal(t, uint64(1), leaseRep.AbuseInProgress, "should have 1 abuse in progress")
	})

	t.Run("Large Numbers", func(t *testing.T) {
		// Test with large ID values
		leaseId := uint64(18446744073709551615) // Max uint64
		lesseeOrgId := uint64(18446744073709551614) // Max uint64 - 1
		abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_MALWARE)

		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should handle large ID values")

		// Verify reputation was created
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should create lease reputation for large ID")
		require.Equal(t, leaseId, leaseRep.Index, "should have correct lease ID")
	})

	t.Run("Rapid State Transitions", func(t *testing.T) {
		// Test rapid state transitions
		leaseId := uint64(5001)
		lesseeOrgId := uint64(6001)
		abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_PHISHING_FRAUD)

		// Submit complaint
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should submit complaint")

		// Immediately resolve
		err = keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "resolved")
		require.NoError(t, err, "should resolve complaint")

		// Verify final state
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should find lease reputation")
		require.Equal(t, uint64(0), leaseRep.AbuseInProgress, "should have no abuse in progress")
		require.Equal(t, uint64(1), leaseRep.AbuseResolved, "should have 1 resolved abuse")
	})

	t.Run("Counter Underflow Protection", func(t *testing.T) {
		// Test that counters don't go below zero
		leaseId := uint64(5002)
		lesseeOrgId := uint64(6002)
		abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)

		// Try to resolve an abuse that was never submitted
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "resolved")
		require.NoError(t, err, "should handle transition from non-existent inprogress state")

		// Verify counters didn't underflow
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should create lease reputation")
		require.Equal(t, uint64(0), leaseRep.AbuseInProgress, "should not have negative abuse in progress")
		require.Equal(t, uint64(1), leaseRep.AbuseResolved, "should have 1 resolved abuse")
	})
}

// TestQueryIndexingAndPerformance tests query performance and indexing
func TestQueryIndexingAndPerformance(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Create test data
	subnetId := uint64(1001)
	lessorOrgId := uint64(2001)
	lesseeOrgId := uint64(4001)

	t.Run("Blacklist Log Indexing", func(t *testing.T) {
		// Create multiple blacklist logs with different criteria
		testCases := []struct {
			leaseId     uint64
			description string
		}{
			{3001, "First lease"},
			{3002, "Second lease"},
			{3003, "Third lease"},
			{3004, "Fourth lease"},
			{3005, "Fifth lease"},
		}

		logIds := make([]uint64, len(testCases))
		for i, tc := range testCases {
			blacklistLog := types.BlacklistLog{
				SubnetId:     subnetId,
				LessorOrgId:  lessorOrgId,
				LeaseId:      tc.leaseId,
				LesseeOrgId:  lesseeOrgId,
				Level1Change: int32(i + 1),
				Level2Change: 0,
				Level3Change: 0,
				Level4Change: 0,
			}

			logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
			require.Greater(t, logId, uint64(0), "should create blacklist log for %s", tc.description)
			logIds[i] = logId
		}

		// Test querying all logs
		allLogs := keeper.GetAllBlacklistLog(ctx)
		require.GreaterOrEqual(t, len(allLogs), 5, "should have at least 5 blacklist logs")

		// Test that all our logs are present
		foundLogs := make(map[uint64]bool)
		for _, log := range allLogs {
			foundLogs[log.Id] = true
		}

		for _, logId := range logIds {
			require.True(t, foundLogs[logId], "should find blacklist log with ID %d", logId)
		}
	})

	t.Run("Reputation Query Performance", func(t *testing.T) {
		// Create many reputation records
		numRecords := 100
		leaseIds := make([]uint64, numRecords)
		
		for i := 0; i < numRecords; i++ {
			leaseId := uint64(7000 + i)
			lesseeOrgId := uint64(8000 + i)
			abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)
			
			// Create reputation record
			err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
			require.NoError(t, err, "should create reputation record %d", i)
			leaseIds[i] = leaseId
		}

		// Test bulk retrieval
		allLeaseReputations := keeper.GetAllLeaseReputation(ctx)
		require.GreaterOrEqual(t, len(allLeaseReputations), numRecords, "should have at least %d lease reputations", numRecords)

		allLesseeReputations := keeper.GetAllLesseeReputation(ctx)
		require.GreaterOrEqual(t, len(allLesseeReputations), numRecords, "should have at least %d lessee reputations", numRecords)

		// Test individual lookups
		for i := 0; i < 10; i++ { // Test first 10 records
			leaseRep, found := keeper.GetLeaseReputation(ctx, leaseIds[i])
			require.True(t, found, "should find lease reputation for ID %d", leaseIds[i])
			require.Equal(t, leaseIds[i], leaseRep.Index, "should have correct lease ID")
		}
	})
}

// TestTimestampConsistency tests timestamp handling across operations
func TestTimestampConsistency(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	
	// Use a specific time for consistency
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(testTime)

	leaseId := uint64(9001)
	lesseeOrgId := uint64(9002)
	abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)

	t.Run("Initial Timestamp", func(t *testing.T) {
		// Submit initial complaint
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should submit complaint")

		// Verify timestamps
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should find lease reputation")
		require.NotNil(t, leaseRep.LastUpdated, "should have last updated timestamp")
		require.Equal(t, testTime, *leaseRep.LastUpdated, "should have correct last updated time")

		// Check abuse-specific timestamp
		require.Contains(t, leaseRep.AbuseByType, abuseCategory, "should have abuse category")
		abuseStats := leaseRep.AbuseByType[abuseCategory]
		require.NotNil(t, abuseStats.LastIncident, "should have last incident timestamp")
		require.Equal(t, testTime, *abuseStats.LastIncident, "should have correct last incident time")
	})

	t.Run("Updated Timestamp", func(t *testing.T) {
		// Move time forward
		newTime := testTime.Add(24 * time.Hour)
		ctx = ctx.WithBlockTime(newTime)

		// Resolve the complaint
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "resolved")
		require.NoError(t, err, "should resolve complaint")

		// Verify timestamps were updated
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should find lease reputation")
		require.NotNil(t, leaseRep.LastUpdated, "should have last updated timestamp")
		require.Equal(t, newTime, *leaseRep.LastUpdated, "should have updated last updated time")

		// Note: LastIncident timestamp should remain unchanged since we didn't create a new incident
		abuseStats := leaseRep.AbuseByType[abuseCategory]
		require.NotNil(t, abuseStats.LastIncident, "should have last incident timestamp")
		require.Equal(t, testTime, *abuseStats.LastIncident, "should keep original last incident time")
	})

	t.Run("Blacklist Log Timestamp", func(t *testing.T) {
		// Create blacklist log
		blacklistLog := types.BlacklistLog{
			SubnetId:     1001,
			LessorOrgId:  2001,
			LeaseId:      leaseId,
			LesseeOrgId:  lesseeOrgId,
			Level1Change: 1,
			Level2Change: 0,
			Level3Change: 0,
			Level4Change: 0,
		}

		logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
		require.Greater(t, logId, uint64(0), "should create blacklist log")

		// Verify log was created
		_, found := keeper.GetBlacklistLog(ctx, logId)
		require.True(t, found, "should find blacklist log")
		// Note: Timestamp is only set when using UpdateSubnetReputation method, not AppendBlacklistLog directly
	})
}

// TestDataIntegrity tests data consistency and integrity
func TestDataIntegrity(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	t.Run("Lease Reputation Consistency", func(t *testing.T) {
		leaseId := uint64(10001)
		lesseeOrgId := uint64(10002)
		
		// Submit multiple complaints of different types
		abuseTypes := []uint32{
			uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM),
			uint32(abusetypes.AbuseType_ABUSE_TYPE_MALWARE),
			uint32(abusetypes.AbuseType_ABUSE_TYPE_PHISHING_FRAUD),
		}

		for _, abuseType := range abuseTypes {
			err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseType, "", "inprogress")
			require.NoError(t, err, "should submit complaint for abuse type %d", abuseType)
		}

		// Verify total counts match category-specific counts
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "should find lease reputation")
		require.Equal(t, uint64(3), leaseRep.AbuseInProgress, "should have 3 abuses in progress")

		// Check that category-specific counts add up
		totalInProgress := uint64(0)
		for _, abuseType := range abuseTypes {
			require.Contains(t, leaseRep.AbuseByType, abuseType, "should have abuse type %d", abuseType)
			stats := leaseRep.AbuseByType[abuseType]
			totalInProgress += stats.InProgress
		}
		require.Equal(t, leaseRep.AbuseInProgress, totalInProgress, "total should match sum of categories")
	})

	t.Run("Lessee Reputation Consistency", func(t *testing.T) {
		// Test that lessee reputation is consistent across multiple leases
		lesseeOrgId := uint64(10003)
		leaseIds := []uint64{10004, 10005, 10006}
		abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)

		// Submit complaints for same lessee across multiple leases
		for _, leaseId := range leaseIds {
			err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
			require.NoError(t, err, "should submit complaint for lease %d", leaseId)
		}

		// Verify lessee reputation reflects all leases
		lesseeRep, found := keeper.GetLesseeReputation(ctx, lesseeOrgId)
		require.True(t, found, "should find lessee reputation")
		require.Equal(t, uint64(3), lesseeRep.AbuseInProgress, "should have 3 abuses in progress across all leases")

		// Verify each lease has its own reputation
		for _, leaseId := range leaseIds {
			leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
			require.True(t, found, "should find lease reputation for lease %d", leaseId)
			require.Equal(t, uint64(1), leaseRep.AbuseInProgress, "each lease should have 1 abuse in progress")
		}
	})
}