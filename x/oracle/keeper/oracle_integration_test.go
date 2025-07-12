package keeper_test

import (
	"encoding/json"
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/sample"
	"ipex/x/oracle/abci"
	"ipex/x/oracle/keeper"
	"ipex/x/oracle/types"
	leasetypes "ipex/x/lease/types"
	orgtypes "ipex/x/org/types"
	subnettypes "ipex/x/subnet/types"
	trusttypes "ipex/x/trust/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestOracleVoteExtensionLifecycle tests the complete oracle vote extension workflow
func TestOracleVoteExtensionLifecycle(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(1000)

	// Set last block time to enable oracle operations
	oracleKeeper.SetLastBlockTime(ctx, ctx.BlockTime().Add(-time.Hour))

	t.Run("Empty Vote Extension Generation", func(t *testing.T) {
		// Create mock vote extension handler
		mockOrgKeeper := &mockOrgKeeper{orgs: []orgtypes.Org{}}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: []subnettypes.Subnet{}}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: []uint64{}}
		mockNetIntel := &mockNetIntelService{ready: true}

		handler := abci.NewVoteExtHandler(
			ctx.Logger(),
			oracleKeeper,
			mockOrgKeeper,
			mockSubnetKeeper,
			mockLeaseKeeper,
			mockNetIntel,
			mockTrustKeeper,
		)

		// Test vote extension with no data to validate
		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should generate empty vote extension")
		require.Empty(t, resp.VoteExtension, "should return empty vote extension when no validation needed")
	})

	t.Run("Organization Domain Validation Extension", func(t *testing.T) {
		// Create organizations that need validation
		testOrgs := []orgtypes.Org{
			{
				Id:                  1001,
				Website:             "https://example1.com",
				DomainVerified:      false,
				DomainLastChecked:   ctx.BlockTime().Add(-2 * time.Hour),
			},
			{
				Id:                  1002,
				Website:             "https://example2.com",
				DomainVerified:      true,
				DomainLastChecked:   ctx.BlockTime().Add(-30 * time.Minute), // Recent check
			},
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: testOrgs}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: []subnettypes.Subnet{}}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: []uint64{}}
		mockNetIntel := &mockNetIntelService{ready: true}

		handler := abci.NewVoteExtHandler(
			ctx.Logger(),
			oracleKeeper,
			mockOrgKeeper,
			mockSubnetKeeper,
			mockLeaseKeeper,
			mockNetIntel,
			mockTrustKeeper,
		)

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should generate vote extension with org validations")
		require.NotEmpty(t, resp.VoteExtension, "should have vote extension data")

		// Parse the vote extension
		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse vote extension")
		require.Equal(t, req.Height, voteExt.Height, "should have correct height")
		require.Contains(t, voteExt.OrgDomainValidation, uint64(1001), "should validate org that needs checking")
		require.NotContains(t, voteExt.OrgDomainValidation, uint64(1002), "should skip recently validated org")
	})

	t.Run("Subnet Validation Extensions", func(t *testing.T) {
		// Create subnets that need various validations
		testSubnets := []subnettypes.Subnet{
			{
				Id:             2001,
				NetworkAddress: 3232235521, // 192.168.1.1
				Mask:           24,
				OriginAsn:      64512,
				Country:        "US",
				OwnerOrgId:     1001,
				RirSyncTime:    ctx.BlockTime().Add(-2 * time.Hour),
				RpkiLastCheck:  ctx.BlockTime().Add(-2 * time.Hour),
				GeoLastCheck:   ctx.BlockTime().Add(-2 * time.Hour),
			},
			{
				Id:             2002,
				NetworkAddress: 3232235776, // 192.168.2.0
				Mask:           24,
				OriginAsn:      64513,
				Country:        "CA",
				OwnerOrgId:     1002,
				RirSyncTime:    ctx.BlockTime().Add(-30 * time.Minute), // Recent sync
			},
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: []orgtypes.Org{}}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: testSubnets}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: []uint64{2001}}
		mockNetIntel := &mockNetIntelService{
			ready: true,
			reputationResults: map[string]map[int]int{
				"192.168.1.0/24": {1: 5, 2: 3, 3: 1, 4: 0},
			},
			geoResults: map[string]string{
				"192.168.1.2": "US",
			},
		}

		handler := abci.NewVoteExtHandler(
			ctx.Logger(),
			oracleKeeper,
			mockOrgKeeper,
			mockSubnetKeeper,
			mockLeaseKeeper,
			mockNetIntel,
			mockTrustKeeper,
		)

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should generate vote extension with subnet validations")
		require.NotEmpty(t, resp.VoteExtension, "should have vote extension data")

		// Parse the vote extension
		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse vote extension")

		// Verify subnet validations are included
		require.Contains(t, voteExt.SubnetRirSync, uint64(2001), "should sync subnet that needs RIR update")
		require.Contains(t, voteExt.RpkiValidation, uint64(2001), "should validate RPKI for subnet")
		require.Contains(t, voteExt.SubnetBlacklistSync, uint64(2001), "should sync blacklist for subnet")
		require.NotContains(t, voteExt.SubnetRirSync, uint64(2002), "should skip recently synced subnet")
	})
}

