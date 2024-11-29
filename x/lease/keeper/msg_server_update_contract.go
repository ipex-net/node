package keeper

import (
	"context"
	"fmt"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateContract(goCtx context.Context, msg *types.MsgUpdateContract) (*types.MsgUpdateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != contract.Owner {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}
	if	contract.Status != types.ContractStatus_CONTRACT_STATUS_PENDING && 
	contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in pending or forsale state to be updated")
	}
	contract.Price = msg.Price
	contract.Deposit = msg.Deposit
	contract.NoticePeriodBlocks = msg.NoticePeriodBlocks
	k.SetContract(ctx, contract)
	return &types.MsgUpdateContractResponse{}, nil
}
