package keeper

import (
	"ipex/x/lease/types"
)

var _ types.QueryServer = Keeper{}
