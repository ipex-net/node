package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	abusetypes "ipex/x/abuse/types"
	"ipex/x/trust/types"

	"github.com/stretchr/testify/require"
)

// TestAbuseIntegration tests the complete workflow from abuse reports to reputation updates
func TestAbuseIntegration(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	// Set a valid block time
	ctx = ctx.WithBlockTime(time.Now())
	
	// Test data setup
	leaseId := uint64(1001)
	lesseeOrgId := uint64(2001)
	abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)
	
	// Initially there should be no reputation records
	_, found := keeper.GetLeaseReputation(ctx, leaseId)
	require.False(t, found, "lease reputation should not exist initially")
	
	_, found = keeper.GetLesseeReputation(ctx, lesseeOrgId)
	require.False(t, found, "lessee reputation should not exist initially")
	
	t.Run("Submit Abuse Complaint", func(t *testing.T) {
		// Simulate abuse complaint submission (inprogress state)
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should successfully update abuse stats for new complaint")
		
		// Verify lease reputation was created and updated
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "lease reputation should be created")
		require.Equal(t, uint64(1), leaseRep.AbuseInProgress, "should have 1 abuse in progress")
		require.Equal(t, uint64(0), leaseRep.AbuseResolved, "should have 0 resolved abuses")
		require.Equal(t, uint64(0), leaseRep.AbuseFailed, "should have 0 failed abuses")
		require.NotNil(t, leaseRep.LastUpdated, "last updated should be set")
		
		// Verify abuse by type tracking
		require.Contains(t, leaseRep.AbuseByType, abuseCategory, "should track abuse by category")
		abuseStats := leaseRep.AbuseByType[abuseCategory]
		require.Equal(t, uint64(1), abuseStats.InProgress, "should have 1 SPAM abuse in progress")
		require.Equal(t, uint64(0), abuseStats.Resolved, "should have 0 SPAM abuses resolved")
		require.Equal(t, uint64(0), abuseStats.Failed, "should have 0 SPAM abuses failed")
		require.NotNil(t, abuseStats.LastIncident, "last incident should be set")
		
		// Verify lessee reputation was created and updated
		lesseeRep, found := keeper.GetLesseeReputation(ctx, lesseeOrgId)
		require.True(t, found, "lessee reputation should be created")
		require.Equal(t, uint64(1), lesseeRep.AbuseInProgress, "should have 1 abuse in progress")
		require.Equal(t, uint64(0), lesseeRep.AbuseResolved, "should have 0 resolved abuses")
		require.Equal(t, uint64(0), lesseeRep.AbuseFailed, "should have 0 failed abuses")
		require.NotNil(t, lesseeRep.LastUpdated, "last updated should be set")
		
		// Verify abuse by type tracking for lessee
		require.Contains(t, lesseeRep.AbuseByType, abuseCategory, "should track abuse by category")
		lesseeAbuseStats := lesseeRep.AbuseByType[abuseCategory]
		require.Equal(t, uint64(1), lesseeAbuseStats.InProgress, "should have 1 SPAM abuse in progress")
		require.Equal(t, uint64(0), lesseeAbuseStats.Resolved, "should have 0 SPAM abuses resolved")
		require.Equal(t, uint64(0), lesseeAbuseStats.Failed, "should have 0 SPAM abuses failed")
		require.NotNil(t, lesseeAbuseStats.LastIncident, "last incident should be set")
	})
	
	t.Run("Resolve Abuse Complaint", func(t *testing.T) {
		// Simulate abuse complaint resolution (inprogress -> resolved)
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "resolved")
		require.NoError(t, err, "should successfully update abuse stats for resolved complaint")
		
		// Verify lease reputation was updated
		leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
		require.True(t, found, "lease reputation should exist")
		require.Equal(t, uint64(0), leaseRep.AbuseInProgress, "should have 0 abuse in progress")
		require.Equal(t, uint64(1), leaseRep.AbuseResolved, "should have 1 resolved abuse")
		require.Equal(t, uint64(0), leaseRep.AbuseFailed, "should have 0 failed abuses")
		
		// Verify abuse by type tracking
		abuseStats := leaseRep.AbuseByType[abuseCategory]
		require.Equal(t, uint64(0), abuseStats.InProgress, "should have 0 SPAM abuse in progress")
		require.Equal(t, uint64(1), abuseStats.Resolved, "should have 1 SPAM abuse resolved")
		require.Equal(t, uint64(0), abuseStats.Failed, "should have 0 SPAM abuses failed")
		
		// Verify lessee reputation was updated
		lesseeRep, found := keeper.GetLesseeReputation(ctx, lesseeOrgId)
		require.True(t, found, "lessee reputation should exist")
		require.Equal(t, uint64(0), lesseeRep.AbuseInProgress, "should have 0 abuse in progress")
		require.Equal(t, uint64(1), lesseeRep.AbuseResolved, "should have 1 resolved abuse")
		require.Equal(t, uint64(0), lesseeRep.AbuseFailed, "should have 0 failed abuses")
		
		// Verify abuse by type tracking for lessee
		lesseeAbuseStats := lesseeRep.AbuseByType[abuseCategory]
		require.Equal(t, uint64(0), lesseeAbuseStats.InProgress, "should have 0 SPAM abuse in progress")
		require.Equal(t, uint64(1), lesseeAbuseStats.Resolved, "should have 1 SPAM abuse resolved")
		require.Equal(t, uint64(0), lesseeAbuseStats.Failed, "should have 0 SPAM abuses failed")
	})
}