// TestOracleConsensusAggregation tests the vote aggregation and consensus mechanism
func TestOracleConsensusAggregation(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(1500)

	// Create mock keepers
	mockOrgKeeper := &mockOrgKeeper{}
	mockSubnetKeeper := &mockSubnetKeeper{}
	mockLeaseKeeper := &mockLeaseKeeper{}
	mockTrustKeeper := &mockTrustKeeper{}

	proposalHandler := abci.NewProposalHandler(
		ctx.Logger(),
		mockOrgKeeper,
		mockSubnetKeeper,
		mockLeaseKeeper,
		mockTrustKeeper,
		&mockValidatorStore{},
	)

	t.Run("Organization Domain Validation Consensus", func(t *testing.T) {
		// Create vote extensions from multiple validators
		validators := []abcitypes.Validator{
			{Address: []byte("validator1"), Power: 100},
			{Address: []byte("validator2"), Power: 200},
			{Address: []byte("validator3"), Power: 300},
		}

		votes := []abcitypes.ExtendedVoteInfo{
			{
				Validator: validators[0],
				BlockIdFlag: 2, // BlockIDFlagCommit
				VoteExtension: createOrgValidationVoteExt(t, ctx.BlockHeight(), map[uint64]bool{
					1001: true,  // Valid
					1002: false, // Invalid
				}),
			},
			{
				Validator: validators[1],
				BlockIdFlag: 2,
				VoteExtension: createOrgValidationVoteExt(t, ctx.BlockHeight(), map[uint64]bool{
					1001: true,  // Valid (consensus: 300 power for valid)
					1002: false, // Invalid
				}),
			},
			{
				Validator: validators[2],
				BlockIdFlag: 2,
				VoteExtension: createOrgValidationVoteExt(t, ctx.BlockHeight(), map[uint64]bool{
					1001: false, // Invalid (total: 300 valid vs 300 invalid = tie, should be false)
					1002: true,  // Valid (total: 300 valid vs 300 invalid = tie, should be false)
				}),
			},
		}

		commitInfo := abcitypes.ExtendedCommitInfo{Votes: votes}

		req := &abcitypes.RequestPrepareProposal{
			Height: ctx.BlockHeight(),
			LocalLastCommit: commitInfo,
		}

		resp, err := proposalHandler.PrepareProposal()(ctx, req)
		require.NoError(t, err, "should prepare proposal with aggregated results")
		require.NotEmpty(t, resp.Txs, "should have injected validation transaction")

		// Parse the injected transaction
		var validationResults abci.NetworkValidationResults
		err = json.Unmarshal(resp.Txs[0], &validationResults)
		require.NoError(t, err, "should parse validation results")

		// Verify consensus results (majority voting power required)
		require.Contains(t, validationResults.OrgValidationResults, uint64(1001))
		require.Contains(t, validationResults.OrgValidationResults, uint64(1002))
		require.False(t, validationResults.OrgValidationResults[1001], "org 1001 should be invalid (tie)")
		require.False(t, validationResults.OrgValidationResults[1002], "org 1002 should be invalid (tie)")
	})

	t.Run("Subnet RIR Sync Consensus", func(t *testing.T) {
		validators := []abcitypes.Validator{
			{Address: []byte("validator1"), Power: 300},
			{Address: []byte("validator2"), Power: 200},
			{Address: []byte("validator3"), Power: 100},
		}

		subnetInfo1 := subnettypes.SubnetInfo{
			Valid:          true,
			OrgID:          "ORG-12345",
			OrgName:        "Example Corp",
			AbuseEmail:     "abuse@example.com",
			AllocationType: "ALLOCATED",
			Rir:            "ARIN",
		}

		subnetInfo2 := subnettypes.SubnetInfo{
			Valid:          true,
			OrgID:          "ORG-67890",
			OrgName:        "Test Inc",
			AbuseEmail:     "abuse@test.com",
			AllocationType: "ALLOCATED",
			Rir:            "RIPE",
		}

		votes := []abcitypes.ExtendedVoteInfo{
			{
				Validator: validators[0],
				BlockIdFlag: 2,
				VoteExtension: createSubnetSyncVoteExt(t, ctx.BlockHeight(), map[uint64]subnettypes.SubnetInfo{
					2001: subnetInfo1, // Highest power validator
				}),
			},
			{
				Validator: validators[1],
				BlockIdFlag: 2,
				VoteExtension: createSubnetSyncVoteExt(t, ctx.BlockHeight(), map[uint64]subnettypes.SubnetInfo{
					2001: subnetInfo2, // Different info, lower power
				}),
			},
			{
				Validator: validators[2],
				BlockIdFlag: 2,
				VoteExtension: createSubnetSyncVoteExt(t, ctx.BlockHeight(), map[uint64]subnettypes.SubnetInfo{
					2001: subnetInfo1, // Same as validator1
				}),
			},
		}

		commitInfo := abcitypes.ExtendedCommitInfo{Votes: votes}

		req := &abcitypes.RequestPrepareProposal{
			Height: ctx.BlockHeight(),
			LocalLastCommit: commitInfo,
		}

		resp, err := proposalHandler.PrepareProposal()(ctx, req)
		require.NoError(t, err, "should prepare proposal")

		var validationResults abci.NetworkValidationResults
		err = json.Unmarshal(resp.Txs[0], &validationResults)
		require.NoError(t, err, "should parse results")

		// Verify subnet info consensus (subnetInfo1 should win with 400 power vs 200)
		require.Contains(t, validationResults.SubnetSyncResults, uint64(2001))
		result := validationResults.SubnetSyncResults[2001]
		require.Equal(t, subnetInfo1.OrgID, result.OrgID, "should use info from highest voting power")
		require.Equal(t, subnetInfo1.Rir, result.Rir, "should use RIR from consensus")
	})

	t.Run("Blacklist Reputation Consensus", func(t *testing.T) {
		validators := []abcitypes.Validator{
			{Address: []byte("validator1"), Power: 250},
			{Address: []byte("validator2"), Power: 250},
			{Address: []byte("validator3"), Power: 100},
		}

		blacklist1 := trusttypes.SubnetBlacklistCount{
			Level1Count: 10,
			Level2Count: 5,
			Level3Count: 2,
			Level4Count: 0,
		}

		blacklist2 := trusttypes.SubnetBlacklistCount{
			Level1Count: 8,
			Level2Count: 3,
			Level3Count: 1,
			Level4Count: 0,
		}

		votes := []abcitypes.ExtendedVoteInfo{
			{
				Validator: validators[0],
				BlockIdFlag: 2,
				VoteExtension: createBlacklistSyncVoteExt(t, ctx.BlockHeight(), map[uint64]trusttypes.SubnetBlacklistCount{
					2001: blacklist1,
				}),
			},
			{
				Validator: validators[1],
				BlockIdFlag: 2,
				VoteExtension: createBlacklistSyncVoteExt(t, ctx.BlockHeight(), map[uint64]trusttypes.SubnetBlacklistCount{
					2001: blacklist1, // Same as validator1
				}),
			},
			{
				Validator: validators[2],
				BlockIdFlag: 2,
				VoteExtension: createBlacklistSyncVoteExt(t, ctx.BlockHeight(), map[uint64]trusttypes.SubnetBlacklistCount{
					2001: blacklist2, // Different data, lower power
				}),
			},
		}

		commitInfo := abcitypes.ExtendedCommitInfo{Votes: votes}

		req := &abcitypes.RequestPrepareProposal{
			Height: ctx.BlockHeight(),
			LocalLastCommit: commitInfo,
		}

		resp, err := proposalHandler.PrepareProposal()(ctx, req)
		require.NoError(t, err, "should prepare proposal")

		var validationResults abci.NetworkValidationResults
		err = json.Unmarshal(resp.Txs[0], &validationResults)
		require.NoError(t, err, "should parse results")

		// Verify blacklist consensus (blacklist1 should win with 500 power vs 100)
		require.Contains(t, validationResults.SubnetBlacklistSyncResults, uint64(2001))
		result := validationResults.SubnetBlacklistSyncResults[2001]
		require.Equal(t, blacklist1.Level1Count, result.Level1Count, "should use blacklist from consensus")
		require.Equal(t, blacklist1.Level2Count, result.Level2Count, "should use blacklist from consensus")
	})
}

