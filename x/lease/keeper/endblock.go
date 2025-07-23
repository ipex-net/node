package keeper

import (
	"context"
	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"
	"time"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TESTING
const gracePeriod = 2 * time.Minute

// ProcessExpiredContracts processes contracts that need status updates or termination
func (k Keeper) ProcessExpiredContracts(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// // TESTING
	// // Only run every 30 blocks
	// if ctx.BlockHeight()%30 != 0 {
	// 	return
	// 	}

	// Process contracts by status using efficient collections queries
	k.processContractsByStatus(ctx, types.ContractStatus_CONTRACT_STATUS_LEASED)
	k.processContractsByStatus(ctx, types.ContractStatus_CONTRACT_STATUS_EXPIRED)
	k.processContractsByStatus(ctx, types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE)
	k.processContractsByStatus(ctx, types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER)
}

// processContractsByStatus efficiently processes contracts with a specific status
func (k Keeper) processContractsByStatus(ctx sdk.Context, status types.ContractStatus) {
	// Use collections to query only contracts with the specific status
	rng := collections.NewPrefixedPairRange[int32, uint64](int32(status))

	err := k.ContractsByStatus.Walk(ctx, rng, func(key collections.Pair[int32, uint64], _ bool) (bool, error) {
		contractId := key.K2()
		contract, found := k.GetContract(ctx, contractId)
		if !found {
			return false, nil
		}

		switch status {
		case types.ContractStatus_CONTRACT_STATUS_LEASED:
			// If contract is in grace period we just change the status to expired
			if contract.PrepaidUntil.Before(ctx.BlockTime()) {
				contract.Status = types.ContractStatus_CONTRACT_STATUS_EXPIRED
				contract.UpdatedAt = ctx.BlockTime()
				k.SetContract(ctx, contract)
				k.logger.Info("contract expired",
					"contract_id", contract.Id,
				)
			}

		case types.ContractStatus_CONTRACT_STATUS_EXPIRED:
			// Check if grace period has passed for termination
			if contract.PrepaidUntil.Add(gracePeriod).Before(ctx.BlockTime()) {
				k.terminateContract(ctx, contract)
			}

		case types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE,
			types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER:
			// Check if prepaid period has passed for termination
			if contract.PrepaidUntil.Before(ctx.BlockTime()) {
				k.terminateContract(ctx, contract)
			}
		}

		return false, nil
	})

	if err != nil {
		k.logger.Error("failed to walk contracts by status", "status", status, "error", err)
	}
}

// terminateContract handles the termination logic for contracts
func (k Keeper) terminateContract(ctx sdk.Context, contract types.Contract) {
	// Release the safety deposit if any
	if contract.SafetyDepositRetained != nil {
		coins := sdk.NewCoins(*contract.SafetyDepositRetained)
		// Get the treasury
		refundAddress, found := k.orgKeeper.GetOrgTreasury(ctx, contract.OwnerOrg)

		// Fail over to orgAdmin if not found
		if !found {
			refundAddress, found = k.orgKeeper.GetOrgAdmin(ctx, contract.OwnerOrg)
		}
		if !found {
			k.logger.Error("can't locate treasury or admin address to process refund on contract termination",
				"contract_id", contract.Id,
				"refund_amount", contract.SafetyDepositRetained)
			return
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			refundAddress,
			coins,
		); err != nil {
			k.logger.Error("can't process refund on contract termination",
				"contract_id", contract.Id,
				"refund_amount", contract.SafetyDepositRetained)
			return
		}
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COMPENSATION, paymentstypes.ModuleID_MODULE_ID_LEASE, contract.Id, (*contract.SafetyDepositRetained), contract.LesseeOrg, contract.OwnerOrg, "module:lease", refundAddress.String())

		// Reset safety deposit
		contract.SafetyDepositRetained = nil
	}
	// Remove the user from the subnet
	k.subnetKeeper.RemoveSubnetUserOrg(ctx, contract.SubnetId)
	// Update the contract data
	currentTime := ctx.BlockTime()
	contract.TerminatedAt = &currentTime
	contract.UpdatedAt = currentTime
	contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED
	k.SetContract(ctx, contract)
	k.logger.Info("Contract terminated",
		"contract_id", contract.Id,
	)
}

// ProcessExpiredBids cancels and refunds all expired bids
func (k Keeper) ProcessExpiredBids(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Interval after expiration that will trigger bid cleanup
	var bidCleanupThresholdHours uint64 = 48

	// Only run every 50 blocks
	if ctx.BlockHeight()%50 != 0 {
		return
	}

	blocktime := ctx.BlockTime()
	bids := k.GetAllBids(ctx)
	for _, bid := range bids {

		// Treating expired bids
		if bid.Status == types.BidStatus_BID_STATUS_SUBMITTED && blocktime.After(bid.CreatedAt.Add(time.Duration(time.Hour*time.Duration(bid.ExpirationHours)))) {
			// Refunding the bid
			bidderAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, bid.BidderOrgId)
			if !found {
				k.logger.Warn("can not refund bid for contract %d - admin for org %d not found", bid.ContractId, bid.BidderOrgId)
				continue
			}
			coins := sdk.NewCoins(bid.Amount)
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				types.ModuleAccountName,
				bidderAdmin,
				coins,
			); err != nil {
				k.logger.Warn("can not refund bid for contract %d and bidder org id %d - transaction failed", bid.ContractId, bid.BidderOrgId)
				continue
			}
			k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_BID_REFUNDED, paymentstypes.ModuleID_MODULE_ID_LEASE, bid.ContractId, bid.Amount, bid.OwnerOrgId, bid.BidderOrgId, "module:lease", bidderAdmin.String())

			bid.Status = types.BidStatus_BID_STATUS_EXPIRED
			bid.UpdatedAt = &blocktime
			k.SetBid(ctx, bid)
		}

		// Remove old bids
		if bid.Status != types.BidStatus_BID_STATUS_SUBMITTED && blocktime.After(bid.CreatedAt.Add(time.Duration(time.Hour*time.Duration(bid.ExpirationHours+bidCleanupThresholdHours)))) {
			k.RemoveBid(ctx, bid.ContractId, bid.BidderOrgId)
		}
	}
}
