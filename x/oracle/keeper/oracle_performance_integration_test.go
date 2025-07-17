package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/x/oracle/abci"
	orgtypes "ipex/x/org/types"
	subnettypes "ipex/x/subnet/types"
	trusttypes "ipex/x/trust/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
)

// TestOraclePerformanceAndScalability tests oracle performance with large datasets
func TestOraclePerformanceAndScalability(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(5000)
	
	// Set up consensus parameters to avoid nil pointer dereference
	consParams := cmtproto.ConsensusParams{
		Abci: &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: 1, // Enable vote extensions from height 1
		},
	}
	ctx = ctx.WithConsensusParams(consParams)

	baseTime := ctx.BlockTime()
	oracleKeeper.SetLastBlockTime(ctx, baseTime.Add(-2*time.Hour))

	t.Run("Large Organization Validation Dataset", func(t *testing.T) {
		// Create large number of organizations needing validation
		numOrgs := 100
		var orgs []orgtypes.Org

		for i := 0; i < numOrgs; i++ {
			org := orgtypes.Org{
				Id:                uint64(1000 + i),
				Name:              fmt.Sprintf("Organization %d", i),
				Website:           fmt.Sprintf("https://example%d.com", i),
				DomainVerified:    false,
				DomainVerifAttemptedAt: &[]time.Time{baseTime.Add(-time.Duration(3+i%12) * time.Hour)}[0], // Vary check times
			}
			orgs = append(orgs, org)
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: orgs}
		mockSubnetKeeper := &mockSubnetKeeper{}
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

		startTime := time.Now()

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should handle large organization dataset")

		processingTime := time.Since(startTime)
		t.Logf("Processed %d organizations in %v", numOrgs, processingTime)

		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse vote extension")

		// Verify all eligible organizations are included
		require.Greater(t, len(voteExt.OrgDomainValidation), 80, "should validate most organizations")
		require.LessOrEqual(t, len(voteExt.OrgDomainValidation), numOrgs, "should not exceed total organizations")

		// Performance assertion: should complete within reasonable time
		require.Less(t, processingTime, 30*time.Second, "should complete organization validation within 30 seconds")
	})

	t.Run("Large Subnet Validation Dataset", func(t *testing.T) {
		// Create large number of subnets needing various validations
		numSubnets := 200
		var subnets []subnettypes.Subnet
		var subnetIds []uint64

		baseIP := uint64(3232235520) // Start at 192.168.1.0

		for i := 0; i < numSubnets; i++ {
			subnetId := uint64(2000 + i)
			subnet := subnettypes.Subnet{
				Id:             subnetId,
				NetworkAddress: baseIP + uint64(i*256), // Increment by /24 networks
				Mask:           24,
				OriginAsn:      uint64(64512 + i%100), // Vary ASNs
				Country:        []string{"US", "CA", "GB", "DE", "FR"}[i%5],
				OwnerOrgId:     uint64(1000 + i%50), // Distribute across 50 orgs
				RirSyncedAt:      &[]time.Time{baseTime.Add(-time.Duration(2+i%48) * time.Hour)}[0],
				RpkiLastSyncedAt: &[]time.Time{baseTime.Add(-time.Duration(2+i%72) * time.Hour)}[0],
				GeoVerifAttemptedAt: &[]time.Time{baseTime.Add(-time.Duration(2+i%96) * time.Hour)}[0],
			}
			subnets = append(subnets, subnet)
			subnetIds = append(subnetIds, subnetId)
		}

		mockOrgKeeper := &mockOrgKeeper{}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: subnets}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: subnetIds}

		// Create NetIntel with data for all subnets
		reputationResults := make(map[string]map[int]int)
		for i := 0; i < numSubnets; i++ {
			cidr := fmt.Sprintf("192.168.%d.0/24", i+1)
			reputationResults[cidr] = map[int]int{
				1: i%20 + 1, // Level 1: 1-20
				2: i%10 + 1, // Level 2: 1-10
				3: i%5 + 1,  // Level 3: 1-5
				4: i % 3,    // Level 4: 0-2
			}
		}

		mockNetIntel := &mockNetIntelService{
			ready:             true,
			reputationResults: reputationResults,
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

		startTime := time.Now()

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should handle large subnet dataset")

		processingTime := time.Since(startTime)
		t.Logf("Processed %d subnets in %v", numSubnets, processingTime)

		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse vote extension")

		// Verify subnet validations are included
		require.Greater(t, len(voteExt.SubnetRirSync), 150, "should sync most subnets for RIR")
		require.Greater(t, len(voteExt.RpkiValidation), 150, "should validate most subnets for RPKI")
		require.Equal(t, numSubnets, len(voteExt.SubnetBlacklistSync), "should sync all subnets for blacklist")

		// Performance assertion
		require.Less(t, processingTime, 60*time.Second, "should complete subnet validation within 60 seconds")

		// Verify data integrity
		for subnetId, blacklistData := range voteExt.SubnetBlacklistSync {
			require.GreaterOrEqual(t, blacklistData.Level1Count, blacklistData.Level2Count,
				"Level 1 should be >= Level 2 for subnet %d", subnetId)
			require.GreaterOrEqual(t, blacklistData.Level2Count, blacklistData.Level3Count,
				"Level 2 should be >= Level 3 for subnet %d", subnetId)
		}
	})

	t.Run("Large Consensus Aggregation Performance", func(t *testing.T) {
		t.Skip("Skipping consensus aggregation performance test - requires complex validator store setup")
		// Test performance with many validators and complex vote extensions
		numValidators := 50
		numOrgs := 100
		numSubnets := 100

		// Create validators with varying power
		validators := make([]abcitypes.Validator, numValidators)
		for i := 0; i < numValidators; i++ {
			validators[i] = abcitypes.Validator{
				Address: []byte(fmt.Sprintf("validator%d", i)),
				Power:   int64(100 + i*10), // Powers from 100 to 590
			}
		}

		// Create vote extensions from each validator
		votes := make([]abcitypes.ExtendedVoteInfo, numValidators)
		for i := 0; i < numValidators; i++ {
			// Each validator validates a subset of orgs and subnets
			orgValidations := make(map[uint64]bool)
			subnetSyncs := make(map[uint64]subnettypes.SubnetInfo)
			blacklistSyncs := make(map[uint64]trusttypes.SubnetBlacklistCount)

			// Each validator validates different subsets (simulating distributed work)
			startOrg := (i * numOrgs) / numValidators
			endOrg := ((i + 1) * numOrgs) / numValidators
			for orgId := startOrg; orgId < endOrg; orgId++ {
				orgValidations[uint64(1000+orgId)] = (orgId+i)%3 != 0 // Vary results
			}

			startSubnet := (i * numSubnets) / numValidators
			endSubnet := ((i + 1) * numSubnets) / numValidators
			for subnetIdx := startSubnet; subnetIdx < endSubnet; subnetIdx++ {
				subnetId := uint64(2000 + subnetIdx)

				subnetSyncs[subnetId] = subnettypes.SubnetInfo{
					Valid:          true,
					OrgID:          fmt.Sprintf("ORG-%d-%d", i, subnetIdx),
					OrgName:        fmt.Sprintf("Validator %d Org %d", i, subnetIdx),
					AbuseEmail:     fmt.Sprintf("abuse%d@example%d.com", i, subnetIdx),
					AllocationType: "ALLOCATED",
					Rir:            []string{"ARIN", "RIPE"}[subnetIdx%2],
				}

				blacklistSyncs[subnetId] = trusttypes.SubnetBlacklistCount{
					Level1Count: uint32((subnetIdx + i) % 20),
					Level2Count: uint32((subnetIdx + i) % 10),
					Level3Count: uint32((subnetIdx + i) % 5),
					Level4Count: uint32((subnetIdx + i) % 3),
				}
			}

			voteExt := abci.OracleVoteExtension{
				Height:              ctx.BlockHeight(),
				OrgDomainValidation: orgValidations,
				SubnetRirSync:       subnetSyncs,
				SubnetBlacklistSync: blacklistSyncs,
			}

			voteExtBytes, err := json.Marshal(voteExt)
			require.NoError(t, err, "should marshal vote extension %d", i)

			votes[i] = abcitypes.ExtendedVoteInfo{
				Validator:     validators[i],
				BlockIdFlag:   2, // BlockIDFlagCommit
				VoteExtension: voteExtBytes,
			}
		}

		// Test consensus aggregation performance
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

		commitInfo := abcitypes.ExtendedCommitInfo{Votes: votes}

		startTime := time.Now()

		req := &abcitypes.RequestPrepareProposal{
			Height:          ctx.BlockHeight(),
			LocalLastCommit: commitInfo,
		}

		resp, err := proposalHandler.PrepareProposal()(ctx, req)
		require.NoError(t, err, "should aggregate large number of vote extensions")

		aggregationTime := time.Since(startTime)
		t.Logf("Aggregated %d validators with %d orgs and %d subnets in %v",
			numValidators, numOrgs, numSubnets, aggregationTime)

		require.NotEmpty(t, resp.Txs, "should have injected validation transaction")

		// Parse aggregated results
		var validationResults abci.NetworkValidationResults
		err = json.Unmarshal(resp.Txs[0], &validationResults)
		require.NoError(t, err, "should parse aggregation results")

		// Verify consensus results
		require.Greater(t, len(validationResults.OrgValidationResults), 80,
			"should have consensus on most organizations")
		require.Greater(t, len(validationResults.SubnetSyncResults), 80,
			"should have consensus on most subnet syncs")
		require.Greater(t, len(validationResults.SubnetBlacklistSyncResults), 80,
			"should have consensus on most blacklist syncs")

		// Performance assertion
		require.Less(t, aggregationTime, 5*time.Second, "should aggregate within 5 seconds")
	})

	t.Run("Concurrent Processing Stress Test", func(t *testing.T) {
		// Simulate high load with concurrent processing
		numConcurrentVoteExtensions := 20
		results := make(chan time.Duration, numConcurrentVoteExtensions)
		errors := make(chan error, numConcurrentVoteExtensions)

		// Create test data
		orgs := make([]orgtypes.Org, 50)
		for i := 0; i < 50; i++ {
			orgs[i] = orgtypes.Org{
				Id:                uint64(1000 + i),
				Website:           fmt.Sprintf("https://concurrent%d.com", i),
				DomainVerifAttemptedAt: &[]time.Time{baseTime.Add(-3 * time.Hour)}[0],
			}
		}

		subnets := make([]subnettypes.Subnet, 50)
		subnetIds := make([]uint64, 50)
		for i := 0; i < 50; i++ {
			subnetId := uint64(2000 + i)
			subnets[i] = subnettypes.Subnet{
				Id:             subnetId,
				NetworkAddress: 3232235520 + uint64(i*256),
				Mask:           24,
				RirSyncedAt:      &[]time.Time{baseTime.Add(-3 * time.Hour)}[0],
				RpkiLastSyncedAt: &[]time.Time{baseTime.Add(-3 * time.Hour)}[0],
			}
			subnetIds[i] = subnetId
		}

		// Start concurrent vote extension generations
		startTime := time.Now()

		for i := 0; i < numConcurrentVoteExtensions; i++ {
			go func(workerID int) {
				workerStart := time.Now()

				mockOrgKeeper := &mockOrgKeeper{orgs: orgs}
				mockSubnetKeeper := &mockSubnetKeeper{subnets: subnets}
				mockLeaseKeeper := &mockLeaseKeeper{}
				mockTrustKeeper := &mockTrustKeeper{subnets: subnetIds}
				mockNetIntel := &mockNetIntelService{
					ready: true,
					reputationResults: map[string]map[int]int{
						fmt.Sprintf("192.168.%d.0/24", workerID): {1: 5, 2: 3, 3: 1, 4: 0},
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
					Height: ctx.BlockHeight() + int64(workerID), // Vary heights
				}

				_, err := handler.ExtendVoteHandler()(ctx, req)
				if err != nil {
					errors <- fmt.Errorf("worker %d failed: %w", workerID, err)
					return
				}

				results <- time.Since(workerStart)
			}(i)
		}

		// Collect results
		var totalTime time.Duration
		var maxTime time.Duration
		var minTime time.Duration = time.Hour // Initialize to high value

		for i := 0; i < numConcurrentVoteExtensions; i++ {
			select {
			case err := <-errors:
				require.NoError(t, err, "concurrent worker should not fail")
			case duration := <-results:
				totalTime += duration
				if duration > maxTime {
					maxTime = duration
				}
				if duration < minTime {
					minTime = duration
				}
			case <-time.After(60 * time.Second):
				t.Fatal("concurrent test timed out")
			}
		}

		overallTime := time.Since(startTime)
		avgTime := totalTime / time.Duration(numConcurrentVoteExtensions)

		t.Logf("Concurrent processing stats:")
		t.Logf("  Overall time: %v", overallTime)
		t.Logf("  Average worker time: %v", avgTime)
		t.Logf("  Max worker time: %v", maxTime)
		t.Logf("  Min worker time: %v", minTime)

		// Performance assertions
		require.Less(t, overallTime, 60*time.Second, "overall concurrent processing should complete within 60 seconds")
		require.Less(t, avgTime, 30*time.Second, "average worker time should be under 30 seconds")
		require.Less(t, maxTime, 40*time.Second, "no worker should take more than 40 seconds")
	})

	t.Run("Memory Usage and Resource Management", func(t *testing.T) {
		// Test with very large datasets to check memory efficiency
		numLargeOrgs := 500
		numLargeSubnets := 1000

		// Create large datasets
		largeOrgs := make([]orgtypes.Org, numLargeOrgs)
		for i := 0; i < numLargeOrgs; i++ {
			largeOrgs[i] = orgtypes.Org{
				Id:                uint64(10000 + i),
				Website:           fmt.Sprintf("https://large%d.example.com", i),
				DomainVerifAttemptedAt: &[]time.Time{baseTime.Add(-5 * time.Hour)}[0],
			}
		}

		largeSubnets := make([]subnettypes.Subnet, numLargeSubnets)
		largeSubnetIds := make([]uint64, numLargeSubnets)
		for i := 0; i < numLargeSubnets; i++ {
			subnetId := uint64(20000 + i)
			largeSubnets[i] = subnettypes.Subnet{
				Id:             subnetId,
				NetworkAddress: 167772160 + uint64(i*256), // Start at 10.0.0.0
				Mask:           24,
				RirSyncedAt:      &[]time.Time{baseTime.Add(-5 * time.Hour)}[0],
				RpkiLastSyncedAt: &[]time.Time{baseTime.Add(-5 * time.Hour)}[0],
			}
			largeSubnetIds[i] = subnetId
		}

		// Create NetIntel with comprehensive reputation data
		largeReputationResults := make(map[string]map[int]int)
		for i := 0; i < numLargeSubnets; i++ {
			cidr := fmt.Sprintf("10.%d.%d.0/24", i/256, i%256)
			largeReputationResults[cidr] = map[int]int{
				1: (i*7)%50 + 1,
				2: (i*5)%25 + 1,
				3: (i*3)%12 + 1,
				4: (i * 2) % 6,
			}
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: largeOrgs}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: largeSubnets}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: largeSubnetIds}
		mockNetIntel := &mockNetIntelService{
			ready:             true,
			reputationResults: largeReputationResults,
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

		startTime := time.Now()

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should handle very large datasets")

		processingTime := time.Since(startTime)
		t.Logf("Processed %d orgs and %d subnets in %v", numLargeOrgs, numLargeSubnets, processingTime)

		require.NotEmpty(t, resp.VoteExtension, "should generate vote extension")

		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse large vote extension")

		// Verify comprehensive processing
		require.Equal(t, numLargeOrgs, len(voteExt.OrgDomainValidation), "should validate all large orgs")
		require.Equal(t, numLargeSubnets, len(voteExt.SubnetRirSync), "should sync all large subnets")
		require.Equal(t, numLargeSubnets, len(voteExt.RpkiValidation), "should validate RPKI for all large subnets")
		require.Equal(t, numLargeSubnets, len(voteExt.SubnetBlacklistSync), "should sync blacklist for all large subnets")

		// Memory efficiency assertion (should complete within reasonable time even with large datasets)
		require.Less(t, processingTime, 2*time.Minute, "should handle large datasets within 2 minutes")

		// Verify vote extension size is reasonable (should be able to marshal/unmarshal)
		voteExtSize := len(resp.VoteExtension)
		t.Logf("Vote extension size: %d bytes (%.2f MB)", voteExtSize, float64(voteExtSize)/(1024*1024))
		require.Less(t, voteExtSize, 10*1024*1024, "vote extension should be under 10MB")
	})
}

