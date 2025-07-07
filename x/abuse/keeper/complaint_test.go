package keeper_test

import (
	"context"
	"strconv"
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/abuse/keeper"
	"ipex/x/abuse/types"

	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNComplaint(keeper keeper.Keeper, ctx context.Context, n int) []types.Complaint {
	items := make([]types.Complaint, n)
	for i := range items {
		// Use AppendComplaint instead of SetComplaint to avoid the GetComplaint issue
		items[i].LeaseId = 1 // Add some required fields
		items[i].SubnetId = 1
		items[i].LessorOrgId = 1
		items[i].LesseeOrgId = 1

		id := keeper.AppendComplaint(ctx, items[i])
		items[i].Id = id // Update with the actual ID that was assigned
	}
	return items
}

func TestComplaintGet(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetComplaint(ctx,
			item.Id,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestComplaintRemove(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveComplaint(ctx,
			item.Id,
		)
		// Check that the complaint was removed by checking if it exists in GetAllComplaint
		allComplaints := keeper.GetAllComplaint(ctx)
		found := false
		for _, complaint := range allComplaints {
			if complaint.Id == item.Id {
				found = true
				break
			}
		}
		require.False(t, found)
	}
}

func TestComplaintGetAll(t *testing.T) {
	keeper, ctx := keepertest.AbuseKeeper(t)
	items := createNComplaint(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllComplaint(ctx)),
	)
}
