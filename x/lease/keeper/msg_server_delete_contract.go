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

func (k msgServer) DeleteContract(goCtx context.Context, msg *types.MsgDeleteContract) (*types.MsgDeleteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}
	if msg.Creator != contract.Owner {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_TERMINATED &&
		contract.Status != types.ContractStatus_CONTRACT_STATUS_PENDING &&
		contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "contract must be in pending, forsale or terminated status to be deleted")	
	}
	err := k.RemoveContract(ctx, msg.Id)

	if err == nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.ContractDeletedEventType,
				sdk.NewAttribute(types.ContractDeletedEventCreator, msg.Creator),
				sdk.NewAttribute(types.ContractDeletedEventContractId, strconv.FormatUint(msg.Id, 10)),
			),
		)
	}

	return &types.MsgDeleteContractResponse{}, nil
}