// TestOracleTimeoutAndFailureHandling tests timeout and failure scenarios
func TestOracleTimeoutAndFailureHandling(t *testing.T) {
	oracleKeeper, ctx := keepertest.OracleKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(6000)
	
	// Set up consensus parameters to avoid nil pointer dereference
	consParams := cmtproto.ConsensusParams{
		Abci: &cmtproto.ABCIParams{
			VoteExtensionsEnableHeight: 1, // Enable vote extensions from height 1
		},
	}
	ctx = ctx.WithConsensusParams(consParams)

	baseTime := ctx.BlockTime()
	oracleKeeper.SetLastBlockTime(ctx, baseTime.Add(-2*time.Hour))

	t.Run("Partial Validation Failure Resilience", func(t *testing.T) {
		// Mix of valid and problematic data
		orgs := []orgtypes.Org{
			{Id: 1001, Website: "https://valid.com", DomainVerifAttemptedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]},
			{Id: 1002, Website: "", DomainVerifAttemptedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]}, // Invalid website
			{Id: 1003, Website: "https://another-valid.com", DomainVerifAttemptedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]},
		}

		subnets := []subnettypes.Subnet{
			{Id: 2001, NetworkAddress: 3232235521, Mask: 24, RirSyncedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]},
			{Id: 2002, NetworkAddress: 0, Mask: 24, RirSyncedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]}, // Invalid IP
			{Id: 2003, NetworkAddress: 3232235777, Mask: 24, RirSyncedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0]},
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: orgs}
		mockSubnetKeeper := &mockSubnetKeeper{subnets: subnets}
		mockLeaseKeeper := &mockLeaseKeeper{}
		mockTrustKeeper := &mockTrustKeeper{subnets: []uint64{2001, 2002, 2003}}
		mockNetIntel := &mockNetIntelService{
			ready: true,
			reputationResults: map[string]map[int]int{
				"192.168.1.0/24": {1: 5, 2: 3, 3: 1, 4: 0},
				"192.168.2.0/24": {1: 8, 2: 4, 3: 2, 4: 1},
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
		require.NoError(t, err, "should handle partial failures gracefully")

		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse vote extension despite failures")

		// Should include valid validations and exclude invalid ones
		require.Contains(t, voteExt.OrgDomainValidation, uint64(1001), "should validate valid org")
		require.Contains(t, voteExt.OrgDomainValidation, uint64(1003), "should validate another valid org")
		// Org 1002 might be included but with failed validation result

		require.Contains(t, voteExt.SubnetRirSync, uint64(2001), "should sync valid subnet")
		require.Contains(t, voteExt.SubnetRirSync, uint64(2003), "should sync another valid subnet")
	})

	t.Run("Validator Vote Extension Size Limits", func(t *testing.T) {
		// Test with data that might exceed practical limits
		numExtremeOrgs := 2000
		var extremeOrgs []orgtypes.Org

		for i := 0; i < numExtremeOrgs; i++ {
			extremeOrgs = append(extremeOrgs, orgtypes.Org{
				Id:                uint64(50000 + i),
				Website:           fmt.Sprintf("https://extreme%d.verylongdomainname.example.com", i),
				DomainVerifAttemptedAt: &[]time.Time{ctx.BlockTime().Add(-5 * time.Hour)}[0],
			})
		}

		mockOrgKeeper := &mockOrgKeeper{orgs: extremeOrgs}
		mockSubnetKeeper := &mockSubnetKeeper{}
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

		req := &abcitypes.RequestExtendVote{
			Height: ctx.BlockHeight(),
		}

		resp, err := handler.ExtendVoteHandler()(ctx, req)
		require.NoError(t, err, "should handle extreme dataset sizes")

		// Check that vote extension is within reasonable size bounds
		voteExtSize := len(resp.VoteExtension)
		t.Logf("Extreme dataset vote extension size: %d bytes", voteExtSize)

		// Should still be manageable size (adjust based on actual requirements)
		require.Less(t, voteExtSize, 50*1024*1024, "vote extension should remain under 50MB even with extreme data")

		var voteExt abci.OracleVoteExtension
		err = json.Unmarshal(resp.VoteExtension, &voteExt)
		require.NoError(t, err, "should parse extreme vote extension")

		require.Equal(t, numExtremeOrgs, len(voteExt.OrgDomainValidation), "should handle all extreme orgs")
	})
}
