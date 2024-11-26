package keeper

import (
	"context"

	"ipex/x/lease/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListContract(goCtx context.Context, req *types.QueryListContractRequest) (*types.QueryListContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))

	var contracts []types.Contract

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var contract types.Contract
		if err := k.cdc.Unmarshal(value, &contract); err != nil {
			return err
		}

		contracts = append(contracts, contract)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListContractResponse{Contract: contracts, Pagination: pageRes}, nil
}
