package keeper

import (
	"context"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ipex/x/org/types"
)

func (k Keeper) OrgByCreator(ctx context.Context, req *types.QueryOrgByCreatorRequest) (*types.QueryOrgByCreatorResponse, error) {
	if req == nil || req.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Get org ID from creator index
	orgId, err := k.OrgsByCreator.Get(ctx, req.Creator)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return &types.QueryOrgByCreatorResponse{
				Org:        nil,
				Pagination: &query.PageResponse{Total: 0},
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get the org
	org, found := k.GetOrg(ctx, orgId)
	if !found {
		return &types.QueryOrgByCreatorResponse{
			Org:        nil,
			Pagination: &query.PageResponse{Total: 0},
		}, nil
	}

	return &types.QueryOrgByCreatorResponse{
		Org:        &org,
		Pagination: &query.PageResponse{Total: 1},
	}, nil
}
