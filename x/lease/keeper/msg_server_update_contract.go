package keeper

import (
	"context"
	"strconv"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UpdateContract(goCtx context.Context, msg *types.MsgUpdateContract) (*types.MsgUpdateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check that the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that contractId is not nil
	if msg.Id == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	// Check that the contract exists
	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.Id)
	}

	// Check the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be updated")
	}

	// Check that the sender is the admin of the owner org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect owner admin %s", msg.Creator)
	}

	// Check that price and safety deposit denoms match
	if contract.SafetyDeposit.Denom != msg.Price.Denom {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDeposit,
			"price denom (%s) must match safety deposit denom (%s)",
			msg.Price.Denom,
			contract.SafetyDeposit.Denom,
		)
	}

	// Check that safety deposit is greater than contract price
	if contract.SafetyDeposit.Amount.LT(msg.Price.Amount) {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDeposit,
			"safety deposit amount (%s) must be greater than or equal to new price (%s)",
			contract.SafetyDeposit.Amount.String(),
			msg.Price.Amount.String(),
		)
	}

	// Update contract
	contract.Price = msg.Price
	contract.UpdatedAt = ctx.BlockTime()
	k.SetContract(ctx, contract)

	// Emit event
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
