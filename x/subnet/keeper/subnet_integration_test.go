package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/x/subnet/types"

	"github.com/stretchr/testify/require"
)

// TestSubnetLifecycleIntegration tests the complete subnet lifecycle from creation to deletion
func TestSubnetLifecycleIntegration(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data - using valid public IP ranges
	networkAddress1 := uint64(0x08080800) // 8.8.8.0 (Google DNS range)
	networkAddress2 := uint64(0x01010100) // 1.1.1.0 (Cloudflare DNS range)
	ownerOrgId := uint64(1001)

	t.Run("Create Multiple Subnets", func(t *testing.T) {
		// Create first subnet
		subnet1 := types.Subnet{
			NetworkAddress: networkAddress1,
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     ownerOrgId,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}

		id1 := keeper.AppendSubnet(ctx, subnet1)
		require.Greater(t, id1, uint64(0), "should create first subnet")

		// Create second subnet
		subnet2 := types.Subnet{
			NetworkAddress: networkAddress2,
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     ownerOrgId,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}

		id2 := keeper.AppendSubnet(ctx, subnet2)
		require.Greater(t, id2, uint64(0), "should create second subnet")
		require.NotEqual(t, id1, id2, "should have different IDs")

		// Verify both subnets exist
		retrievedSubnet1, found := keeper.GetSubnet(ctx, id1)
		require.True(t, found, "should find first subnet")
		require.Equal(t, id1, retrievedSubnet1.Id, "should have correct ID")
		require.Equal(t, networkAddress1, retrievedSubnet1.NetworkAddress, "should have correct network address")

		retrievedSubnet2, found := keeper.GetSubnet(ctx, id2)
		require.True(t, found, "should find second subnet")
		require.Equal(t, id2, retrievedSubnet2.Id, "should have correct ID")
		require.Equal(t, networkAddress2, retrievedSubnet2.NetworkAddress, "should have correct network address")
	})

	t.Run("Update Subnet Status", func(t *testing.T) {
		// Create subnet for status testing
		subnet := types.Subnet{
			NetworkAddress: uint64(0x04040400), // 4.4.4.0 (Level3 DNS)
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     ownerOrgId,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}

		id := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, id, uint64(0), "should create subnet")

		// Update status to READY
		retrievedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")

		retrievedSubnet.Status = types.SubnetStatus_SUBNET_STATUS_READY
		retrievedSubnet.UpdatedAt = ctx.BlockTime().Add(time.Hour)
		keeper.SetSubnet(ctx, retrievedSubnet)

		// Verify status update
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find updated subnet")
		require.Equal(t, types.SubnetStatus_SUBNET_STATUS_READY, updatedSubnet.Status, "should have updated status")
		require.True(t, updatedSubnet.UpdatedAt.After(updatedSubnet.CreatedAt), "should have newer updated time")
	})

	t.Run("Subnet User Management", func(t *testing.T) {
		// Create subnet for user management testing
		subnet := types.Subnet{
			NetworkAddress: uint64(0x09090900), // 9.9.9.0 (Quad9 DNS)
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     ownerOrgId,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}

		id := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, id, uint64(0), "should create subnet")

		// Set user org
		userOrgId := uint64(2001)
		keeper.SetSubnetUserOrg(ctx, id, userOrgId)

		// Verify user org was set
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.Equal(t, userOrgId, updatedSubnet.UserOrgId, "should have user org set")

		// Remove user org
		keeper.RemoveSubnetUserOrg(ctx, id)

		// Verify user org was removed
		finalSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.Equal(t, uint64(0), finalSubnet.UserOrgId, "should have user org removed")
		require.Equal(t, types.SubnetStatus_SUBNET_STATUS_READY, finalSubnet.Status, "should reset to ready status")
	})

	t.Run("Subnet Removal", func(t *testing.T) {
		// Create subnet for removal testing
		subnet := types.Subnet{
			NetworkAddress: uint64(0x0A0A0A00), // 10.10.10.0 (example range)
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     ownerOrgId,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}

		id := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, id, uint64(0), "should create subnet")

		// Verify subnet exists
		_, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet before removal")

		// Remove subnet
		keeper.RemoveSubnet(ctx, id)

		// Verify subnet was removed
		_, found = keeper.GetSubnet(ctx, id)
		require.False(t, found, "should not find subnet after removal")

		// Verify subnet is not in the list
		allSubnets := keeper.GetAllSubnet(ctx)
		for _, s := range allSubnets {
			require.NotEqual(t, id, s.Id, "removed subnet should not appear in list")
		}
	})
}

