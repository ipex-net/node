package keeper

import (
	"context"

	"ipex/x/subnet/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListSubnet(goCtx context.Context, req *types.QueryListSubnetRequest) (*types.QueryListSubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SubnetKey))

	var subnets []types.Subnet

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var subnet types.Subnet
		if err := k.cdc.Unmarshal(value, &subnet); err != nil {
			return err
		}

		subnets = append(subnets, subnet)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListSubnetResponse{Subnet: subnets, Pagination: pageRes}, nil

}
