package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "ipex/testutil/keeper"
	"ipex/x/trust/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.TrustKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
