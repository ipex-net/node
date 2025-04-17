package keeper

import (
	"context"
	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"
	"time"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const gracePeriodDays = 7
const gracePeriodDaysTest = 2

// TerminateCancelledContracts terminates expired contracts
func (k Keeper) ProcessExpiredContracts(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var contract types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)
		// Terminating the contract and refunding the safety deposit in the following cases:
		// contract was cancelled and prepaid period has passed
		// contract expired and grace period has passed
		//
		if ((contract.Status == types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE || contract.Status == types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER) && contract.PrepaidUntil.Before(ctx.BlockTime())) ||
			//Production
			// (contract.Status == types.ContractStatus_CONTRACT_STATUS_EXPIRED && contract.PrepaidUntil.AddDate(0, 0, gracePeriodDays).Before(ctx.BlockTime())) {
			//Testing
			(contract.Status == types.ContractStatus_CONTRACT_STATUS_EXPIRED && contract.PrepaidUntil.Add(time.Minute*gracePeriodDaysTest).Before(ctx.BlockTime())) {
			//releasing the Safetydeposit if any
			if contract.SafetyDepositRetained != nil {
				coins := sdk.NewCoins(*contract.SafetyDepositRetained)
				//Getting the treasury
				refundAddress, found := k.orgKeeper.GetOrgTreasury(ctx, contract.OwnerOrg)

				//Failing over to orgAdmin if not found
				if !found {
					refundAddress, found = k.orgKeeper.GetOrgAdmin(ctx, contract.OwnerOrg)
				}
				if !found {
					k.logger.Error("can't locate treasury or admin address to process refund on contract termination",
						"contract_id", contract.Id,
						"refund_amount", contract.SafetyDepositRetained)
					continue
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

					continue
				}
				k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND, paymentstypes.ModuleID_MODULE_ID_LEASE, contract.Id, (*contract.SafetyDepositRetained), contract.LesseeOrg, contract.OwnerOrg, "module:lease", refundAddress.String())

				//Ressetting safety deposit
				contract.SafetyDepositRetained = nil
			}
			//Removing the user from the subnet
			k.subnetKeeper.RemoveSubnetUserOrg(ctx, contract.SubnetId)
			//Updating the contract data
			currentTime := ctx.BlockTime()
			contract.TerminatedAt = &currentTime
			contract.UpdatedAt = currentTime
			contract.Status = types.ContractStatus_CONTRACT_STATUS_TERMINATED
			k.SetContract(ctx, contract)
			k.logger.Info("Contract terminated",
				"contract_id", contract.Id,
			)
		}
		// If contract is in grace period we just change the status to expired
		if contract.Status == types.ContractStatus_CONTRACT_STATUS_LEASED && contract.PrepaidUntil.Before(ctx.BlockTime()) {
			contract.Status = types.ContractStatus_CONTRACT_STATUS_EXPIRED
			contract.UpdatedAt = ctx.BlockTime()
			k.SetContract(ctx, contract)
			k.logger.Info("contract expired",
				"contract_id", contract.Id,
			)
		}
	}
}