// TestOracleVoteExtensionVerification tests vote extension verification logic
func TestOracleVoteExtensionVerification(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(2000)

	// Create test organizations and subnets in state
	mockOrgKeeper := &mockOrgKeeper{
		orgs: []orgtypes.Org{
			{Id: 1001, Website: "https://example.com"},
			{Id: 1002, Website: "https://test.com"},
		},
	}
	mockSubnetKeeper := &mockSubnetKeeper{
		subnets: []subnettypes.Subnet{
			{Id: 2001, NetworkAddress: 3232235521},
			{Id: 2002, NetworkAddress: 3232235776},
		},
	}
	mockLeaseKeeper := &mockLeaseKeeper{}
	mockTrustKeeper := &mockTrustKeeper{}
	mockNetIntel := &mockNetIntelService{ready: true}

	handler := abci.NewVoteExtHandler(
		ctx.Logger(),
		oracleKeeper,
		mockOrgKeeper,
		mockSubnetKeeper,
		mockLeaseKeeper,
		mockNetIntel,
		mockTrustKeeper,
	)

	t.Run("Valid Vote Extension Verification", func(t *testing.T) {
		voteExt := abci.OracleVoteExtension{
			Height: ctx.BlockHeight(),
			OrgDomainValidation: map[uint64]bool{
				1001: true,
				1002: false,
			},
			SubnetRirSync: map[uint64]subnettypes.SubnetInfo{
				2001: {Valid: true, OrgID: "ORG-123"},
			},
			SubnetBlacklistSync: map[uint64]trusttypes.SubnetBlacklistCount{
				2001: {Level1Count: 5},
			},
		}

		voteExtBytes, err := json.Marshal(voteExt)
		require.NoError(t, err, "should marshal vote extension")

		req := &abcitypes.RequestVerifyVoteExtension{
			Height:        ctx.BlockHeight(),
			VoteExtension: voteExtBytes,
		}

		resp, err := handler.VerifyVoteExtensionHandler()(ctx, req)
		require.NoError(t, err, "should verify valid vote extension")
		require.Equal(t, abcitypes.ResponseVerifyVoteExtension_ACCEPT, resp.Status, "should accept valid vote extension")
	})

	t.Run("Invalid Height Verification", func(t *testing.T) {
		voteExt := abci.OracleVoteExtension{
			Height: ctx.BlockHeight() + 100, // Wrong height
			OrgDomainValidation: map[uint64]bool{1001: true},
		}

		voteExtBytes, err := json.Marshal(voteExt)
		require.NoError(t, err, "should marshal vote extension")

		req := &abcitypes.RequestVerifyVoteExtension{
			Height:        ctx.BlockHeight(),
			VoteExtension: voteExtBytes,
		}

		_, err = handler.VerifyVoteExtensionHandler()(ctx, req)
		require.Error(t, err, "should reject vote extension with wrong height")
		require.Contains(t, err.Error(), "height does not match", "should have height mismatch error")
	})

	t.Run("Invalid Organization ID Verification", func(t *testing.T) {
		voteExt := abci.OracleVoteExtension{
			Height: ctx.BlockHeight(),
			OrgDomainValidation: map[uint64]bool{
				9999: true, // Non-existent org
			},
		}

		voteExtBytes, err := json.Marshal(voteExt)
		require.NoError(t, err, "should marshal vote extension")

		req := &abcitypes.RequestVerifyVoteExtension{
			Height:        ctx.BlockHeight(),
			VoteExtension: voteExtBytes,
		}

		_, err = handler.VerifyVoteExtensionHandler()(ctx, req)
		require.Error(t, err, "should reject vote extension with invalid org ID")
		require.Contains(t, err.Error(), "does not exist", "should have org not found error")
	})

	t.Run("Invalid Subnet ID Verification", func(t *testing.T) {
		voteExt := abci.OracleVoteExtension{
			Height: ctx.BlockHeight(),
			SubnetRirSync: map[uint64]subnettypes.SubnetInfo{
				9999: {Valid: true}, // Non-existent subnet
			},
		}

		voteExtBytes, err := json.Marshal(voteExt)
		require.NoError(t, err, "should marshal vote extension")

		req := &abcitypes.RequestVerifyVoteExtension{
			Height:        ctx.BlockHeight(),
			VoteExtension: voteExtBytes,
		}

		_, err = handler.VerifyVoteExtensionHandler()(ctx, req)
		require.Error(t, err, "should reject vote extension with invalid subnet ID")
		require.Contains(t, err.Error(), "does not exist", "should have subnet not found error")
	})

	t.Run("Empty Vote Extension Acceptance", func(t *testing.T) {
		req := &abcitypes.RequestVerifyVoteExtension{
			Height:        ctx.BlockHeight(),
			VoteExtension: []byte{}, // Empty extension
		}

		resp, err := handler.VerifyVoteExtensionHandler()(ctx, req)
		require.NoError(t, err, "should accept empty vote extension")
		require.Equal(t, abcitypes.ResponseVerifyVoteExtension_ACCEPT, resp.Status, "should accept empty vote extension")
	})
}

