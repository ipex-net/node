package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker handles contract timeouts at the end of each block
func (k Keeper) EndBlocker(ctx sdk.Context) {
    k.ProcessEnrollmentTimeouts(ctx)
    // Only process timeouts every 100 blocks (or whatever interval makes sense)
    // if ctx.BlockHeight() % 100 == 0 {
    //     k.ProcessEnrollmentTimeouts(sdk.WrapSDKContext(ctx))
    // }
}