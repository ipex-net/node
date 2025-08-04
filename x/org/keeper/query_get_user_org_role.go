package keeper

import (
	"context"

	"ipex/x/org/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GetUserOrgRole(goCtx context.Context, req *types.QueryGetUserOrgRoleRequest) (*types.QueryGetUserOrgRoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate address format
	_, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address format")
	}

	// Check if user is a creator of any org
	creatorOrg, isCreator := k.GetOrgByCreatorAddress(ctx, req.Address)
	
	// Check if user is an admin of any org  
	adminOrg, isAdmin := k.GetOrgByAdminAddress(ctx, req.Address)

	// Validate that user can only be part of one org
	if isCreator && isAdmin && creatorOrg.Id != adminOrg.Id {
		return nil, status.Error(codes.Internal, "user cannot be creator of one org and admin of another")
	}

	// Initialize response
	response := &types.QueryGetUserOrgRoleResponse{}

	if isAdmin {
		response.OrgId = adminOrg.Id
		response.LegalName = adminOrg.LegalName
		response.IsVerified = adminOrg.DomainVerified
		response.IsAdmin = true
	}
	
	if isCreator {
		response.OrgId = creatorOrg.Id
		response.LegalName = creatorOrg.LegalName
		response.IsVerified = creatorOrg.DomainVerified
		response.IsCreator = true
	}

	return response, nil
}
