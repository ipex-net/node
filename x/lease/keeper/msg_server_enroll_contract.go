package keeper

import (
	"context"
	"strconv"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) EnrollContract(goCtx context.Context, msg *types.MsgEnrollContract) (*types.MsgEnrollContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	//NOTE: only org admin can pay for the contract. All refunds go to org admin.
	//We have treasury but it is only used to accumulate funds received from leasing subnets.

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if the org id exists
	if msg.LesseeOrg != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.LesseeOrg) {
			return nil, errorsmod.Wrapf(ErrInvalidOrgId, "org with ID %d not found", msg.LesseeOrg)
		}
	} else {
		return nil, errorsmod.Wrap(ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.LesseeOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.LesseeOrg)
	}

	//Getting the contract
	contract, found := k.GetContract(ctx, msg.Contract)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.Contract)
	}

	//Checking the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be enrolled in")
	}

	//Checking that user has enough funds:
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	userBalance := k.bankKeeper.SpendableCoins(ctx, creator)
	if !userBalance.IsAllGTE(sdk.NewCoins(contract.Price)) {
		return nil, errorsmod.Wrapf(ErrInsufficientFunds,
			"insufficient funds to enroll: required %s, available %s",
			contract.Price.String(), userBalance.String())
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleAccountName, sdk.NewCoins(contract.Price)); err != nil {
		k.Logger().Error("failed to send coins to escrow", "error", err)
		return nil, errorsmod.Wrapf(ErrFundTransfer, "failed to transfer funds to escrow: %s", err)
	}

	//Updating the contract:
	contract.Status = types.ContractStatus_CONTRACT_STATUS_LEASED
	contract.LesseeOrg = msg.LesseeOrg
	blockTime := ctx.BlockTime()
	prepaidUntil := blockTime.AddDate(0, 1, 0) // 0 years, 1 month, 0 days
	contract.UpdatedAt = blockTime
	contract.ActivatedAt = &blockTime
	depositRetained := contract.Price
	contract.SafetyDepositRetained = &depositRetained
	contract.PrepaidUntil = &prepaidUntil
	// Store updated contract
	k.SetContract(ctx, contract)

	// Updating the subnet user Org
	k.subnetKeeper.SetSubnetUserOrg(ctx, contract.SubnetId, msg.LesseeOrg)

	// Emit enrollment event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ContractEnrolledEventType,
			sdk.NewAttribute(types.ContractEnrolledCreator, msg.Creator),
			sdk.NewAttribute(types.ContractEnrolledContractId, strconv.FormatUint(msg.Contract, 10)),
			sdk.NewAttribute(types.ContractEnrolledLesseeOrg, strconv.FormatUint(msg.LesseeOrg, 10)),
			sdk.NewAttribute(types.ContractEnrolledPriceAmount, contract.Price.Amount.String()),
			sdk.NewAttribute(types.ContractEnrolledPriceDenom, contract.Price.Denom),
		),
	)

	return &types.MsgEnrollContractResponse{}, nil
}
