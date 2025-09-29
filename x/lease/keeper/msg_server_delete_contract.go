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

	// Check that the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that contractId is not nil
	if msg.Id == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	// Fetch contract
	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	// Only FORSALE contracts can be deleted
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "contract must be in forsale status")
	}

	// Only admin of owner org can delete contract
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect contract owner")
	}

	// Set subnet status to INITIAL
	err = k.subnetKeeper.SetSubnetActive(ctx, contract.SubnetId, false)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrSubnetNotReady, "can not set subnet status to initial (%s)", err)
	}

	// Remove contract
	err = k.RemoveContract(ctx, msg.Id)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "failed to remove contract %d: %s", msg.Id, err)
	}

	k.trustKeeper.RemoveLeaseReputation(ctx, msg.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractDeletedEventType,
			sdk.NewAttribute(types.ContractDeletedEventCreator, msg.Creator),
			sdk.NewAttribute(types.ContractDeletedEventContractId, strconv.FormatUint(msg.Id, 10)),
		),
	)

	return &types.MsgDeleteContractResponse{}, nil
}
