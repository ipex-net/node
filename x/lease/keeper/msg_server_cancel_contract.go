package keeper

import (
	"context"
	"strconv"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CancelContract(goCtx context.Context, msg *types.MsgCancelContract) (*types.MsgCancelContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	//Fetching the contract
	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Contract %d doesn't exist", msg.Id)
	}

	//Checking the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED {
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "can not cancel contract that is not in LEASED state")
	}

	//Checking that the sender is the admin of the lessee org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.LesseeOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect lessee admin %s", msg.Creator)
	}

	//Updating the contract
	contract.Status = types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE
	contract.UpdatedAt = ctx.BlockTime()

	k.SetContract(ctx, contract)

	// Emit the event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(types.ContractCanceledEventType,
            sdk.NewAttribute(types.ContractCanceledCreator, msg.Creator),
            sdk.NewAttribute(types.ContractCanceledContractId, strconv.FormatUint(msg.Id, 10)),
        ),
    )

	return &types.MsgCancelContractResponse{}, nil
}
