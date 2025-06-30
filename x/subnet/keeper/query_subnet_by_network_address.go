package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ipex/x/subnet/types"
)

func (k Keeper) SubnetByNetworkAddress(ctx context.Context, req *types.QuerySubnetByNetworkAddressRequest) (*types.QuerySubnetByNetworkAddressResponse, error) {
	if req == nil || req.NetworkAddress == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnets []types.Subnet

	// Iterate over all subnets with this network address using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.NetworkAddress)
	err := k.SubnetsByNetworkAddress.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// key.K2() is the subnet ID
		subnetId := key.K2()
		subnet, found := k.GetSubnet(ctx, subnetId)
		if found {
			subnets = append(subnets, subnet)
		}
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySubnetByNetworkAddressResponse{
		Subnet:     subnets,
		Pagination: &query.PageResponse{Total: uint64(len(subnets))},
	}, nil
}