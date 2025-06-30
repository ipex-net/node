package keeper

import (
	"context"

	"ipex/x/org/types"
)

func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return types.Params{}
	}
	return params
}

func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.ParamsStore.Set(ctx, params)
}
