package keeper

import (
	"context"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
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