// TestSubnetIndexingIntegration tests the indexing system for efficient queries
func TestSubnetIndexingIntegration(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	ownerOrg1 := uint64(1001)
	ownerOrg2 := uint64(1002)
	userOrg1 := uint64(2001)
	userOrg2 := uint64(2002)

	t.Run("Owner Organization Indexing", func(t *testing.T) {
		// Create subnets with different owner orgs
		subnet1 := types.Subnet{
			NetworkAddress: uint64(0x08080800), // 8.8.8.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     ownerOrg1,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id1 := keeper.AppendSubnet(ctx, subnet1)

		subnet2 := types.Subnet{
			NetworkAddress: uint64(0x01010100), // 1.1.1.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     ownerOrg1,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id2 := keeper.AppendSubnet(ctx, subnet2)

		subnet3 := types.Subnet{
			NetworkAddress: uint64(0x04040400), // 4.4.4.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     ownerOrg2,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id3 := keeper.AppendSubnet(ctx, subnet3)

		// Verify indexing by checking all subnets
		allSubnets := keeper.GetAllSubnet(ctx)

		// Count subnets by owner org
		ownerOrg1Count := 0
		ownerOrg2Count := 0
		var ownerOrg1IDs, ownerOrg2IDs []uint64

		for _, subnet := range allSubnets {
			if subnet.OwnerOrgId == ownerOrg1 {
				ownerOrg1Count++
				ownerOrg1IDs = append(ownerOrg1IDs, subnet.Id)
			} else if subnet.OwnerOrgId == ownerOrg2 {
				ownerOrg2Count++
				ownerOrg2IDs = append(ownerOrg2IDs, subnet.Id)
			}
		}

		require.Equal(t, 2, ownerOrg1Count, "should have 2 subnets for owner org 1")
		require.Equal(t, 1, ownerOrg2Count, "should have 1 subnet for owner org 2")
		require.Contains(t, ownerOrg1IDs, id1, "should contain subnet 1 for owner org 1")
		require.Contains(t, ownerOrg1IDs, id2, "should contain subnet 2 for owner org 1")
		require.Contains(t, ownerOrg2IDs, id3, "should contain subnet 3 for owner org 2")
	})

	t.Run("User Organization Indexing", func(t *testing.T) {
		// Create subnet and assign user orgs
		subnet := types.Subnet{
			NetworkAddress: uint64(0x09090900), // 9.9.9.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     ownerOrg1,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Set user org
		keeper.SetSubnetUserOrg(ctx, id, userOrg1)

		// Verify user org assignment
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.Equal(t, userOrg1, updatedSubnet.UserOrgId, "should have correct user org")

		// Change user org
		keeper.SetSubnetUserOrg(ctx, id, userOrg2)

		// Verify user org change
		finalSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.Equal(t, userOrg2, finalSubnet.UserOrgId, "should have updated user org")
	})

	t.Run("Network Address Indexing", func(t *testing.T) {
		// Create subnets with unique network addresses
		networkAddresses := []uint64{
			0x03030300, // 3.3.3.0
			0x05050500, // 5.5.5.0
			0x06060600, // 6.6.6.0
		}

		var subnetIDs []uint64
		for _, netAddr := range networkAddresses {
			subnet := types.Subnet{
				NetworkAddress: netAddr,
				Mask:           24,
				Status:         types.SubnetStatus_SUBNET_STATUS_READY,
				OwnerOrgId:     ownerOrg1,
				CreatedAt:      ctx.BlockTime(),
				UpdatedAt:      ctx.BlockTime(),
			}
			id := keeper.AppendSubnet(ctx, subnet)
			subnetIDs = append(subnetIDs, id)
		}

		// Verify all subnets with unique network addresses exist
		allSubnets := keeper.GetAllSubnet(ctx)
		foundNetworkAddresses := make(map[uint64]bool)

		for _, subnet := range allSubnets {
			foundNetworkAddresses[subnet.NetworkAddress] = true
		}

		for _, netAddr := range networkAddresses {
			require.True(t, foundNetworkAddresses[netAddr], "should find network address %d", netAddr)
		}
	})
}

// TestSubnetValidationIntegration tests validation workflows and external checks
func TestSubnetValidationIntegration(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	t.Run("RIR Validation Workflow", func(t *testing.T) {
		// Create subnet for RIR validation
		subnet := types.Subnet{
			NetworkAddress: uint64(0x08080800), // 8.8.8.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     1001,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Test getting subnets for RIR sync
		subnetsForSync := keeper.GetSubnetsForRirSync(ctx, baseTime)
		found := false
		for _, s := range subnetsForSync {
			if s.Id == id {
				found = true
				break
			}
		}
		require.True(t, found, "should include new subnet in RIR sync list")

		// Simulate successful RIR validation
		subnetInfo := types.SubnetInfo{
			Valid:          true,
			AbuseEmail:     "abuse@example.com",
			OrgID:          "ORG-123",
			OrgName:        "Example Organization",
			AllocationType: "assigned",
			Rir:            "ARIN",
		}

		err := keeper.UpdateSubnetRirDetails(ctx, id, subnetInfo)
		require.NoError(t, err, "should update RIR details successfully")

		// Verify RIR validation results
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.Equal(t, types.SubnetStatus_SUBNET_STATUS_READY, updatedSubnet.Status, "should have ready status")
		require.True(t, updatedSubnet.RirRecordValid, "should have valid RIR record")
		require.Equal(t, "abuse@example.com", updatedSubnet.RirAbuseEmail, "should have abuse email")
		require.Equal(t, "ORG-123", updatedSubnet.RirOwnerOrgId, "should have RIR org ID")
		require.Equal(t, "ARIN", updatedSubnet.Rir, "should have RIR")
		require.NotNil(t, updatedSubnet.RirSyncedAt, "should have RIR sync timestamp")
	})

	t.Run("RIR Validation Failure", func(t *testing.T) {
		// Create subnet for failed RIR validation
		subnet := types.Subnet{
			NetworkAddress: uint64(0x01010100), // 1.1.1.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
			OwnerOrgId:     1001,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Simulate failed RIR validation
		subnetInfo := types.SubnetInfo{
			Valid: false,
		}

		err := keeper.UpdateSubnetRirDetails(ctx, id, subnetInfo)
		require.Error(t, err, "should return error for invalid RIR record")

		// Verify subnet was removed
		_, found := keeper.GetSubnet(ctx, id)
		require.False(t, found, "should remove subnet with invalid RIR record")
	})

	t.Run("RPKI Validation Workflow", func(t *testing.T) {
		// Create subnet for RPKI validation
		subnet := types.Subnet{
			NetworkAddress:            uint64(0x04040400), // 4.4.4.0
			Mask:                      24,
			Status:                    types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                1001,
			UserOrgId:                 2001,
			OriginAsn:                 64512, // Private ASN
			OriginTransferRequestedAt: &baseTime,
			CreatedAt:                 ctx.BlockTime(),
			UpdatedAt:                 ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Test getting subnets for RPKI validation
		subnetsForRpki := keeper.GetSubnetsForRPKIValidation(ctx, baseTime.Add(time.Hour))
		found := false
		for _, s := range subnetsForRpki {
			if s.Id == id {
				found = true
				break
			}
		}
		require.True(t, found, "should include subnet with origin ASN in RPKI validation list")

		// Simulate RPKI validation
		err := keeper.UpdateSubnetRpkiValidationStatus(ctx, id, true)
		require.NoError(t, err, "should update RPKI status successfully")

		// Verify RPKI validation results
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.True(t, updatedSubnet.RpkiOriginValid, "should have valid RPKI origin")
		require.NotNil(t, updatedSubnet.RpkiValidityChangedAt, "should have RPKI validity change timestamp")
		require.NotNil(t, updatedSubnet.RpkiLastSyncedAt, "should have RPKI sync timestamp")
	})

	t.Run("Geolocation Validation Workflow", func(t *testing.T) {
		// Create subnet for geo validation
		subnet := types.Subnet{
			NetworkAddress:      uint64(0x09090900), // 9.9.9.0
			Mask:                24,
			Status:              types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:          1001,
			Country:             "US",
			GeoDetailsChangedAt: &baseTime,
			CreatedAt:           ctx.BlockTime(),
			UpdatedAt:           ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Test getting subnets for geo validation
		subnetsForGeo := keeper.GetSubnetsForGeoValidation(ctx, baseTime.Add(time.Hour))
		found := false
		for _, s := range subnetsForGeo {
			if s.Id == id {
				found = true
				break
			}
		}
		require.True(t, found, "should include subnet with country in geo validation list")

		// Simulate geo validation
		err := keeper.UpdateSubnetGeoValidationStatus(ctx, id, true)
		require.NoError(t, err, "should update geo validation status successfully")

		// Verify geo validation results
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.True(t, updatedSubnet.TraceGeoVerified, "should have verified geo location")
		require.NotNil(t, updatedSubnet.GeoVerifAttemptedAt, "should have geo verification attempt timestamp")
	})

	t.Run("Ownership Verification Workflow", func(t *testing.T) {
		// Create subnet for ownership verification
		subnet := types.Subnet{
			NetworkAddress: uint64(0x03030300), // 3.3.3.0
			Mask:           24,
			Status:         types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:     1001,
			CreatedAt:      ctx.BlockTime(),
			UpdatedAt:      ctx.BlockTime(),
		}
		id := keeper.AppendSubnet(ctx, subnet)

		// Test getting subnets for ownership verification
		subnetsForOwnership := keeper.GetSubnetsForOwnershipVerification(ctx, baseTime.Add(time.Hour))
		found := false
		for _, s := range subnetsForOwnership {
			if s.Id == id {
				found = true
				break
			}
		}
		require.True(t, found, "should include subnet in ownership verification list")

		// Simulate ownership verification
		err := keeper.UpdateSubnetOwnershipVerification(ctx, id, true)
		require.NoError(t, err, "should update ownership verification successfully")

		// Verify ownership verification results
		updatedSubnet, found := keeper.GetSubnet(ctx, id)
		require.True(t, found, "should find subnet")
		require.True(t, updatedSubnet.OwnershipVerified, "should have verified ownership")
		require.NotNil(t, updatedSubnet.OwnershipVerifChangedAt, "should have ownership verification change timestamp")
		require.NotNil(t, updatedSubnet.OwnershipVerifLastAttemptedAt, "should have ownership verification attempt timestamp")
	})
}

// TestSubnetAuthorizationIntegration tests basic authorization functionality
func TestSubnetAuthorizationIntegration(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	ownerOrgId := uint64(2001)
	userOrgId := uint64(3001)

	// Create subnet for all tests to use
	subnet := types.Subnet{
		NetworkAddress: uint64(0x08080800), // 8.8.8.0
		Mask:           24,
		Status:         types.SubnetStatus_SUBNET_STATUS_READY,
		OwnerOrgId:     ownerOrgId,
		CreatedAt:      ctx.BlockTime(),
		UpdatedAt:      ctx.BlockTime(),
	}
	subnetId := keeper.AppendSubnet(ctx, subnet)

	t.Run("Subnet Organization Management", func(t *testing.T) {
		// Verify initial subnet has owner org but no user org
		initialSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.Equal(t, ownerOrgId, initialSubnet.OwnerOrgId, "should have owner org")
		require.Equal(t, uint64(0), initialSubnet.UserOrgId, "should not have user org initially")

		// Set user org
		keeper.SetSubnetUserOrg(ctx, subnetId, userOrgId)

		// Verify user org was set
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.Equal(t, userOrgId, updatedSubnet.UserOrgId, "should have user org set")

		// Remove user org
		keeper.RemoveSubnetUserOrg(ctx, subnetId)

		// Verify user org was removed
		finalSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.Equal(t, uint64(0), finalSubnet.UserOrgId, "should have user org removed")
	})

	t.Run("Lease Availability Check", func(t *testing.T) {
		// Test lease availability for ready subnet
		isAvailable := keeper.CheckSubnetLeaseAvailability(ctx, subnetId)
		require.True(t, isAvailable, "should allow lease for ready subnet")

		// Update subnet to leased status
		subnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")

		subnet.Status = types.SubnetStatus_SUBNET_STATUS_LEASED
		keeper.SetSubnet(ctx, subnet)

		// Test lease availability for leased subnet
		isAvailableLeased := keeper.CheckSubnetLeaseAvailability(ctx, subnetId)
		require.True(t, isAvailableLeased, "should still allow lease check for leased subnet")

		// Test with non-existent subnet
		isAvailableNonExistent := keeper.CheckSubnetLeaseAvailability(ctx, 99999)
		require.False(t, isAvailableNonExistent, "should reject non-existent subnet")
	})

	t.Run("RIR Record Validation", func(t *testing.T) {
		// Test RIR record validation check
		isValid := keeper.IsRirRecordValid(ctx, subnetId)
		require.False(t, isValid, "should not be valid initially")

		// Update subnet with valid RIR record
		subnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")

		subnet.RirRecordValid = true
		keeper.SetSubnet(ctx, subnet)

		// Test RIR record validation check after update
		isValidAfterUpdate := keeper.IsRirRecordValid(ctx, subnetId)
		require.True(t, isValidAfterUpdate, "should be valid after update")

		// Test with non-existent subnet
		isValidNonExistent := keeper.IsRirRecordValid(ctx, 99999)
		require.False(t, isValidNonExistent, "should be false for non-existent subnet")
	})
}

// TestSubnetPerformanceAndScalability tests performance with large numbers of subnets
func TestSubnetPerformanceAndScalability(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	t.Run("Bulk Subnet Operations", func(t *testing.T) {
		// Create many subnets
		numSubnets := 100
		var subnetIDs []uint64

		for i := 0; i < numSubnets; i++ {
			subnet := types.Subnet{
				NetworkAddress: uint64(0x08000000 + (i * 256)), // 8.0.0.0, 8.0.1.0, 8.0.2.0, etc.
				Mask:           24,
				Status:         types.SubnetStatus_SUBNET_STATUS_READY,
				OwnerOrgId:     uint64(1000 + (i % 10)), // Distribute across 10 orgs
				CreatedAt:      ctx.BlockTime(),
				UpdatedAt:      ctx.BlockTime(),
			}
			id := keeper.AppendSubnet(ctx, subnet)
			subnetIDs = append(subnetIDs, id)
		}

		require.Len(t, subnetIDs, numSubnets, "should create all subnets")

		// Test bulk retrieval
		allSubnets := keeper.GetAllSubnet(ctx)
		require.GreaterOrEqual(t, len(allSubnets), numSubnets, "should retrieve at least created subnets")

		// Test individual lookups
		for i := 0; i < 10; i++ { // Test first 10
			subnet, found := keeper.GetSubnet(ctx, subnetIDs[i])
			require.True(t, found, "should find subnet %d", i)
			require.Equal(t, subnetIDs[i], subnet.Id, "should have correct ID")
		}

		// Test removal of some subnets
		for i := 0; i < 10; i++ {
			keeper.RemoveSubnet(ctx, subnetIDs[i])
		}

		// Verify remaining subnets
		remainingSubnets := keeper.GetAllSubnet(ctx)
		removedIDs := make(map[uint64]bool)
		for i := 0; i < 10; i++ {
			removedIDs[subnetIDs[i]] = true
		}

		for _, subnet := range remainingSubnets {
			require.False(t, removedIDs[subnet.Id], "removed subnet should not appear in list")
		}
	})

	t.Run("Index Performance", func(t *testing.T) {
		// Create subnets with specific patterns for indexing tests
		ownerOrgs := []uint64{5001, 5002, 5003}
		userOrgs := []uint64{6001, 6002, 6003}

		var subnetsByOwner = make(map[uint64][]uint64)
		var subnetsByUser = make(map[uint64][]uint64)

		for i := 0; i < 30; i++ {
			ownerOrgId := ownerOrgs[i%len(ownerOrgs)]
			userOrgId := userOrgs[i%len(userOrgs)]

			subnet := types.Subnet{
				NetworkAddress: uint64(0x0C000000 + (i * 256)), // 12.0.0.0, 12.0.1.0, etc.
				Mask:           24,
				Status:         types.SubnetStatus_SUBNET_STATUS_READY,
				OwnerOrgId:     ownerOrgId,
				UserOrgId:      userOrgId,
				CreatedAt:      ctx.BlockTime(),
				UpdatedAt:      ctx.BlockTime(),
			}
			id := keeper.AppendSubnet(ctx, subnet)

			subnetsByOwner[ownerOrgId] = append(subnetsByOwner[ownerOrgId], id)
			subnetsByUser[userOrgId] = append(subnetsByUser[userOrgId], id)
		}

		// Verify indexing by counting subnets per organization
		allSubnets := keeper.GetAllSubnet(ctx)

		actualOwnerCounts := make(map[uint64]int)
		actualUserCounts := make(map[uint64]int)

		for _, subnet := range allSubnets {
			if subnet.OwnerOrgId >= 5001 && subnet.OwnerOrgId <= 5003 {
				actualOwnerCounts[subnet.OwnerOrgId]++
			}
			if subnet.UserOrgId >= 6001 && subnet.UserOrgId <= 6003 {
				actualUserCounts[subnet.UserOrgId]++
			}
		}

		// Each org should have 10 subnets (30 total / 3 orgs)
		for _, orgId := range ownerOrgs {
			require.Equal(t, 10, actualOwnerCounts[orgId], "should have 10 subnets per owner org")
		}

		for _, orgId := range userOrgs {
			require.Equal(t, 10, actualUserCounts[orgId], "should have 10 subnets per user org")
		}
	})

	t.Run("Validation Queue Performance", func(t *testing.T) {
		baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

		// Create subnets in different validation states
		subnetsForValidation := []struct {
			status     types.SubnetStatus
			hasOrigin  bool
			hasCountry bool
		}{
			{types.SubnetStatus_SUBNET_STATUS_INITIAL, false, false}, // RIR sync
			{types.SubnetStatus_SUBNET_STATUS_READY, true, false},    // RPKI validation
			{types.SubnetStatus_SUBNET_STATUS_READY, false, true},    // Geo validation
			{types.SubnetStatus_SUBNET_STATUS_READY, false, false},   // Ownership verification
		}

		var subnetIDs []uint64
		for i, config := range subnetsForValidation {
			subnet := types.Subnet{
				NetworkAddress: uint64(0x0E000000 + (i * 256)), // 14.0.0.0, 14.0.1.0, etc.
				Mask:           24,
				Status:         config.status,
				OwnerOrgId:     7001,
				CreatedAt:      baseTime,
				UpdatedAt:      baseTime,
			}

			if config.hasOrigin {
				subnet.OriginAsn = 64512
				subnet.OriginTransferRequestedAt = &baseTime
			}

			if config.hasCountry {
				subnet.Country = "US"
				subnet.GeoDetailsChangedAt = &baseTime
			}

			id := keeper.AppendSubnet(ctx, subnet)
			subnetIDs = append(subnetIDs, id)
		}

		// Test validation queue methods
		testTime := baseTime.Add(time.Hour)

		// Test RIR sync queue
		rirSubnets := keeper.GetSubnetsForRirSync(ctx, testTime)
		require.GreaterOrEqual(t, len(rirSubnets), 1, "should have subnets for RIR sync")

		// Test RPKI validation queue
		rpkiSubnets := keeper.GetSubnetsForRPKIValidation(ctx, testTime)
		require.GreaterOrEqual(t, len(rpkiSubnets), 1, "should have subnets for RPKI validation")

		// Test geo validation queue
		geoSubnets := keeper.GetSubnetsForGeoValidation(ctx, testTime)
		require.GreaterOrEqual(t, len(geoSubnets), 1, "should have subnets for geo validation")

		// Test ownership verification queue
		ownershipSubnets := keeper.GetSubnetsForOwnershipVerification(ctx, testTime)
		require.GreaterOrEqual(t, len(ownershipSubnets), 1, "should have subnets for ownership verification")
	})
}