// TestOracleBlockTimeTracking tests the oracle's block time tracking functionality
func TestOracleBlockTimeTracking(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)

	t.Run("Set and Get Last Block Time", func(t *testing.T) {
		testTime := time.Now().UTC()
		ctx = ctx.WithBlockTime(testTime).WithBlockHeight(500)

		// Initially no block time should be stored
		_, found := oracleKeeper.GetLastBlockTime(ctx)
		require.False(t, found, "should not find initial block time")

		// Set block time
		oracleKeeper.SetLastBlockTime(ctx, testTime)

		// Retrieve block time
		retrievedTime, found := oracleKeeper.GetLastBlockTime(ctx)
		require.True(t, found, "should find stored block time")
		require.Equal(t, testTime.Unix(), retrievedTime.Unix(), "should retrieve correct block time")
	})

	t.Run("Block Time Update Workflow", func(t *testing.T) {
		initialTime := time.Now().UTC()
		laterTime := initialTime.Add(time.Hour)

		// Set initial time
		oracleKeeper.SetLastBlockTime(ctx, initialTime)

		// Update to later time
		oracleKeeper.SetLastBlockTime(ctx, laterTime)

		// Verify update
		retrievedTime, found := oracleKeeper.GetLastBlockTime(ctx)
		require.True(t, found, "should find updated block time")
		require.Equal(t, laterTime.Unix(), retrievedTime.Unix(), "should have updated time")
		require.True(t, retrievedTime.After(initialTime), "should be later than initial time")
	})
}

