package keeper

import (
	"context"
	"fmt"
	"strconv"

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
	// if msg.Creator != contract.Owner {
	// 	return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	// }
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be updated")
	}
	contract.Price = msg.Price

	k.SetContract(ctx, contract)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractUpdatedEventType,
			sdk.NewAttribute(types.ContractUpdatedEventCreator, msg.Creator),
			sdk.NewAttribute(types.ContractUpdatedEventContractId, strconv.FormatUint(msg.Id, 10)),
			sdk.NewAttribute(types.ContractUpdatedEventPrice, msg.Price.Amount.String()),
			sdk.NewAttribute(types.ContractUpdatedEventPriceDenom, msg.Price.Denom),
			// sdk.NewAttribute(types.ContractUpdatedEventDepositAmount, msg.Deposit.Amount.String()),
			// sdk.NewAttribute(types.ContractUpdatedEventNoticePeriodBlocks, strconv.FormatUint(msg.NoticePeriodBlocks, 10)),
		),
	)
	return &types.MsgUpdateContractResponse{}, nil
}
