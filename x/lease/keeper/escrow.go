package keeper

import (
	"context"
	"ipex/x/lease/types"
	"time"

	storetypes "cosmossdk.io/store/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendToEscrow sends coins from an account to the module account
func (k Keeper) SendToEscrow(ctx context.Context, from sdk.AccAddress, amount sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, from, types.ModuleAccountName, amount); err != nil {
		k.Logger().Error("failed to send coins to escrow", "error", err)
		return errorsmod.Wrapf(types.ErrFundTransfer, "failed to send %s from %s to escrow: %s",
			amount, from, err)
	}

	k.Logger().Info("coins sent to escrow", "from", from, "amount", amount)

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"escrow_deposit",
			sdk.NewAttribute("sender", from.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// ReleaseFromEscrow sends coins from the module account to an account
func (k Keeper) ReleaseFromEscrow(ctx context.Context, to sdk.AccAddress, amount sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, to, amount); err != nil {
		k.Logger().Error("failed to release coins from escrow", "error", err)
		return errorsmod.Wrapf(types.ErrFundTransfer, "failed to send %s from escrow to %s: %s",
			amount, to, err)
	}

	k.Logger().Info("coins released from escrow", "to", to, "amount", amount)

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"escrow_release",
			sdk.NewAttribute("receiver", to.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// ProcessEnrollmentTimeouts - processes all the contract in CONTRACT_STATUS_ENROLLED to reverse those in timeout
func (k Keeper) ProcessEnrollmentTimeouts(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	//timeoutDuration := time.Hour * 72 // 72 hours timeout
	timeoutDuration := time.Minute // 72 hours timeout

	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(sdkCtx))
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

		if sdkCtx.BlockTime().Sub(contract.EnrolledAt) > timeoutDuration {
			//Processing timeout actions
			contract.Status = types.ContractStatus_CONTRACT_STATUS_TIMEDOUT
			contract.UpdatedAt = sdkCtx.BlockTime()
			refundAddr, found := k.orgKeeper.GetOrgTreasury(ctx, contract.LesseeOrg)
		
			if !found {
				k.logger.Error("failed to find treasury address for organization",
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
		}

	
	}
}