// Helper functions for creating vote extensions

func createOrgValidationVoteExt(t *testing.T, height int64, validations map[uint64]bool) []byte {
	voteExt := abci.OracleVoteExtension{
		Height:              height,
		OrgDomainValidation: validations,
	}
	bytes, err := json.Marshal(voteExt)
	require.NoError(t, err)
	return bytes
}

func createSubnetSyncVoteExt(t *testing.T, height int64, syncResults map[uint64]subnettypes.SubnetInfo) []byte {
	voteExt := abci.OracleVoteExtension{
		Height:        height,
		SubnetRirSync: syncResults,
	}
	bytes, err := json.Marshal(voteExt)
	require.NoError(t, err)
	return bytes
}

func createBlacklistSyncVoteExt(t *testing.T, height int64, blacklistResults map[uint64]trusttypes.SubnetBlacklistCount) []byte {
	voteExt := abci.OracleVoteExtension{
		Height:              height,
		SubnetBlacklistSync: blacklistResults,
	}
	bytes, err := json.Marshal(voteExt)
	require.NoError(t, err)
	return bytes
}

// Mock implementations

type mockOrgKeeper struct {
	orgs []orgtypes.Org
}

func (m *mockOrgKeeper) GetOrg(ctx sdk.Context, id uint64) (orgtypes.Org, bool) {
	for _, org := range m.orgs {
		if org.Id == id {
			return org, true
		}
	}
	return orgtypes.Org{}, false
}

