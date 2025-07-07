package keeper_test

import (
	"context"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/subnet/keeper"
	"ipex/x/subnet/types"

	"github.com/stretchr/testify/require"
)

func createNSubnet(keeper keeper.Keeper, ctx context.Context, n int) []types.Subnet {
	items := make([]types.Subnet, n)
	for i := range items {
		items[i].Id = keeper.AppendSubnet(ctx, items[i])
	}
	return items
}

func TestSubnetGet(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	items := createNSubnet(keeper, ctx, 10)
	for _, item := range items {
		got, found := keeper.GetSubnet(ctx, item.Id)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&got),
		)
	}
}

func TestSubnetRemove(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	items := createNSubnet(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveSubnet(ctx, item.Id)
		// Check that the subnet was removed by checking if it exists in GetAllSubnet
		allSubnets := keeper.GetAllSubnet(ctx)
		found := false
		for _, subnet := range allSubnets {
			if subnet.Id == item.Id {
				found = true
				break
			}
		}
		require.False(t, found)
	}
}

func TestSubnetGetAll(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	items := createNSubnet(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllSubnet(ctx)),
	)
}

func TestSubnetCount(t *testing.T) {
	keeper, ctx := keepertest.SubnetKeeper(t)
	items := createNSubnet(keeper, ctx, 10)
	count := uint64(len(items))
	allSubnets := keeper.GetAllSubnet(ctx)
	require.Equal(t, count, uint64(len(allSubnets)))
}
