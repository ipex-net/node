package keeper

import (
	"context"
	"strconv"
	"time"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RevokeContract(goCtx context.Context, msg *types.MsgRevokeContract) (*types.MsgRevokeContractResponse, error) {
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
		return nil, errorsmod.Wrap(types.ErrInvalidContractStatus, "can not revoke that is not in LEASED state")
	}

	//Checking that the sender is the admin of the owner org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect owner admin %s", msg.Creator)
	}

	//Updating the contract
	blockTime := ctx.BlockTime()
	//Getting the Revocation time from the subnet was prepaid until and Revocation Notice period
	//Rerouting subnet prior to this time involves a penalty and reputational damage
	revocationTime := contract.PrepaidUntil.Add(contract.RevocationPeriod)
	contract.Status = types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER
	contract.UpdatedAt = blockTime
	contract.RevocationEffectiveAt = &revocationTime
	k.SetContract(ctx, contract)

	// Emit the revocation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractRevokedEventType,
			sdk.NewAttribute(types.ContractRevokedCreator, msg.Creator),
			sdk.NewAttribute(types.ContractRevokedContractId, strconv.FormatUint(msg.Id, 10)),
			sdk.NewAttribute(types.ContractRevokedRevocationEffectiveAt, revocationTime.Format(time.RFC3339)),
		),
	)
	return &types.MsgRevokeContractResponse{}, nil
}
