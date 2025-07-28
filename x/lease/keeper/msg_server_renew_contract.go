package keeper

import (
	"context"
	"strconv"
	"time"

	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RenewContract(goCtx context.Context, msg *types.MsgRenewContract) (*types.MsgRenewContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Percent amount of retained renewal payment
	const retainer int = 10

	// NOTE: only org admin can renew for the contract. All refunds go to org admin as well.

	// Check that the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that contractId is not nil
	if msg.ContractId == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	contract, found := k.GetContract(ctx, msg.ContractId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.ContractId)
	}

	// Check the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED && contract.Status != types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER && contract.Status != types.ContractStatus_CONTRACT_STATUS_EXPIRED {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in leased, canceled by owner or expired status to be renewed")
	}

	// Only allow renewal if the PrepaidUntil date is withtin 1 month
	if contract.PrepaidUntil.After(ctx.BlockTime().AddDate(0, 1, 0)) {
		return nil, errorsmod.Wrapf(types.ErrRenewalDateTooDistant, "contract renewal date %s is too far in the future", contract.PrepaidUntil)
	}

	// Check that if we renew the contract we won't go beyond the revocation date
	if contract.RevocationEffectiveAt != nil && contract.PrepaidUntil.AddDate(0, 1, 0).After(*contract.RevocationEffectiveAt) {
		return nil, errorsmod.Wrap(types.ErrRenewalBeyondRevocationDate, "renewal period extends beyond contract revocation effective date")
	}

	// Check if the lessee org exists
	if contract.LesseeOrg != 0 {
		if !k.orgKeeper.HasOrg(ctx, contract.LesseeOrg) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "Lessee org with ID %d not found", contract.LesseeOrg)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "Contract doesn't have lessee org")
	}

	// Check that the contract owner has treasury account setup
	ownerTreasury, found := k.orgKeeper.GetOrgTreasury(ctx, contract.OwnerOrg)
	if !found {
		return nil, errorsmod.Wrap(types.ErrOrgTreasuryNotFound, "owner org doesn't have treasury")
	}

	// Verify creator is lessee org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, contract.LesseeOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, contract.LesseeOrg)
	}

	// Check that user has enough funds:
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	userBalance := k.bankKeeper.SpendableCoins(ctx, creator)
	if !userBalance.IsAllGTE(sdk.NewCoins(contract.Price)) {
		return nil, errorsmod.Wrapf(types.ErrInsufficientFunds,
			"insufficient funds to renew: required %s, available %s",
			contract.Price.String(), userBalance.String())
	}

	// Calculate the amounts to retain and to send to owner
	depositTargetAmount := contract.SafetyDeposit.Amount

	var depositRetainedAmount math.Int

	if contract.SafetyDepositRetained != nil {
		depositRetainedAmount = contract.SafetyDepositRetained.Amount
	} else {
		depositRetainedAmount = math.ZeroInt()
	}

	// By default pay everything to owner
	renewalFeeAmount := contract.Price.Amount
	payoutOwnerAmount := renewalFeeAmount
	payoutEscrowAmount := math.ZeroInt()

	// Only retain funds if contract haven't met the safety deposit target yet
	if depositRetainedAmount.LT(depositTargetAmount) {

		// Calculate the amount needed to reach the deposit target
		depositRemainingAmount := depositTargetAmount.Sub(depositRetainedAmount)

		// Calculate retained amount of the renewal fee
		retainerAmount := renewalFeeAmount.MulRaw(int64(retainer)).QuoRaw(100)

		// Determine the amount to add to escrow (min of what's remaining to collect and retainer amount)
		payoutEscrowAmount = math.MinInt(depositRemainingAmount, retainerAmount)

		// Subtract the escrow amount from owner payout
		payoutOwnerAmount = renewalFeeAmount.Sub(payoutEscrowAmount)
	}

	// Convert amounts back to Coins for the payout
	payoutOwner := sdk.NewCoin(contract.Price.Denom, payoutOwnerAmount)

	// Collect the retainer to module account
	if payoutEscrowAmount != math.ZeroInt() {
		payoutEscrow := sdk.NewCoin(contract.Price.Denom, payoutEscrowAmount)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleAccountName, sdk.NewCoins(payoutEscrow)); err != nil {
			k.Logger().Error("failed to transfer funds to safety deposit", "error", err)
			return nil, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer funds to safety deposit: %s", err)
		}

		// Log transaction
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT, paymentstypes.ModuleID_MODULE_ID_LEASE, msg.ContractId, payoutEscrow, contract.LesseeOrg, contract.OwnerOrg, creator.String(), "module:lease")

		depositRetainedAmount = depositRetainedAmount.Add(payoutEscrowAmount)

		if contract.SafetyDepositRetained == nil {
			contract.SafetyDepositRetained = &sdk.Coin{Denom: contract.Price.Denom, Amount: depositRetainedAmount}
		} else {
			contract.SafetyDepositRetained.Amount = depositRetainedAmount
		}

	}

	// Send the payout to owner's treasury
	if err := k.bankKeeper.SendCoins(ctx, creator, ownerTreasury, sdk.NewCoins(payoutOwner)); err != nil {
		k.Logger().Error("failed to send coins to escrow", "error", err)
		return nil, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer funds to escrow: %s", err)
	}
	k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT, paymentstypes.ModuleID_MODULE_ID_LEASE, msg.ContractId, payoutOwner, contract.LesseeOrg, contract.OwnerOrg, creator.String(), ownerTreasury.String())

	// Update the contract:
	blockTime := ctx.BlockTime()

	// Production:
	//prepaidUntil := contract.PrepaidUntil.AddDate(0, 1, 0) // 0 years, 1 month, 0 days
	// Testing:
	prepaidUntil := contract.PrepaidUntil.Add(time.Minute)

	// Reactivate expired contract that is still in grace period and therefore not terminated
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_LEASED {
		contract.Status = types.ContractStatus_CONTRACT_STATUS_LEASED
	}
	contract.UpdatedAt = blockTime
	contract.PrepaidUntil = &prepaidUntil

	// Store updated contract
	k.SetContract(ctx, contract)

	// Log usage stats
	ipCount := k.subnetKeeper.GetSubnetIpCount(ctx, contract.SubnetId)
	k.trustKeeper.LogLesseeContractRenewal(ctx, contract.LesseeOrg, ipCount)

	// Emit renewal event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ContractRenewedEventType,
			sdk.NewAttribute(types.ContractRenewedCreator, msg.Creator),
			sdk.NewAttribute(types.ContractRenewedContractId, strconv.FormatUint(msg.ContractId, 10)),
			sdk.NewAttribute(types.ContractRenewedLesseeOrg, strconv.FormatUint(contract.LesseeOrg, 10)),
			sdk.NewAttribute(types.ContractRenewedPriceAmount, contract.Price.Amount.String()),
			sdk.NewAttribute(types.ContractRenewedPriceDenom, contract.Price.Denom),
			sdk.NewAttribute(types.ContractRenewedPrepaidUntil, prepaidUntil.Format(time.RFC3339)),
			sdk.NewAttribute(types.ContractRenewedDepositRetained, contract.SafetyDepositRetained.Amount.String()),
		),
	)

	return &types.MsgRenewContractResponse{}, nil
}
