package keeper

import (
	"context"
	"fmt"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ShowContract(goCtx context.Context, req *types.QueryShowContractRequest) (*types.QueryShowContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Check that contractId is not nil
	if req.Id == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, found := k.GetContract(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	// Get subnet market details
	networkAddress, mask, country, rir, err := k.subnetKeeper.GetSubnetMarketDetails(ctx, contract.SubnetId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get subnet market details: %v", err))
	}

	return &types.QueryShowContractResponse{
		Contract:       contract,
		NetworkAddress: networkAddress,
		Mask:           mask,
		Country:        country,
		Rir:            rir,
	}, nil
}
