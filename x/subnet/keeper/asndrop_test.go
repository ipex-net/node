package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/x/subnet/types"

	"github.com/stretchr/testify/require"
)

// TestASNDROPFunctionality tests the ASNDROP-related functions in the subnet keeper
func TestASNDROPFunctionality(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	ownerOrgId := uint64(1001)
	userOrgId := uint64(2001)

	t.Run("FlagUserRequestedASNDROPOrigin", func(t *testing.T) {
		// Create a subnet with origin settings
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x08080800), // 8.8.8.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          userOrgId,
			OriginAsn:                          65001,
			OriginCheckPending:                 true,
			UserRequestedASNDROPOriginTransfer: false,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, subnetId, uint64(0), "should create subnet")

		// Call FlagUserRequestedASNDROPOrigin
		err := keeper.FlagUserRequestedASNDROPOrigin(ctx, subnetId)
		require.NoError(t, err, "should flag ASNDROP origin transfer")

		// Verify the changes
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.True(t, updatedSubnet.UserRequestedASNDROPOriginTransfer, "should have ASNDROP flag set")
		require.Equal(t, uint64(0), updatedSubnet.OriginAsn, "should reset origin ASN to 0")
		require.False(t, updatedSubnet.OriginCheckPending, "should clear origin check pending")
	})

	t.Run("FlagUserRequestedASNDROPOrigin - Already Flagged", func(t *testing.T) {
		// Create a subnet that already has ASNDROP flag set
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x01010100), // 1.1.1.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          userOrgId,
			OriginAsn:                          65002,
			OriginCheckPending:                 false,
			UserRequestedASNDROPOriginTransfer: true, // Already flagged
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, subnetId, uint64(0), "should create subnet")

		// Call FlagUserRequestedASNDROPOrigin - should be idempotent
		err := keeper.FlagUserRequestedASNDROPOrigin(ctx, subnetId)
		require.NoError(t, err, "should handle already flagged subnet")

		// Verify no changes occurred
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.True(t, updatedSubnet.UserRequestedASNDROPOriginTransfer, "should still have ASNDROP flag set")
	})

	t.Run("FlagUserRequestedASNDROPOrigin - Non-existent Subnet", func(t *testing.T) {
		// Test with non-existent subnet ID
		err := keeper.FlagUserRequestedASNDROPOrigin(ctx, 99999)
		require.Error(t, err, "should return error for non-existent subnet")
		require.Contains(t, err.Error(), "not found", "error should mention subnet not found")
	})

	t.Run("SetSubnetUserOrg with ASNDROP Parameters", func(t *testing.T) {
		// Create a base subnet
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x02020200), // 2.2.2.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          0, // No user org initially
			AllowUserASNDROPOriginTransfer:     false,
			UserRequestedASNDROPOriginTransfer: false,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Set user org with ASNDROP transfer allowed
		keeper.SetSubnetUserOrg(ctx, subnetId, userOrgId, true, true) // allowUserGeoChange=true, allowUserASNDROPOriginTransfer=true

		// Verify the settings
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.Equal(t, userOrgId, updatedSubnet.UserOrgId, "should set user org ID")
		require.True(t, updatedSubnet.AllowUserGeoChange, "should allow user geo change")
		require.True(t, updatedSubnet.AllowUserASNDROPOriginTransfer, "should allow user ASNDROP origin transfer")
		require.False(t, updatedSubnet.UserRequestedASNDROPOriginTransfer, "should not have requested flag set")
	})

	t.Run("SetSubnetUserOrg without ASNDROP Permission", func(t *testing.T) {
		// Create a base subnet
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x03030300), // 3.3.3.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          0,
			AllowUserASNDROPOriginTransfer:     false,
			UserRequestedASNDROPOriginTransfer: false,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Set user org without ASNDROP transfer permission
		keeper.SetSubnetUserOrg(ctx, subnetId, userOrgId, false, false) // allowUserGeoChange=false, allowUserASNDROPOriginTransfer=false

		// Verify the settings
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.Equal(t, userOrgId, updatedSubnet.UserOrgId, "should set user org ID")
		require.False(t, updatedSubnet.AllowUserGeoChange, "should not allow user geo change")
		require.False(t, updatedSubnet.AllowUserASNDROPOriginTransfer, "should not allow user ASNDROP origin transfer")
	})

	t.Run("RemoveSubnetUserOrg Clears ASNDROP Flags", func(t *testing.T) {
		// Create a subnet with ASNDROP settings enabled
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x04040400), // 4.4.4.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          userOrgId,
			OriginAsn:                          65003,
			OriginCheckPending:                 true,
			AllowUserASNDROPOriginTransfer:     true,
			UserRequestedASNDROPOriginTransfer: true,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Remove subnet user org
		keeper.RemoveSubnetUserOrg(ctx, subnetId)

		// Verify all ASNDROP and origin-related flags are cleared
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.Equal(t, uint64(0), updatedSubnet.UserOrgId, "should clear user org ID")
		require.Equal(t, uint64(0), updatedSubnet.OriginAsn, "should clear origin ASN")
		require.False(t, updatedSubnet.OriginCheckPending, "should clear origin check pending")
		require.False(t, updatedSubnet.UserRequestedASNDROPOriginTransfer, "should clear ASNDROP request flag")
		require.Equal(t, types.SubnetStatus_SUBNET_STATUS_READY, updatedSubnet.Status, "should reset status to ready")
	})

	t.Run("ASNDROP Field Persistence", func(t *testing.T) {
		// Create subnet and verify ASNDROP fields persist through updates
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x05050500), // 5.5.5.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          userOrgId,
			AllowUserASNDROPOriginTransfer:     true,
			UserRequestedASNDROPOriginTransfer: false,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Flag ASNDROP origin
		err := keeper.FlagUserRequestedASNDROPOrigin(ctx, subnetId)
		require.NoError(t, err, "should flag ASNDROP")

		// Update subnet (simulate other operations)
		updatedSubnet, _ := keeper.GetSubnet(ctx, subnetId)
		updatedSubnet.Mask = 25 // Change something else
		keeper.SetSubnet(ctx, updatedSubnet)

		// Verify ASNDROP flag persists
		finalSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.Equal(t, uint64(25), finalSubnet.Mask, "should have updated mask")
		require.True(t, finalSubnet.AllowUserASNDROPOriginTransfer, "should preserve AllowUserASNDROPOriginTransfer")
		require.True(t, finalSubnet.UserRequestedASNDROPOriginTransfer, "should preserve UserRequestedASNDROPOriginTransfer")
	})

	t.Run("ASNDROP Integration with Origin Check Pending", func(t *testing.T) {
		// Test the interaction between ASNDROP functionality and origin check pending
		subnet := types.Subnet{
			NetworkAddress:                     uint64(0x06060600), // 6.6.6.0
			Mask:                               24,
			Status:                             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:                         ownerOrgId,
			UserOrgId:                          userOrgId,
			OriginAsn:                          65004,
			OriginCheckPending:                 true,
			AllowUserASNDROPOriginTransfer:     true,
			UserRequestedASNDROPOriginTransfer: false,
			CreatedAt:                          ctx.BlockTime(),
			UpdatedAt:                          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Verify subnet is in pending origin check list
		pendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.Contains(t, pendingSubnets, subnetId, "subnet should be in pending origin check")

		// Flag ASNDROP origin (should clear pending status)
		err := keeper.FlagUserRequestedASNDROPOrigin(ctx, subnetId)
		require.NoError(t, err, "should flag ASNDROP origin")

		// Verify subnet is no longer in pending list
		pendingSubnetsAfter := keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.NotContains(t, pendingSubnetsAfter, subnetId, "subnet should not be in pending origin check after ASNDROP flag")

		// Verify flags are correct
		finalSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find subnet")
		require.False(t, finalSubnet.OriginCheckPending, "should clear origin check pending")
		require.True(t, finalSubnet.UserRequestedASNDROPOriginTransfer, "should set ASNDROP flag")
		require.Equal(t, uint64(0), finalSubnet.OriginAsn, "should reset origin ASN")
	})
}
