package keeper

import (
	"context"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ipex/x/org/types"
)

func (k Keeper) OrgByAdmin(ctx context.Context, req *types.QueryOrgByAdminRequest) (*types.QueryOrgByAdminResponse, error) {
	if req == nil || req.Admin == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Get org ID from admin index
	orgId, err := k.OrgsByAdmin.Get(ctx, req.Admin)
	if err != nil {
		if strings.Contains(err.Error(), "collections: not found") {
			return &types.QueryOrgByAdminResponse{
				Org: nil,
				Pagination: &query.PageResponse{Total: 0},
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get the org
	org, found := k.GetOrg(ctx, orgId)
	if !found {
		return &types.QueryOrgByAdminResponse{
			Org: nil,
			Pagination: &query.PageResponse{Total: 0},
		}, nil
	}

	return &types.QueryOrgByAdminResponse{
		Org: &org,
		Pagination: &query.PageResponse{Total: 1},
	}, nil
}