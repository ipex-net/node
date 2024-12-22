package keeper

import (
	"context"

	"ipex/x/subnet/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SubnetAll(ctx context.Context, req *types.QueryAllSubnetRequest) (*types.QueryAllSubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnets []types.Subnet

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	subnetStore := prefix.NewStore(store, types.KeyPrefix(types.SubnetKey))

	pageRes, err := query.Paginate(subnetStore, req.Pagination, func(key []byte, value []byte) error {
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

	return &types.QueryAllSubnetResponse{Subnet: subnets, Pagination: pageRes}, nil
}

func (k Keeper) Subnet(ctx context.Context, req *types.QueryGetSubnetRequest) (*types.QueryGetSubnetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	subnet, found := k.GetSubnet(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetSubnetResponse{Subnet: subnet}, nil
}