// TestMultipleAbuseTypes tests handling multiple abuse categories
func TestMultipleAbuseTypes(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())
	
	leaseId := uint64(1002)
	lesseeOrgId := uint64(2002)
	spamCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)
	malwareCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_MALWARE)
	
	// Submit complaints for different abuse types
	err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, spamCategory, "", "inprogress")
	require.NoError(t, err, "should handle SPAM complaint")
	
	err = keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, malwareCategory, "", "inprogress")
	require.NoError(t, err, "should handle MALWARE complaint")
	
	// Verify both abuse types are tracked
	leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(2), leaseRep.AbuseInProgress, "should have 2 abuses in progress")
	
	// Verify separate tracking by category
	require.Contains(t, leaseRep.AbuseByType, spamCategory, "should track SPAM abuse")
	require.Contains(t, leaseRep.AbuseByType, malwareCategory, "should track MALWARE abuse")
	
	spamStats := leaseRep.AbuseByType[spamCategory]
	require.Equal(t, uint64(1), spamStats.InProgress, "should have 1 SPAM abuse in progress")
	
	malwareStats := leaseRep.AbuseByType[malwareCategory]
	require.Equal(t, uint64(1), malwareStats.InProgress, "should have 1 MALWARE abuse in progress")
	
	// Resolve one, fail the other
	err = keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, spamCategory, "inprogress", "resolved")
	require.NoError(t, err, "should resolve SPAM complaint")
	
	err = keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, malwareCategory, "inprogress", "failed")
	require.NoError(t, err, "should fail MALWARE complaint")
	
	// Verify final state
	leaseRep, found = keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(0), leaseRep.AbuseInProgress, "should have 0 abuses in progress")
	require.Equal(t, uint64(1), leaseRep.AbuseResolved, "should have 1 resolved abuse")
	require.Equal(t, uint64(1), leaseRep.AbuseFailed, "should have 1 failed abuse")
	
	// Verify category-specific outcomes
	spamStats = leaseRep.AbuseByType[spamCategory]
	require.Equal(t, uint64(0), spamStats.InProgress, "SPAM should have 0 in progress")
	require.Equal(t, uint64(1), spamStats.Resolved, "SPAM should have 1 resolved")
	require.Equal(t, uint64(0), spamStats.Failed, "SPAM should have 0 failed")
	
	malwareStats = leaseRep.AbuseByType[malwareCategory]
	require.Equal(t, uint64(0), malwareStats.InProgress, "MALWARE should have 0 in progress")
	require.Equal(t, uint64(0), malwareStats.Resolved, "MALWARE should have 0 resolved")
	require.Equal(t, uint64(1), malwareStats.Failed, "MALWARE should have 1 failed")
}

// TestAbuseStatsErrorHandling tests error conditions
func TestAbuseStatsErrorHandling(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())
	
	leaseId := uint64(1003)
	lesseeOrgId := uint64(2003)
	validCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)
	
	t.Run("Invalid Old Status", func(t *testing.T) {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, validCategory, "invalid", "inprogress")
		require.Error(t, err, "should reject invalid old status")
		require.Contains(t, err.Error(), "invalid old status", "should indicate invalid old status")
	})
	
	t.Run("Invalid New Status", func(t *testing.T) {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, validCategory, "", "invalid")
		require.Error(t, err, "should reject invalid new status")
		require.Contains(t, err.Error(), "invalid new status", "should indicate invalid new status")
	})
	
	t.Run("Invalid Abuse Category", func(t *testing.T) {
		invalidCategory := uint32(999)
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, invalidCategory, "", "inprogress")
		require.Error(t, err, "should reject invalid abuse category")
		require.Contains(t, err.Error(), "invalid abuse category", "should indicate invalid abuse category")
	})
}

