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

func (k msgServer) RequestLesseeRefund(goCtx context.Context, msg *types.MsgRequestLesseeRefund) (*types.MsgRequestLesseeRefundResponse, error) {

	// Check that the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return &types.MsgRequestLesseeRefundResponse{Success: false}, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that contractId is not nil
	if msg.ContractId == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Fetch the contract
	contract, found := k.GetContract(ctx, msg.ContractId)
	if !found {
		return &types.MsgRequestLesseeRefundResponse{Success: false}, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Contract %d doesn't exist", msg.ContractId)
	}

	// Check the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED && contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE && contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER {
		return &types.MsgRequestLesseeRefundResponse{Success: false}, types.ErrInvalidContractStatus
	}

	// Check that the sender is the admin of the lessee org
	if !k.orgKeeper.VerifyOrgAdmin(ctx, contract.LesseeOrg, msg.Creator) {
		return &types.MsgRequestLesseeRefundResponse{Success: false}, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "incorrect lessee admin %s", msg.Creator)
	}

	// Check that contract is not expired yet
	if contract.PrepaidUntil.Before(ctx.BlockTime()) {
		return &types.MsgRequestLesseeRefundResponse{Success: false}, types.ErrContractExpired
	}

	// Get linked subnet
	if k.subnetKeeper.IsUserRefundAllowed(ctx, contract.SubnetId) {
		// Get lessee org admin address
		// NB lessee org admin validation logic may change, so we want to be sure to send refund to correct address
		lesseeAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, contract.LesseeOrg)
		if !found {
			return &types.MsgRequestLesseeRefundResponse{Success: false}, errorsmod.Wrapf(types.ErrOrgAdminNotFound, "admin for org %d not found", contract.LesseeOrg)
		}

		// Instantiate refund amount in coins
		coins := sdk.NewCoins(*contract.SafetyDepositRetained)

		// Process refund
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			lesseeAdmin,
			coins,
		); err != nil {
			return &types.MsgRequestLesseeRefundResponse{Success: false}, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer safety deposit to lessee org admin: %s", err)
		}

		// Record transaction in log
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COMPENSATION, paymentstypes.ModuleID_MODULE_ID_LEASE, contract.Id, (*contract.SafetyDepositRetained), contract.LesseeOrg, contract.OwnerOrg, "module:lease", lesseeAdmin.String())

		// Update contract
		blockTime := ctx.BlockTime()
		contract.SafetyDepositRetained = nil
		contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED_FAILED_ORIGIN_TRANSFER
		contract.UpdatedAt = blockTime
		contract.TerminatedAt = &blockTime
		k.SetContract(ctx, contract)

		// Remove user from the subnet
		k.subnetKeeper.RemoveSubnetUserOrg(ctx, contract.SubnetId)

		ipCount := k.subnetKeeper.GetSubnetIpCount(ctx, contract.SubnetId)
		k.trustKeeper.LogLesseeContractTermination(ctx, contract.LesseeOrg, false, ipCount)
		// Emit refund event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(types.LesseeRefundEmittedEventType,
				sdk.NewAttribute(types.LesseeRefundEmittedCreator, msg.Creator),
				sdk.NewAttribute(types.LesseeRefundEmittedContractId, strconv.FormatUint(contract.Id, 10)),
				sdk.NewAttribute(types.LesseeRefundEmittedAt, contract.UpdatedAt.Format(time.RFC3339)),
			),
		)
		return &types.MsgRequestLesseeRefundResponse{Success: true}, nil
	}

	return &types.MsgRequestLesseeRefundResponse{Success: false}, types.ErrRefundNotAllowed
}
