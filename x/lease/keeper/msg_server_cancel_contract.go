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

func (k msgServer) CancelContract(goCtx context.Context, msg *types.MsgCancelContract) (*types.MsgCancelContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED {
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "contract must be in pending, forsale or terminated status to be deleted")
	}

	if msg.Creator != contract.Owner && msg.Creator != contract.Lessee {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if msg.Creator == contract.Owner {
		contract.Status = types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER
		contract.StatusChangeBlock = uint64(ctx.BlockHeight())
		contract.NoticeSentBlock = uint64(ctx.BlockHeight())
		k.SetContract(ctx, contract)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.ContractCanceledByOwnerEventType,
				sdk.NewAttribute(types.ContractCanceledByOwnerCreator, msg.Creator),
				sdk.NewAttribute(types.ContractCanceledByOwnerContractId, strconv.FormatUint(msg.Id, 10)),
			),
		)
	}

	if msg.Creator == contract.Lessee {
		contract.Status = types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE
		contract.StatusChangeBlock = uint64(ctx.BlockHeight())
		contract.NoticeSentBlock = uint64(ctx.BlockHeight())
		k.SetContract(ctx, contract)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.ContractCanceledByLesseeEventType,
				sdk.NewAttribute(types.ContractCanceledByLesseeCreator, msg.Creator),
				sdk.NewAttribute(types.ContractCanceledByLesseeContractId, strconv.FormatUint(msg.Id, 10)),
			),
		)
	}

	return &types.MsgCancelContractResponse{}, nil
}