func (m *mockOrgKeeper) GetOrgsForDomainValidation(ctx sdk.Context, referenceTime time.Time) []orgtypes.Org {
	var result []orgtypes.Org
	for _, org := range m.orgs {
		// Return orgs that need validation (last checked > 1 hour ago)
		if org.DomainLastChecked.Before(referenceTime.Add(-time.Hour)) {
			result = append(result, org)
		}
	}
	return result
}

func (m *mockOrgKeeper) UpdateOrgDomainValidationStatus(ctx sdk.Context, orgID uint64, isValid bool) error {
	return nil
}

func (m *mockOrgKeeper) HasOrg(ctx sdk.Context, id uint64) bool {
	_, found := m.GetOrg(ctx, id)
	return found
}

func (m *mockOrgKeeper) SetOrg(ctx sdk.Context, org orgtypes.Org) {}

type mockSubnetKeeper struct {
	subnets []subnettypes.Subnet
}

func (m *mockSubnetKeeper) GetSubnet(ctx sdk.Context, id uint64) (subnettypes.Subnet, bool) {
	for _, subnet := range m.subnets {
		if subnet.Id == id {
			return subnet, true
		}
	}
	return subnettypes.Subnet{}, false
}

func (m *mockSubnetKeeper) GetSubnetsForRirSync(ctx sdk.Context, referenceTime time.Time) []subnettypes.Subnet {
	var result []subnettypes.Subnet
	for _, subnet := range m.subnets {
		if subnet.RirSyncTime.Before(referenceTime.Add(-time.Hour)) {
			result = append(result, subnet)
		}
	}
	return result
}

func (m *mockSubnetKeeper) GetSubnetsForRPKIValidation(ctx sdk.Context, referenceTime time.Time) []subnettypes.Subnet {
	var result []subnettypes.Subnet
	for _, subnet := range m.subnets {
		if subnet.RpkiLastCheck.Before(referenceTime.Add(-time.Hour)) {
			result = append(result, subnet)
		}
	}
	return result
}

