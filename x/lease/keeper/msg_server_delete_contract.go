package keeper

import (
	"context"
	"fmt"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DeleteContract(goCtx context.Context, msg *types.MsgDeleteContract) (*types.MsgDeleteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != val.Owner {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}
	if val.Status != types.ContractStatus_CONTRACT_STATUS_TERMINATED && 
	val.Status != types.ContractStatus_CONTRACT_STATUS_PENDING && 
	val.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in pending, forsale or terminated status to be deleted")
	}
	k.RemoveContract(ctx, msg.Id)

	return &types.MsgDeleteContractResponse{}, nil
}
