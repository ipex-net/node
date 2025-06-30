package keeper

import (
	"context"

	"ipex/x/abuse/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return params
	}
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.ParamsStore.Set(ctx, params)
}