func (m *mockSubnetKeeper) GetSubnetsForGeoValidation(ctx sdk.Context, referenceTime time.Time) []subnettypes.Subnet {
	var result []subnettypes.Subnet
	for _, subnet := range m.subnets {
		if subnet.GeoLastCheck.Before(referenceTime.Add(-time.Hour)) {
			result = append(result, subnet)
		}
	}
	return result
}

func (m *mockSubnetKeeper) GetSubnetsForOwnershipVerification(ctx sdk.Context, referenceTime time.Time) []subnettypes.Subnet {
	return []subnettypes.Subnet{}
}

func (m *mockSubnetKeeper) CheckSubnetLeaseAvailability(ctx sdk.Context, subnetID uint64) bool {
	return true
}

func (m *mockSubnetKeeper) SetSubnet(ctx sdk.Context, subnet subnettypes.Subnet) {}

func (m *mockSubnetKeeper) UpdateSubnetRirDetails(ctx sdk.Context, subnetID uint64, info subnettypes.SubnetInfo) error {
	return nil
}

func (m *mockSubnetKeeper) UpdateSubnetRpkiValidationStatus(ctx sdk.Context, subnetID uint64, isValid bool) error {
	return nil
}

func (m *mockSubnetKeeper) UpdateSubnetGeoValidationStatus(ctx sdk.Context, subnetID uint64, isValid bool) error {
	return nil
}

func (m *mockSubnetKeeper) UpdateSubnetOwnershipVerification(ctx sdk.Context, subnetId uint64, isValid bool) error {
	return nil
}

func (m *mockSubnetKeeper) IsRirRecordValid(ctx sdk.Context, subnetId uint64) bool {
	return true
}

type mockLeaseKeeper struct{}

func (m *mockLeaseKeeper) GetContract(ctx sdk.Context, id uint64) (leasetypes.Contract, bool) {
	return leasetypes.Contract{}, false
}

func (m *mockLeaseKeeper) SetContract(ctx sdk.Context, contract leasetypes.Contract) {}

func (m *mockLeaseKeeper) GetFirstLeasedContractBySubnet(ctx sdk.Context, subnetId uint64) (uint64, bool) {
	return 0, false
}

type mockTrustKeeper struct {
	subnets []uint64
}

func (m *mockTrustKeeper) GetSubnetsForReputationScanning(ctx sdk.Context, referenceTime time.Time) []uint64 {
	return m.subnets
}

func (m *mockTrustKeeper) UpdateSubnetReputation(ctx sdk.Context, subnetId uint64, leaseId uint64, lessorOrgId uint64, lesseeOrgId uint64, reputation trusttypes.SubnetBlacklistCount) error {
	return nil
}

type mockNetIntelService struct {
	ready             bool
	reputationResults map[string]map[int]int // cidr -> level -> count
	geoResults        map[string]string      // ip -> country
}

func (m *mockNetIntelService) IsReady() bool {
	return m.ready
}

func (m *mockNetIntelService) GetLastUpdate() time.Time {
	return time.Now()
}

func (m *mockNetIntelService) GetLoadedLevels() []int {
	return []int{1, 2, 3, 4}
}

func (m *mockNetIntelService) GetLoadedCountries() []string {
	return []string{"US", "CA", "GB"}
}

func (m *mockNetIntelService) CheckReputation(ip string, level int) int {
	if results, exists := m.reputationResults[ip]; exists {
		if count, exists := results[level]; exists {
			return count
		}
	}
	return 0
}

func (m *mockNetIntelService) ScanSubnetReputation(subnet string, level int) int {
	return m.CheckReputation(subnet, level)
}

func (m *mockNetIntelService) CheckCountry(ip string, countryCode string) bool {
	if country, exists := m.geoResults[ip]; exists {
		return country == countryCode
	}
	return false
}

type mockValidatorStore struct{}

func (m *mockValidatorStore) GetPubKeyByConsAddr(sdk.Context, sdk.ConsAddress) (crypto.PubKey, error) {
	return nil, nil
}