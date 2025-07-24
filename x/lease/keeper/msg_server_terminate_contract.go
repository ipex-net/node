package keeper

import (
	"context"
	"strconv"
	"time"

	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"

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

	// Can only terminate contracts that are leased, cancelled or expired
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED && contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE && contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER && contract.Status != types.ContractStatus_CONTRACT_STATUS_EXPIRED {
		return nil, types.ErrInvalidContractStatus
	}

	// Determine if Lessor eligible for termination without penalty
	noPenalty := k.CanTerminateWithoutPenalty(ctx, msg.Id)


	// Checking that the sender is the admin of the owner org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect owner admin %s", msg.Creator)
	}

	// Getting the lessee org admin
	lesseeAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, contract.LesseeOrg)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrOrgAdminNotFound, "admin for org %d not found", contract.LesseeOrg)
	}

	// Transfer Safety Deposit to lessee only if there is a deposit and noPenalty is false
	if contract.SafetyDepositRetained != nil && !noPenalty {
		coins := sdk.NewCoins(*contract.SafetyDepositRetained)

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			lesseeAdmin,
			coins,
		); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer safety deposit to lessee org admin: %s", err)
		}
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COMPENSATION, paymentstypes.ModuleID_MODULE_ID_LEASE, contract.Id, (*contract.SafetyDepositRetained), contract.LesseeOrg, contract.OwnerOrg, "module:lease", lesseeAdmin.String())

		contract.SafetyDepositRetained = nil
	}

	blockTime := ctx.BlockTime()
	oneMonthAfterActivation := contract.ActivatedAt.AddDate(0, 1, 0)

	var statuslog string

	// Determine the appropriate termination status based on timing
	if noPenalty {
		// Contract terminated during the first month
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_WITH_LESSEE_CONTRACT_VIOLATION
		statuslog = "terminated_without_lessor_penalty"
	} else if blockTime.Before(oneMonthAfterActivation) {
		// Contract terminated during the first month
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_FIRST_MONTH
		statuslog = "terminated_first_month"

	} else {
		// Contract terminated after the first month
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_WITHOUT_NOTICE
		statuslog = "terminated_without_notice"
	}
	contract.UpdatedAt = blockTime
	contract.TerminatedAt = &blockTime
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
