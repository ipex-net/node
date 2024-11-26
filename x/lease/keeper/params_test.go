package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "ipex/testutil/keeper"
	"ipex/x/lease/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.LeaseKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
