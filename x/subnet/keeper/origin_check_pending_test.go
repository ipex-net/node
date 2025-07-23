package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/x/subnet/types"

	"github.com/stretchr/testify/require"
)

// TestOriginCheckPendingIndex tests the origin check pending index functionality
func TestOriginCheckPendingIndex(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// Test data
	ownerOrgId := uint64(1001)

	t.Run("Set Origin Check Pending", func(t *testing.T) {
		// Create subnet with origin check pending false
		subnet := types.Subnet{
			NetworkAddress:     uint64(0x08080800), // 8.8.8.0
			Mask:               24,
			Status:             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:         ownerOrgId,
			OriginCheckPending: false,
			CreatedAt:          ctx.BlockTime(),
			UpdatedAt:          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, subnetId, uint64(0), "should create subnet")

		// Verify subnet is not in pending index initially
		pendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.NotContains(t, pendingSubnets, subnetId, "subnet should not be in pending index initially")

		// Set origin check pending
		err := keeper.SetSubnetOriginCheckPending(ctx, subnetId)
		require.NoError(t, err, "should set origin check pending")

		// Verify subnet is now in pending index
		pendingSubnets = keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.Contains(t, pendingSubnets, subnetId, "subnet should be in pending index")

		// Verify subnet field is updated
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.True(t, updatedSubnet.OriginCheckPending, "subnet should have origin check pending true")
	})

	t.Run("Clear Origin Check Pending", func(t *testing.T) {
		// Create subnet with origin check pending true
		subnet := types.Subnet{
			NetworkAddress:     uint64(0x01010100), // 1.1.1.0
			Mask:               24,
			Status:             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:         ownerOrgId,
			OriginCheckPending: true,
			CreatedAt:          ctx.BlockTime(),
			UpdatedAt:          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)
		require.Greater(t, subnetId, uint64(0), "should create subnet")

		// Verify subnet is in pending index initially
		pendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.Contains(t, pendingSubnets, subnetId, "subnet should be in pending index initially")

		// Clear origin check pending
		err := keeper.ClearSubnetOriginCheckPending(ctx, subnetId)
		require.NoError(t, err, "should clear origin check pending")

		// Verify subnet is no longer in pending index
		pendingSubnets = keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.NotContains(t, pendingSubnets, subnetId, "subnet should not be in pending index")

		// Verify subnet field is updated
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.False(t, updatedSubnet.OriginCheckPending, "subnet should have origin check pending false")
	})

	t.Run("Query With Limit", func(t *testing.T) {
		// Create multiple subnets with origin check pending
		var subnetIDs []uint64
		for i := 0; i < 7; i++ {
			subnet := types.Subnet{
				NetworkAddress:     uint64(0x02000000 + (i * 256)), // 2.0.0.0, 2.0.1.0, etc.
				Mask:               24,
				Status:             types.SubnetStatus_SUBNET_STATUS_READY,
				OwnerOrgId:         ownerOrgId,
				OriginCheckPending: true,
				CreatedAt:          ctx.BlockTime(),
				UpdatedAt:          ctx.BlockTime(),
			}

			subnetId := keeper.AppendSubnet(ctx, subnet)
			subnetIDs = append(subnetIDs, subnetId)
		}

		// Query with limit of 5
		pendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 5)
		require.LessOrEqual(t, len(pendingSubnets), 5, "should return at most 5 subnets")

		// Query with no limit (0)
		allPendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 0)
		require.GreaterOrEqual(t, len(allPendingSubnets), 7, "should return all pending subnets")
	})

	t.Run("Remove Subnet User Org Clears Pending", func(t *testing.T) {
		// Create subnet with origin check pending and user org
		subnet := types.Subnet{
			NetworkAddress:     uint64(0x04040400), // 4.4.4.0
			Mask:               24,
			Status:             types.SubnetStatus_SUBNET_STATUS_READY,
			OwnerOrgId:         ownerOrgId,
			UserOrgId:          2001,
			OriginCheckPending: true,
			CreatedAt:          ctx.BlockTime(),
			UpdatedAt:          ctx.BlockTime(),
		}

		subnetId := keeper.AppendSubnet(ctx, subnet)

		// Verify subnet is in pending index
		pendingSubnets := keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.Contains(t, pendingSubnets, subnetId, "subnet should be in pending index")

		// Remove subnet user org (this should clear origin check pending)
		keeper.RemoveSubnetUserOrg(ctx, subnetId)

		// Verify subnet is no longer in pending index
		pendingSubnets = keeper.GetOriginCheckPendingSubnets(ctx, 10)
		require.NotContains(t, pendingSubnets, subnetId, "subnet should not be in pending index after removing user org")

		// Verify subnet field is cleared
		updatedSubnet, found := keeper.GetSubnet(ctx, subnetId)
		require.True(t, found, "should find updated subnet")
		require.False(t, updatedSubnet.OriginCheckPending, "subnet should have origin check pending false")
		require.Equal(t, uint64(0), updatedSubnet.UserOrgId, "subnet should have user org cleared")
	})

	t.Run("Error Cases", func(t *testing.T) {
		// Test with non-existent subnet ID
		err := keeper.SetSubnetOriginCheckPending(ctx, 99999)
		require.Error(t, err, "should return error for non-existent subnet")
		require.Contains(t, err.Error(), "not found", "error should mention subnet not found")

		err = keeper.ClearSubnetOriginCheckPending(ctx, 99999)
		require.Error(t, err, "should return error for non-existent subnet")
		require.Contains(t, err.Error(), "not found", "error should mention subnet not found")
	})
}
