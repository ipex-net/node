package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/x/abuse/types"

	"github.com/stretchr/testify/require"
)

// TestComplaintLifecycleIntegration tests the complete complaint lifecycle
func TestComplaintLifecycleIntegration(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data - using valid fields for a complete complaint
	leaseId := uint64(1001)
	subnetId := uint64(2001)
	lessorOrgId := uint64(3001)
	lesseeOrgId := uint64(4001)
	sourceIp := uint64(0x08080808) // 8.8.8.8
	abuseType := types.AbuseType_ABUSE_TYPE_SPAM
	incidentTime := ctx.BlockTime().Add(-time.Hour)
	blockTime := ctx.BlockTime()
	deadline := blockTime.Add(time.Hour * 24)

	t.Run("Create Complaint", func(t *testing.T) {
		// Create complaint using all required fields
		complaint := types.Complaint{
			LeaseId:            leaseId,
			SubnetId:           subnetId,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId,
			AbuseType:          abuseType,
			IncidentTime:       &incidentTime,
			SourceIp:           sourceIp,
			LessorTicketRef:    "TICKET-123",
			EvidenceHash:       "QmT4AeWE9Q9EaoyLJiqaZuYQ8mJeq4ZBncjjFH9dQ9uDVA", // Valid IPFS hash
			Summary:            "This is a test spam complaint with sufficient length to meet requirements",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &blockTime,
			ResolutionDeadline: &deadline,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &blockTime,
			ClosedAt:           nil,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)
		require.Greater(t, complaintId, uint64(0), "should create complaint with valid ID")

		// Verify complaint was created correctly
		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find created complaint")
		require.Equal(t, complaintId, retrievedComplaint.Id, "should have correct ID")
		require.Equal(t, leaseId, retrievedComplaint.LeaseId, "should have correct lease ID")
		require.Equal(t, subnetId, retrievedComplaint.SubnetId, "should have correct subnet ID")
		require.Equal(t, types.ComplaintState_STATE_SUBMITTED, retrievedComplaint.State, "should be in submitted state")
		require.Equal(t, abuseType, retrievedComplaint.AbuseType, "should have correct abuse type")
		require.Equal(t, sourceIp, retrievedComplaint.SourceIp, "should have correct source IP")
		require.False(t, retrievedComplaint.ResolutionFailed, "should not be marked as failed initially")
	})

	t.Run("Update Complaint State", func(t *testing.T) {
		// Create a new complaint for state testing
		complaint := types.Complaint{
			LeaseId:            leaseId + 1,
			SubnetId:           subnetId + 1,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId,
			AbuseType:          types.AbuseType_ABUSE_TYPE_MALWARE,
			IncidentTime:       &incidentTime,
			SourceIp:           sourceIp + 1,
			Summary:            "Malware complaint for state transition testing purposes",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &blockTime,
			ResolutionDeadline: &deadline,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &blockTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Update state to RESPONDED
		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")

		retrievedComplaint.State = types.ComplaintState_STATE_RESPONDED
		updatedTime := ctx.BlockTime().Add(time.Hour)
		retrievedComplaint.LastActivityAt = &updatedTime
		keeper.SetComplaint(ctx, retrievedComplaint)

		// Verify state update
		updatedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find updated complaint")
		require.Equal(t, types.ComplaintState_STATE_RESPONDED, updatedComplaint.State, "should have updated state")
		require.Equal(t, updatedTime, *updatedComplaint.LastActivityAt, "should have updated timestamp")
	})

	t.Run("Thread Management", func(t *testing.T) {
		// Create complaint for thread testing
		complaint := types.Complaint{
			LeaseId:            leaseId + 2,
			SubnetId:           subnetId + 2,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId,
			AbuseType:          types.AbuseType_ABUSE_TYPE_PHISHING_FRAUD,
			IncidentTime:       &incidentTime,
			SourceIp:           sourceIp + 2,
			Summary:            "Phishing complaint for thread testing and message management",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &blockTime,
			ResolutionDeadline: &deadline,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &blockTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Add lessee response to thread
		lesseeResponseTime := ctx.BlockTime().Add(time.Hour)
		lesseeMessage := &types.ThreadMessage{
			MessageType:  types.MessageType_MESSAGE_TYPE_LESSEE_RESPONSE,
			Content:      "This appears to be a false positive. We have investigated and found no phishing activity.",
			EvidenceHash: "",
			SentAt:       &lesseeResponseTime,
		}

		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")

		retrievedComplaint.Thread = append(retrievedComplaint.Thread, lesseeMessage)
		retrievedComplaint.State = types.ComplaintState_STATE_RESPONDED
		keeper.SetComplaint(ctx, retrievedComplaint)

		// Add lessor follow-up
		lessorFollowupTime := ctx.BlockTime().Add(time.Hour * 2)
		lessorMessage := &types.ThreadMessage{
			MessageType:  types.MessageType_MESSAGE_TYPE_LESSOR_FOLLOWUP,
			Content:      "We have additional evidence. Please review the attached logs showing suspicious activity.",
			EvidenceHash: "QmT4AeWE9Q9EaoyLJiqaZuYQ8mJeq4ZBncjjFH9dQ9uDVB", // Different valid IPFS hash
			SentAt:       &lessorFollowupTime,
		}

		finalComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")

		finalComplaint.Thread = append(finalComplaint.Thread, lessorMessage)
		finalComplaint.State = types.ComplaintState_STATE_AWAITING_RESPONSE
		keeper.SetComplaint(ctx, finalComplaint)

		// Verify thread management
		threadComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint with thread")
		require.Len(t, threadComplaint.Thread, 2, "should have 2 messages in thread")
		require.Equal(t, types.MessageType_MESSAGE_TYPE_LESSEE_RESPONSE, threadComplaint.Thread[0].MessageType, "first message should be lessee response")
		require.Equal(t, types.MessageType_MESSAGE_TYPE_LESSOR_FOLLOWUP, threadComplaint.Thread[1].MessageType, "second message should be lessor followup")
		require.Equal(t, types.ComplaintState_STATE_AWAITING_RESPONSE, threadComplaint.State, "should be awaiting response")
	})

	t.Run("Complaint Closure", func(t *testing.T) {
		// Create complaint for closure testing
		complaint := types.Complaint{
			LeaseId:            leaseId + 3,
			SubnetId:           subnetId + 3,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId,
			AbuseType:          types.AbuseType_ABUSE_TYPE_DOS_DDOS,
			IncidentTime:       &incidentTime,
			SourceIp:           sourceIp + 3,
			Summary:            "DDoS attack complaint that will be resolved and closed",
			State:              types.ComplaintState_STATE_RESPONDED,
			SubmittedAt:        &blockTime,
			ResolutionDeadline: &deadline,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &blockTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Close the complaint
		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")

		closedTime := ctx.BlockTime().Add(time.Hour * 3)
		retrievedComplaint.State = types.ComplaintState_STATE_CLOSED
		retrievedComplaint.ClosedAt = &closedTime
		retrievedComplaint.LastActivityAt = &closedTime
		keeper.SetComplaint(ctx, retrievedComplaint)

		// Verify closure
		closedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find closed complaint")
		require.Equal(t, types.ComplaintState_STATE_CLOSED, closedComplaint.State, "should be in closed state")
		require.NotNil(t, closedComplaint.ClosedAt, "should have closed timestamp")
		require.Equal(t, closedTime, *closedComplaint.ClosedAt, "should have correct closed time")
	})

	t.Run("Complaint Removal", func(t *testing.T) {
		// Create complaint for removal testing
		complaint := types.Complaint{
			LeaseId:            leaseId + 4,
			SubnetId:           subnetId + 4,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId,
			AbuseType:          types.AbuseType_ABUSE_TYPE_BRUTEFORCE,
			IncidentTime:       &incidentTime,
			SourceIp:           sourceIp + 4,
			Summary:            "Brute force complaint that will be removed for testing",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &blockTime,
			ResolutionDeadline: &deadline,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &blockTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Verify complaint exists
		_, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint before removal")

		// Remove complaint
		keeper.RemoveComplaint(ctx, complaintId)

		// Verify complaint was removed
		_, found = keeper.GetComplaint(ctx, complaintId)
		require.False(t, found, "should not find complaint after removal")

		// Verify complaint is not in the list
		allComplaints := keeper.GetAllComplaint(ctx)
		for _, c := range allComplaints {
			require.NotEqual(t, complaintId, c.Id, "removed complaint should not appear in list")
		}
	})
}

// TestComplaintIndexingIntegration tests the indexing system for efficient queries
func TestComplaintIndexingIntegration(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	baseTime := ctx.BlockTime()
	deadlineTime := baseTime.Add(time.Hour)

	t.Run("Index by Subnet", func(t *testing.T) {
		subnetId1 := uint64(5001)
		subnetId2 := uint64(5002)

		// Create complaints for different subnets
		complaint1 := types.Complaint{
			LeaseId:            1,
			SubnetId:           subnetId1,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_SPAM,
			IncidentTime:       &baseTime,
			SourceIp:           0x08080808,
			Summary:            "Spam complaint for subnet indexing test case number one",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaint2 := types.Complaint{
			LeaseId:            2,
			SubnetId:           subnetId1, // Same subnet
			LessorOrgId:        1,
			LesseeOrgId:        2,
			AbuseType:          types.AbuseType_ABUSE_TYPE_MALWARE,
			IncidentTime:       &baseTime,
			SourceIp:           0x08080809,
			Summary:            "Malware complaint for subnet indexing test case number two",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaint3 := types.Complaint{
			LeaseId:            3,
			SubnetId:           subnetId2, // Different subnet
			LessorOrgId:        2,
			LesseeOrgId:        3,
			AbuseType:          types.AbuseType_ABUSE_TYPE_PHISHING_FRAUD,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080A,
			Summary:            "Phishing complaint for subnet indexing test case number three",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		id1 := keeper.AppendComplaint(ctx, complaint1)
		id2 := keeper.AppendComplaint(ctx, complaint2)
		id3 := keeper.AppendComplaint(ctx, complaint3)

		// Verify indexing by counting complaints per subnet
		allComplaints := keeper.GetAllComplaint(ctx)
		
		subnet1Count := 0
		subnet2Count := 0
		var subnet1IDs, subnet2IDs []uint64

		for _, complaint := range allComplaints {
			if complaint.SubnetId == subnetId1 {
				subnet1Count++
				subnet1IDs = append(subnet1IDs, complaint.Id)
			} else if complaint.SubnetId == subnetId2 {
				subnet2Count++
				subnet2IDs = append(subnet2IDs, complaint.Id)
			}
		}

		require.Equal(t, 2, subnet1Count, "should have 2 complaints for subnet 1")
		require.Equal(t, 1, subnet2Count, "should have 1 complaint for subnet 2")
		require.Contains(t, subnet1IDs, id1, "should contain complaint 1 for subnet 1")
		require.Contains(t, subnet1IDs, id2, "should contain complaint 2 for subnet 1")
		require.Contains(t, subnet2IDs, id3, "should contain complaint 3 for subnet 2")
	})

	t.Run("Index by Lease", func(t *testing.T) {
		leaseId1 := uint64(6001)
		leaseId2 := uint64(6002)

		// Create complaints for different leases
		complaint1 := types.Complaint{
			LeaseId:            leaseId1,
			SubnetId:           1,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_HACKING,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080B,
			Summary:            "Hacking complaint for lease indexing test case",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaint2 := types.Complaint{
			LeaseId:            leaseId2,
			SubnetId:           2,
			LessorOrgId:        2,
			LesseeOrgId:        2,
			AbuseType:          types.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080C,
			Summary:            "Copyright DMCA complaint for lease indexing test",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		id1 := keeper.AppendComplaint(ctx, complaint1)
		id2 := keeper.AppendComplaint(ctx, complaint2)

		// Verify lease indexing
		allComplaints := keeper.GetAllComplaint(ctx)
		foundLease1 := false
		foundLease2 := false

		for _, complaint := range allComplaints {
			if complaint.Id == id1 {
				require.Equal(t, leaseId1, complaint.LeaseId, "should have correct lease ID for complaint 1")
				foundLease1 = true
			}
			if complaint.Id == id2 {
				require.Equal(t, leaseId2, complaint.LeaseId, "should have correct lease ID for complaint 2")
				foundLease2 = true
			}
		}

		require.True(t, foundLease1, "should find complaint for lease 1")
		require.True(t, foundLease2, "should find complaint for lease 2")
	})

	t.Run("Index by Organization", func(t *testing.T) {
		lessorOrgId := uint64(7001)
		lesseeOrgId1 := uint64(7002)
		lesseeOrgId2 := uint64(7003)

		// Create complaints for different organizations
		complaint1 := types.Complaint{
			LeaseId:            1,
			SubnetId:           1,
			LessorOrgId:        lessorOrgId,
			LesseeOrgId:        lesseeOrgId1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_PORT_SCANNING,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080D,
			Summary:            "Port scanning complaint for organization indexing test",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaint2 := types.Complaint{
			LeaseId:            2,
			SubnetId:           2,
			LessorOrgId:        lessorOrgId, // Same lessor
			LesseeOrgId:        lesseeOrgId2, // Different lessee
			AbuseType:          types.AbuseType_ABUSE_TYPE_OTHER,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080E,
			Summary:            "Other abuse complaint for organization indexing test",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		id1 := keeper.AppendComplaint(ctx, complaint1)
		id2 := keeper.AppendComplaint(ctx, complaint2)

		// Verify organization indexing
		allComplaints := keeper.GetAllComplaint(ctx)
		
		lessorCount := 0
		lessee1Count := 0
		lessee2Count := 0

		for _, complaint := range allComplaints {
			if complaint.LessorOrgId == lessorOrgId {
				lessorCount++
			}
			if complaint.LesseeOrgId == lesseeOrgId1 {
				lessee1Count++
			}
			if complaint.LesseeOrgId == lesseeOrgId2 {
				lessee2Count++
			}
		}

		require.GreaterOrEqual(t, lessorCount, 2, "should have at least 2 complaints for lessor org")
		require.GreaterOrEqual(t, lessee1Count, 1, "should have at least 1 complaint for lessee org 1")
		require.GreaterOrEqual(t, lessee2Count, 1, "should have at least 1 complaint for lessee org 2")

		// Verify specific complaint IDs were created
		foundId1 := false
		foundId2 := false
		for _, complaint := range allComplaints {
			if complaint.Id == id1 {
				foundId1 = true
			}
			if complaint.Id == id2 {
				foundId2 = true
			}
		}
		require.True(t, foundId1, "should find complaint 1")
		require.True(t, foundId2, "should find complaint 2")
	})

	t.Run("Index by State", func(t *testing.T) {
		// Create complaints in different states
		submittedComplaint := types.Complaint{
			LeaseId:            1,
			SubnetId:           1,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_SPAM,
			IncidentTime:       &baseTime,
			SourceIp:           0x0808080F,
			Summary:            "Submitted state complaint for state indexing test",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		respondedComplaint := types.Complaint{
			LeaseId:            2,
			SubnetId:           2,
			LessorOrgId:        2,
			LesseeOrgId:        2,
			AbuseType:          types.AbuseType_ABUSE_TYPE_MALWARE,
			IncidentTime:       &baseTime,
			SourceIp:           0x08080810,
			Summary:            "Responded state complaint for state indexing test",
			State:              types.ComplaintState_STATE_RESPONDED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		closedTime := baseTime.Add(time.Hour * 2)
		closedComplaint := types.Complaint{
			LeaseId:            3,
			SubnetId:           3,
			LessorOrgId:        3,
			LesseeOrgId:        3,
			AbuseType:          types.AbuseType_ABUSE_TYPE_PHISHING_FRAUD,
			IncidentTime:       &baseTime,
			SourceIp:           0x08080811,
			Summary:            "Closed state complaint for state indexing test",
			State:              types.ComplaintState_STATE_CLOSED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			ClosedAt:           &closedTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &closedTime,
			ResolutionFailed:   false,
		}

		id1 := keeper.AppendComplaint(ctx, submittedComplaint)
		id2 := keeper.AppendComplaint(ctx, respondedComplaint)
		id3 := keeper.AppendComplaint(ctx, closedComplaint)

		// Verify state indexing
		allComplaints := keeper.GetAllComplaint(ctx)
		
		stateMap := make(map[types.ComplaintState][]uint64)
		for _, complaint := range allComplaints {
			stateMap[complaint.State] = append(stateMap[complaint.State], complaint.Id)
		}

		require.Contains(t, stateMap[types.ComplaintState_STATE_SUBMITTED], id1, "should find submitted complaint")
		require.Contains(t, stateMap[types.ComplaintState_STATE_RESPONDED], id2, "should find responded complaint")
		require.Contains(t, stateMap[types.ComplaintState_STATE_CLOSED], id3, "should find closed complaint")
	})
}

// TestComplaintValidationIntegration tests validation logic and constraints
func TestComplaintValidationIntegration(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()
	deadlineTime := baseTime.Add(time.Hour)

	t.Run("Abuse Type Validation", func(t *testing.T) {
		// Test all valid abuse types
		validAbuseTypes := []types.AbuseType{
			types.AbuseType_ABUSE_TYPE_COPYRIGHT_DMCA,
			types.AbuseType_ABUSE_TYPE_SPAM,
			types.AbuseType_ABUSE_TYPE_MALWARE,
			types.AbuseType_ABUSE_TYPE_PHISHING_FRAUD,
			types.AbuseType_ABUSE_TYPE_DOS_DDOS,
			types.AbuseType_ABUSE_TYPE_HACKING,
			types.AbuseType_ABUSE_TYPE_BRUTEFORCE,
			types.AbuseType_ABUSE_TYPE_PORT_SCANNING,
			types.AbuseType_ABUSE_TYPE_VIOLENCE_HATE_CSAM,
			types.AbuseType_ABUSE_TYPE_OTHER,
		}

		for i, abuseType := range validAbuseTypes {
			complaint := types.Complaint{
				LeaseId:            uint64(i + 1),
				SubnetId:           uint64(i + 1),
				LessorOrgId:        1,
				LesseeOrgId:        1,
				AbuseType:          abuseType,
				IncidentTime:       &baseTime,
				SourceIp:           uint64(0x08080800 + i),
				Summary:            "Valid abuse type complaint for validation testing purposes",
				State:              types.ComplaintState_STATE_SUBMITTED,
				SubmittedAt:        &baseTime,
				ResolutionDeadline: &deadlineTime,
				Thread:             []*types.ThreadMessage{},
				LastActivityAt:     &baseTime,
				ResolutionFailed:   false,
			}

			complaintId := keeper.AppendComplaint(ctx, complaint)
			require.Greater(t, complaintId, uint64(0), "should create complaint for abuse type %s", abuseType.String())

			retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
			require.True(t, found, "should find complaint for abuse type %s", abuseType.String())
			require.Equal(t, abuseType, retrievedComplaint.AbuseType, "should have correct abuse type %s", abuseType.String())
		}
	})

	t.Run("State Transition Validation", func(t *testing.T) {
		// Test all valid state transitions
		complaint := types.Complaint{
			LeaseId:            9001,
			SubnetId:           9001,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_SPAM,
			IncidentTime:       &baseTime,
			SourceIp:           0x09090909,
			Summary:            "State transition complaint for validation testing purposes",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &deadlineTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Test STATE_SUBMITTED → STATE_RESPONDED
		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")
		
		retrievedComplaint.State = types.ComplaintState_STATE_RESPONDED
		updatedActivityTime1 := baseTime.Add(time.Hour)
		retrievedComplaint.LastActivityAt = &updatedActivityTime1
		keeper.SetComplaint(ctx, retrievedComplaint)

		// Test STATE_RESPONDED → STATE_AWAITING_RESPONSE
		respondedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find responded complaint")
		require.Equal(t, types.ComplaintState_STATE_RESPONDED, respondedComplaint.State, "should be in responded state")

		respondedComplaint.State = types.ComplaintState_STATE_AWAITING_RESPONSE
		updatedActivityTime3 := baseTime.Add(time.Hour * 2)
		respondedComplaint.LastActivityAt = &updatedActivityTime3
		keeper.SetComplaint(ctx, respondedComplaint)

		// Test STATE_AWAITING_RESPONSE → STATE_CLOSED
		awaitingComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find awaiting complaint")
		require.Equal(t, types.ComplaintState_STATE_AWAITING_RESPONSE, awaitingComplaint.State, "should be awaiting response")

		awaitingComplaint.State = types.ComplaintState_STATE_CLOSED
		closedTime := baseTime.Add(time.Hour * 3)
		awaitingComplaint.ClosedAt = &closedTime
		awaitingComplaint.LastActivityAt = &closedTime
		keeper.SetComplaint(ctx, awaitingComplaint)

		// Verify final state
		finalComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find final complaint")
		require.Equal(t, types.ComplaintState_STATE_CLOSED, finalComplaint.State, "should be in closed state")
		require.NotNil(t, finalComplaint.ClosedAt, "should have closed timestamp")
	})

	t.Run("Timestamp Validation", func(t *testing.T) {
		// Test with various timestamps
		pastTime := baseTime.Add(-time.Hour * 24)
		futureTime := baseTime.Add(time.Hour * 24)

		// Valid incident time (in the past)
		validComplaint := types.Complaint{
			LeaseId:            9002,
			SubnetId:           9002,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_MALWARE,
			IncidentTime:       &pastTime,
			SourceIp:           0x0A0A0A0A,
			Summary:            "Valid timestamp complaint for validation testing purposes",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &futureTime,
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, validComplaint)
		require.Greater(t, complaintId, uint64(0), "should create complaint with valid timestamps")

		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")
		require.Equal(t, pastTime, *retrievedComplaint.IncidentTime, "should have correct incident time")
		require.Equal(t, baseTime, *retrievedComplaint.SubmittedAt, "should have correct submitted time")
		require.Equal(t, futureTime, *retrievedComplaint.ResolutionDeadline, "should have correct deadline")
	})

	t.Run("Resolution Failure Handling", func(t *testing.T) {
		// Test resolution failure scenarios
		shortDeadline := baseTime.Add(time.Minute)
		complaint := types.Complaint{
			LeaseId:            9003,
			SubnetId:           9003,
			LessorOrgId:        1,
			LesseeOrgId:        1,
			AbuseType:          types.AbuseType_ABUSE_TYPE_DOS_DDOS,
			IncidentTime:       &baseTime,
			SourceIp:           0x0B0B0B0B,
			Summary:            "Resolution failure complaint for validation testing purposes",
			State:              types.ComplaintState_STATE_SUBMITTED,
			SubmittedAt:        &baseTime,
			ResolutionDeadline: &shortDeadline, // Short deadline for testing
			Thread:             []*types.ThreadMessage{},
			LastActivityAt:     &baseTime,
			ResolutionFailed:   false,
		}

		complaintId := keeper.AppendComplaint(ctx, complaint)

		// Simulate deadline failure
		retrievedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find complaint")

		retrievedComplaint.ResolutionFailed = true
		updatedActivityTime2 := baseTime.Add(time.Hour)
		retrievedComplaint.LastActivityAt = &updatedActivityTime2
		keeper.SetComplaint(ctx, retrievedComplaint)

		// Verify failure state
		failedComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find failed complaint")
		require.True(t, failedComplaint.ResolutionFailed, "should be marked as resolution failed")

		// Test recovery from failure
		failedComplaint.ResolutionFailed = false
		failedComplaint.State = types.ComplaintState_STATE_CLOSED
		closedTime := baseTime.Add(time.Hour * 2)
		failedComplaint.ClosedAt = &closedTime
		failedComplaint.LastActivityAt = &closedTime
		keeper.SetComplaint(ctx, failedComplaint)

		// Verify recovery
		recoveredComplaint, found := keeper.GetComplaint(ctx, complaintId)
		require.True(t, found, "should find recovered complaint")
		require.False(t, recoveredComplaint.ResolutionFailed, "should no longer be marked as failed")
		require.Equal(t, types.ComplaintState_STATE_CLOSED, recoveredComplaint.State, "should be closed")
	})
}

// TestComplaintPerformanceAndScalability tests performance with many complaints
func TestComplaintPerformanceAndScalability(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	baseTime := ctx.BlockTime()
	deadlineTime := baseTime.Add(time.Hour)

	t.Run("Bulk Complaint Operations", func(t *testing.T) {
		// Create many complaints
		numComplaints := 100
		var complaintIDs []uint64

		for i := 0; i < numComplaints; i++ {
			incidentTime := baseTime.Add(time.Duration(-i) * time.Minute)
			bulkDeadline := baseTime.Add(time.Hour * 24)
			complaint := types.Complaint{
				LeaseId:            uint64(10000 + i),
				SubnetId:           uint64(10000 + i),
				LessorOrgId:        uint64(1 + (i % 10)), // Distribute across 10 lessor orgs
				LesseeOrgId:        uint64(1 + (i % 20)), // Distribute across 20 lessee orgs
				AbuseType:          types.AbuseType(1 + (i % 10)), // Cycle through abuse types
				IncidentTime:       &incidentTime,
				SourceIp:           uint64(0x08000000 + i),
				Summary:            "Bulk complaint for performance testing and scalability validation",
				State:              types.ComplaintState_STATE_SUBMITTED,
				SubmittedAt:        &baseTime,
				ResolutionDeadline: &bulkDeadline,
				Thread:             []*types.ThreadMessage{},
				LastActivityAt:     &baseTime,
				ResolutionFailed:   false,
			}

			complaintId := keeper.AppendComplaint(ctx, complaint)
			complaintIDs = append(complaintIDs, complaintId)
		}

		require.Len(t, complaintIDs, numComplaints, "should create all complaints")

		// Test bulk retrieval
		allComplaints := keeper.GetAllComplaint(ctx)
		require.GreaterOrEqual(t, len(allComplaints), numComplaints, "should retrieve at least created complaints")

		// Test individual lookups
		for i := 0; i < 10; i++ { // Test first 10
			complaint, found := keeper.GetComplaint(ctx, complaintIDs[i])
			require.True(t, found, "should find complaint %d", i)
			require.Equal(t, complaintIDs[i], complaint.Id, "should have correct ID")
		}

		// Test removal of some complaints
		for i := 0; i < 10; i++ {
			keeper.RemoveComplaint(ctx, complaintIDs[i])
		}

		// Verify remaining complaints
		remainingComplaints := keeper.GetAllComplaint(ctx)
		removedIDs := make(map[uint64]bool)
		for i := 0; i < 10; i++ {
			removedIDs[complaintIDs[i]] = true
		}

		for _, complaint := range remainingComplaints {
			require.False(t, removedIDs[complaint.Id], "removed complaint should not appear in list")
		}
	})

	t.Run("Index Performance", func(t *testing.T) {
		// Create complaints with specific patterns for indexing tests
		subnets := []uint64{20001, 20002, 20003, 20004, 20005}
		leases := []uint64{20001, 20002, 20003, 20004, 20005}
		lessorOrgs := []uint64{20001, 20002, 20003}
		lesseeOrgs := []uint64{20004, 20005, 20006}

		var complaintsPerSubnet = make(map[uint64][]uint64)
		var complaintsPerLease = make(map[uint64][]uint64)
		var complaintsPerLessor = make(map[uint64][]uint64)
		var complaintsPerLessee = make(map[uint64][]uint64)

		for i := 0; i < 50; i++ {
			subnetId := subnets[i%len(subnets)]
			leaseId := leases[i%len(leases)]
			lessorOrgId := lessorOrgs[i%len(lessorOrgs)]
			lesseeOrgId := lesseeOrgs[i%len(lesseeOrgs)]

			complaint := types.Complaint{
				LeaseId:            leaseId,
				SubnetId:           subnetId,
				LessorOrgId:        lessorOrgId,
				LesseeOrgId:        lesseeOrgId,
				AbuseType:          types.AbuseType_ABUSE_TYPE_SPAM,
				IncidentTime:       &baseTime,
				SourceIp:           uint64(0x0C000000 + i),
				Summary:            "Index performance complaint for testing and validation",
				State:              types.ComplaintState_STATE_SUBMITTED,
				SubmittedAt:        &baseTime,
				ResolutionDeadline: &deadlineTime,
				Thread:             []*types.ThreadMessage{},
				LastActivityAt:     &baseTime,
				ResolutionFailed:   false,
			}

			complaintId := keeper.AppendComplaint(ctx, complaint)
			
			complaintsPerSubnet[subnetId] = append(complaintsPerSubnet[subnetId], complaintId)
			complaintsPerLease[leaseId] = append(complaintsPerLease[leaseId], complaintId)
			complaintsPerLessor[lessorOrgId] = append(complaintsPerLessor[lessorOrgId], complaintId)
			complaintsPerLessee[lesseeOrgId] = append(complaintsPerLessee[lesseeOrgId], complaintId)
		}

		// Verify indexing by counting complaints per entity
		allComplaints := keeper.GetAllComplaint(ctx)
		
		actualSubnetCounts := make(map[uint64]int)
		actualLeaseCounts := make(map[uint64]int)
		actualLessorCounts := make(map[uint64]int)
		actualLesseeCounts := make(map[uint64]int)

		for _, complaint := range allComplaints {
			// Only count the complaints we just created (with specific subnet IDs)
			for _, subnetId := range subnets {
				if complaint.SubnetId == subnetId {
					actualSubnetCounts[complaint.SubnetId]++
					actualLeaseCounts[complaint.LeaseId]++
					actualLessorCounts[complaint.LessorOrgId]++
					actualLesseeCounts[complaint.LesseeOrgId]++
					break
				}
			}
		}

		// Each subnet should have 10 complaints (50 total / 5 subnets)
		for _, subnetId := range subnets {
			require.Equal(t, 10, actualSubnetCounts[subnetId], "should have 10 complaints per subnet")
		}

		// Each lease should have 10 complaints (50 total / 5 leases)
		for _, leaseId := range leases {
			require.Equal(t, 10, actualLeaseCounts[leaseId], "should have 10 complaints per lease")
		}

		// Verify organization distribution
		totalLessorComplaints := 0
		for _, count := range actualLessorCounts {
			totalLessorComplaints += count
		}
		require.Equal(t, 50, totalLessorComplaints, "should have 50 total lessor complaints")

		totalLesseeComplaints := 0
		for _, count := range actualLesseeCounts {
			totalLesseeComplaints += count
		}
		require.Equal(t, 50, totalLesseeComplaints, "should have 50 total lessee complaints")
	})
}