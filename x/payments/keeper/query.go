package keeper

import (
	"ipex/x/payments/types"
)

var _ types.QueryServer = Keeper{}
