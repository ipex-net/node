package keeper

import (
	"context"
	"ipex/x/lease/types"
	"strconv"
	"time"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProcessEnrollmentTimeouts - processes all the contract in CONTRACT_STATUS_ENROLLED to reverse those in timeout
func (k Keeper) ProcessEnrollmentTimeouts(goCtx context.Context) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	//timeoutDuration := time.Hour * 72 // 72 hours timeout
	timeoutDuration := time.Minute * 2 // 72 hours timeout

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.ContractKey))

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()


	for ; iterator.Valid(); iterator.Next() {
		var contract types.Contract
		k.cdc.MustUnmarshal(iterator.Value(), &contract)

		// Check if the contract is in enrolled status
		if contract.Status != types.ContractStatus_CONTRACT_STATUS_ENROLLED {
			continue
		}

		// Check if the contract was enrolled more than 72 hours ago
		if contract.EnrolledAt.IsZero() {
			// Skip if enrolledAt is not set
			continue
		}

		if contract.EnrolledAt != nil && ctx.BlockTime().Sub(*contract.EnrolledAt) > timeoutDuration {
			//Processing timeout actions
			contract.Status = types.ContractStatus_CONTRACT_STATUS_TIMEDOUT
			contract.UpdatedAt = ctx.BlockTime()
			refundAddr, found := k.orgKeeper.GetOrgAdmin(ctx, contract.LesseeOrg)
		
			if !found {
				k.logger.Error("failed to find admin address for organization",
				"contract_id", contract.Id,
				"lessee_org_id", contract.LesseeOrg)
				continue
			} 
	
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, refundAddr, sdk.NewCoins(contract.Price)); err != nil {
				k.logger.Error("failed to release coins from escrow", 
					"contract_id", contract.Id,
					"lessee_org_id", contract.LesseeOrg,
					"amount", contract.Price,
					"refund_addr", refundAddr,
					"error", err)
				continue				
			}

			k.SetContract(ctx, contract)

			// Log successful processing
			k.logger.Info("contract enrollment timed out and processed successfully",
				"contract_id", contract.Id,
				"enrolled_at", contract.EnrolledAt,
				"refunded_amount", contract.Price,
				"refund_address", refundAddr)

			// Emit enrollment reverted event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EnrollementRevertedEventType,
					sdk.NewAttribute(types.EnrollementRevertedContractId, strconv.FormatUint(contract.Id, 10)),
					sdk.NewAttribute(types.EnrollementRevertedLesseeOrg, strconv.FormatUint(contract.LesseeOrg, 10)),
					sdk.NewAttribute(types.EnrollementRevertedPriceAmount, contract.Price.Amount.String()),
					sdk.NewAttribute(types.EnrollementRevertedPriceDenom, contract.Price.Denom),
				),
			)
		}
	}
}