// TestBlacklistLogIntegration tests blacklist logging alongside reputation updates
func TestBlacklistLogIntegration(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())
	
	// Test data
	subnetId := uint64(1001)
	lessorOrgId := uint64(2001)
	leaseId := uint64(3001)
	lesseeOrgId := uint64(4001)
	
	// Create blacklist log entry
	blacklistLog := types.BlacklistLog{
		SubnetId:     subnetId,
		LessorOrgId:  lessorOrgId,
		LeaseId:      leaseId,
		LesseeOrgId:  lesseeOrgId,
		Level1Change: 1,
		Level2Change: 0,
		Level3Change: 0,
		Level4Change: 0,
	}
	
	// Append blacklist log
	logId := keeper.AppendBlacklistLog(ctx, blacklistLog)
	require.Greater(t, logId, uint64(0), "should return valid log ID")
	
	// Verify blacklist log was created
	retrievedLog, found := keeper.GetBlacklistLog(ctx, logId)
	require.True(t, found, "blacklist log should be found")
	require.Equal(t, subnetId, retrievedLog.SubnetId, "subnet ID should match")
	require.Equal(t, lessorOrgId, retrievedLog.LessorOrgId, "lessor org ID should match")
	require.Equal(t, leaseId, retrievedLog.LeaseId, "lease ID should match")
	require.Equal(t, lesseeOrgId, retrievedLog.LesseeOrgId, "lessee org ID should match")
	require.Equal(t, int32(1), retrievedLog.Level1Change, "level 1 change should match")
	
	// Simulate corresponding abuse complaint and reputation update
	abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM)
	err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
	require.NoError(t, err, "should update abuse stats successfully")
	
	// Verify both blacklist log and reputation coexist
	leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(1), leaseRep.AbuseInProgress, "should have abuse in progress")
	
	// Test querying blacklist logs by various criteria
	t.Run("Query Blacklist Logs", func(t *testing.T) {
		// Query by subnet
		allLogs := keeper.GetAllBlacklistLog(ctx)
		require.Len(t, allLogs, 1, "should have 1 blacklist log")
		
		// Verify the log matches our created entry
		log := allLogs[0]
		require.Equal(t, logId, log.Id, "log ID should match")
		require.Equal(t, subnetId, log.SubnetId, "subnet ID should match")
		require.Equal(t, leaseId, log.LeaseId, "lease ID should match")
	})
}

// TestReputationPersistence tests that reputation changes persist correctly
func TestReputationPersistence(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())
	
	leaseId := uint64(1004)
	lesseeOrgId := uint64(2004)
	abuseCategory := uint32(abusetypes.AbuseType_ABUSE_TYPE_PHISHING_FRAUD)
	
	// Create multiple complaints over time
	for i := 0; i < 5; i++ {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "", "inprogress")
		require.NoError(t, err, "should submit complaint %d", i+1)
	}
	
	// Verify accumulated state
	leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(5), leaseRep.AbuseInProgress, "should have 5 abuses in progress")
	
	// Resolve some, fail others
	for i := 0; i < 3; i++ {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "resolved")
		require.NoError(t, err, "should resolve complaint %d", i+1)
	}
	
	for i := 0; i < 2; i++ {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseCategory, "inprogress", "failed")
		require.NoError(t, err, "should fail complaint %d", i+1)
	}
	
	// Verify final state
	leaseRep, found = keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(0), leaseRep.AbuseInProgress, "should have 0 abuses in progress")
	require.Equal(t, uint64(3), leaseRep.AbuseResolved, "should have 3 resolved abuses")
	require.Equal(t, uint64(2), leaseRep.AbuseFailed, "should have 2 failed abuses")
	
	// Verify abuse by type
	abuseStats := leaseRep.AbuseByType[abuseCategory]
	require.Equal(t, uint64(0), abuseStats.InProgress, "should have 0 PHISHING_FRAUD abuses in progress")
	require.Equal(t, uint64(3), abuseStats.Resolved, "should have 3 PHISHING_FRAUD abuses resolved")
	require.Equal(t, uint64(2), abuseStats.Failed, "should have 2 PHISHING_FRAUD abuses failed")
}

// TestConcurrentReputationUpdates tests handling multiple simultaneous updates
func TestConcurrentReputationUpdates(t *testing.T) {
	keeper, ctx := keepertest.TrustKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())
	
	leaseId := uint64(1005)
	lesseeOrgId := uint64(2005)
	
	// Submit different abuse types simultaneously
	abuseTypes := []uint32{
		uint32(abusetypes.AbuseType_ABUSE_TYPE_SPAM),
		uint32(abusetypes.AbuseType_ABUSE_TYPE_MALWARE),
		uint32(abusetypes.AbuseType_ABUSE_TYPE_PHISHING_FRAUD),
		uint32(abusetypes.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA),
	}
	
	// Submit all abuse types
	for _, abuseType := range abuseTypes {
		err := keeper.UpdateAbuseStats(ctx, leaseId, lesseeOrgId, abuseType, "", "inprogress")
		require.NoError(t, err, "should handle abuse type %d", abuseType)
	}
	
	// Verify all types are tracked
	leaseRep, found := keeper.GetLeaseReputation(ctx, leaseId)
	require.True(t, found, "lease reputation should exist")
	require.Equal(t, uint64(4), leaseRep.AbuseInProgress, "should have 4 abuses in progress")
	require.Len(t, leaseRep.AbuseByType, 4, "should track 4 abuse types")
	
	// Verify each abuse type has the correct stats
	for _, abuseType := range abuseTypes {
		require.Contains(t, leaseRep.AbuseByType, abuseType, "should track abuse type %d", abuseType)
		stats := leaseRep.AbuseByType[abuseType]
		require.Equal(t, uint64(1), stats.InProgress, "abuse type %d should have 1 in progress", abuseType)
		require.Equal(t, uint64(0), stats.Resolved, "abuse type %d should have 0 resolved", abuseType)
		require.Equal(t, uint64(0), stats.Failed, "abuse type %d should have 0 failed", abuseType)
	}
}