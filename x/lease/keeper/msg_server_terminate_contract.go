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

// TerminateContract can only be called by contract owner, leads to immediate contract termination and release of all funds retained in the safety deposit to the lessee
func (k msgServer) TerminateContract(goCtx context.Context, msg *types.MsgTerminateContract) (*types.MsgTerminateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Fetching the contract
	contract, found := k.GetContract(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Contract %d doesn't exist", msg.Id)
	}

	// Checking the contract status
	// Even if the contract was cancelled, premature termination will lead to safety deposit transfer to lessee
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED && 
	contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE && 
	contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER {
		return nil, errorsmod.Wrapf(types.ErrInvalidContractStatus, "can not terminate contract in %s state", contract.Status)
	}

	// Checking that the sender is the admin of the owner org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect owner admin %s", msg.Creator)
	}

	// Getting the lessee org admin
	lesseeAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, contract.LesseeOrg)
	if !found {
		return nil, errorsmod.Wrapf(ErrOrgAdminNotFound, "admin for org %d not found", contract.LesseeOrg)
	}

	// Checking if there's safety deposit to refund and that revocationEffectiveAt is in future or is not set
	// If revocation date already passed the contract will be terminated automatically in endblocker func, not here
	if contract.SafetyDepositRetained != nil && (contract.RevocationEffectiveAt == nil || contract.RevocationEffectiveAt.After(ctx.BlockTime()))  {
		coins := sdk.NewCoins(*contract.SafetyDepositRetained)

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx, 
			types.ModuleAccountName, 
			lesseeAdmin, // Add the recipient address here
			coins,
		); err != nil {
			return nil, errorsmod.Wrapf(ErrFundTransfer, "failed to transfer safety deposit to lessee org admin: %s", err)
		}
	
	}

	//Updating the contract
	blockTime := ctx.BlockTime()
	oneMonthAfterActivation := contract.ActivatedAt.AddDate(0, 1, 0)

	var statuslog string
	// Determine the appropriate termination status based on timing
	if blockTime.Before(oneMonthAfterActivation) {
		// Contract terminated during the first month
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_FIRST_MONTH
		statuslog = "terminated_first_month"

	} else {
		// Contract terminated after the first month
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_WITHOUT_NOTICE
		statuslog = "terminated_without_notice"
	}
	contract.UpdatedAt = blockTime
	k.SetContract(ctx, contract)

	//Removing the user from the subnet
	k.subnetKeeper.RemoveSubnetUserOrg(ctx, contract.SubnetId)

	// Emit the termination event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractTerminatedByOwnerEventType,
			sdk.NewAttribute(types.ContractTerminatedByOwnerCreator, msg.Creator),
			sdk.NewAttribute(types.ContractTerminatedByOwnerContractId, strconv.FormatUint(contract.Id, 10)),
			sdk.NewAttribute(types.ContractDeletedEventType, statuslog),
			sdk.NewAttribute(types.ContractTerminatedByOwnerAt, contract.UpdatedAt.Format(time.RFC3339)),
		),
	)
	return &types.MsgTerminateContractResponse{}, nil
}
