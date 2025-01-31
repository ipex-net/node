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

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be updated")
	}

	//Checking that the sender is the admin of the owner org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect owner admin %s", msg.Creator)
	}


	contract.Price = msg.Price
	contract.UpdatedAt = ctx.BlockTime()
	k.SetContract(ctx, contract)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractUpdatedEventType,
			sdk.NewAttribute(types.ContractUpdatedEventCreator, msg.Creator),
			sdk.NewAttribute(types.ContractUpdatedEventContractId, strconv.FormatUint(msg.Id, 10)),
			sdk.NewAttribute(types.ContractUpdatedEventPrice, msg.Price.Amount.String()),
			sdk.NewAttribute(types.ContractUpdatedEventPriceDenom, msg.Price.Denom),
		),
	)
	return &types.MsgUpdateContractResponse{}, nil
}
