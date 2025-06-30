package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ipex/x/subnet/types"
)

func (k Keeper) SubnetByOwnerOrg(ctx context.Context, req *types.QuerySubnetByOwnerOrgRequest) (*types.QuerySubnetByOwnerOrgResponse, error) {
	if req == nil || req.OwnerOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var subnets []types.Subnet

	// Iterate over all subnets with this owner org ID using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.OwnerOrgId)
	err := k.SubnetsByOwnerOrg.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
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

	return &types.QuerySubnetByOwnerOrgResponse{
		Subnet:     subnets,
		Pagination: &query.PageResponse{Total: uint64(len(subnets))},
	}, nil
}