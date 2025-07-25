package keeper

import (
	"context"
	"strconv"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) UndoCancelContract(goCtx context.Context, msg *types.MsgUndoCancelContract) (*types.MsgUndoCancelContractResponse, error) {
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

	// Fetch the contract
	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Contract %d doesn't exist", msg.Id)
	}

	// Check the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE {
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "can not undo cancellation of contract that is not in CANCELED_LESSEE state")
	}

	// Check if contract was prepaid
	if contract.PrepaidUntil == nil {
		return nil, errorsmod.Wrap(types.ErrContractExpired, "Contract has no prepaid period set")
	}

	// Check that it hasn'e expired
	if contract.PrepaidUntil.Before(ctx.BlockTime()) {
		return nil, errorsmod.Wrap(types.ErrContractExpired, "Cannot undo cancellation for expired contract")
	}

	// Check that the sender is the admin of the lessee org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.LesseeOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect lessee admin %s", msg.Creator)
	}

	// Set back to LEASED state
	contract.Status = types.ContractStatus_CONTRACT_STATUS_LEASED
	contract.UpdatedAt = ctx.BlockTime()

	k.SetContract(ctx, contract)

	// Emit the event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractCancellationRevokedEventType,
			sdk.NewAttribute(types.ContractCancellationRevokedCreator, msg.Creator),
			sdk.NewAttribute(types.ContractCancellationRevokedContractId, strconv.FormatUint(msg.Id, 10)),
		),
	)

	return &types.MsgUndoCancelContractResponse{}, nil
}
