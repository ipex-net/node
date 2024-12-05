package keeper

import (
	"context"

	"ipex/x/org/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) OrgAll(ctx context.Context, req *types.QueryAllOrgRequest) (*types.QueryAllOrgResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var orgs []types.Org

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	orgStore := prefix.NewStore(store, types.KeyPrefix(types.OrgKey))

	pageRes, err := query.Paginate(orgStore, req.Pagination, func(key []byte, value []byte) error {
		var org types.Org
		if err := k.cdc.Unmarshal(value, &org); err != nil {
			return err
		}

		orgs = append(orgs, org)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllOrgResponse{Org: orgs, Pagination: pageRes}, nil
}

func (k Keeper) Org(ctx context.Context, req *types.QueryGetOrgRequest) (*types.QueryGetOrgResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	org, found := k.GetOrg(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetOrgResponse{Org: org}, nil
}
