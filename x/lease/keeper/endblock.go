package keeper

import (
	"context"
	"ipex/x/lease/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExecuteLeaseEndBlock performs lease-related operations at the end of each block
func (k Keeper) TerminateContractsCancelledByLessee(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx) // Important step to unwrap SDK context
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var contract types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)

		if contract.Status!=types.ContractStatus_CONTRACT_STATUS_CANCELED_LESSEE || contract.PrepaidUntil.After(ctx.BlockTime()) {continue}
		//Terminating contracts that are cancelled by lessee and expired

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
				"refund_amount", contract.SafetyDepositRetained,)
				continue
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(
				ctx,
				types.ModuleAccountName,
				refundAddress, // Add the recipient address here
				coins,
			); err != nil {
				k.logger.Error("can't process refund on contract termination",
				"contract_id", contract.Id,
				"refund_amount", contract.SafetyDepositRetained,)
				continue
			}
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
		k.logger.Info("contract cancelled by lessee terminated on expiration",
		"contract_id", contract.Id,
		)
	}
}

func (k Keeper) ExpireUnpaidContracts(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx) // Important step to unwrap SDK context
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var contract types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)
		if contract.Status == types.ContractStatus_CONTRACT_STATUS_LEASED && contract.PrepaidUntil.After(ctx.BlockTime()) {
			contract.Status = types.ContractStatus_CONTRACT_STATUS_EXPIRED
			contract.UpdatedAt = ctx.BlockTime()
			k.SetContract(ctx, contract)
		}
		
	}
}