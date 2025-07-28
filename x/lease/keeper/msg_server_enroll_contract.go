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

func (k msgServer) EnrollContract(goCtx context.Context, msg *types.MsgEnrollContract) (*types.MsgEnrollContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// NOTE: only org admin can pay for the contract. All refunds go to org admin.
	// We have treasury but it is only used to accumulate funds received from leasing subnets.
	// There are two scenarios for enrolement: automatic acceptance and manual bid selection

	// Check that the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check that contractId is not nil
	if msg.ContractId == 0 {
		return nil, types.ErrContractIdCantBeNil
	}

	// Check if the org id exists
	if msg.LesseeOrg != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.LesseeOrg) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "org with ID %d not found", msg.LesseeOrg)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.LesseeOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.LesseeOrg)
	}

	// Gett the contract
	contract, found := k.GetContract(ctx, msg.ContractId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInexistentContract, "contract with id %d doesn't exist", msg.ContractId)
	}

	// Check for matching org verification status
	if !contract.AllowUnverifiedBidders && !k.orgKeeper.IsOrgVerified(ctx, msg.LesseeOrg) {
		return nil, errorsmod.Wrapf(types.ErrOrgNotVerified, "contract only accepts verified orgs")
	}

	// Check the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be enrolled in")
	}

	// Check that user has enough funds:
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	userBalance := k.bankKeeper.SpendableCoins(ctx, creator)
	if !userBalance.IsAllGTE(sdk.NewCoins(contract.Price)) {
		return nil, errorsmod.Wrapf(types.ErrInsufficientFunds,
			"insufficient funds to enroll: required %s, available %s",
			contract.Price.String(), userBalance.String())
	}
	blockTime := ctx.BlockTime()

	// If the contract automatically accepts first bidder we skip entire bidding logic and proceed with contract activation
	if contract.AcceptFirstBidder {

		// Perform transaction
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleAccountName, sdk.NewCoins(contract.Price)); err != nil {
			k.Logger().Error("failed to send coins to escrow", "error", err)
			return nil, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer funds to escrow: %s", err)
		}
		// Log payment
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT, paymentstypes.ModuleID_MODULE_ID_LEASE, msg.ContractId, contract.Price, msg.LesseeOrg, contract.OwnerOrg, msg.Creator, "module:lease")

		// Update the contract:
		contract.Status = types.ContractStatus_CONTRACT_STATUS_LEASED
		contract.LesseeOrg = msg.LesseeOrg

		// Production:
		//prepaidUntil := blockTime.AddDate(0, 1, 0) // 0 years, 1 month, 0 days
		// Testing:
		prepaidUntil := blockTime.Add(time.Minute * 10)

		contract.UpdatedAt = blockTime
		contract.ActivatedAt = &blockTime
		depositRetained := contract.Price
		contract.SafetyDepositRetained = &depositRetained
		contract.PrepaidUntil = &prepaidUntil
		// Store updated contract
		k.SetContract(ctx, contract)

		// Initialize the lease reputation
		k.trustKeeper.InitializeLeaseReputation(ctx, contract.Id)

		// Update the subnet user Org
		k.subnetKeeper.SetSubnetUserOrg(ctx, contract.SubnetId, msg.LesseeOrg, contract.GeoChangeAllowed, contract.AllowASNDROPOriginTransfer)

		ipCount := k.subnetKeeper.GetSubnetIpCount(ctx, contract.SubnetId)
		// Log usage stats
		k.trustKeeper.LogLesseeContractActivation(ctx, msg.LesseeOrg, ipCount)
		// Emit enrollment event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.ContractEnrolledEventType,
				sdk.NewAttribute(types.ContractEnrolledCreator, msg.Creator),
				sdk.NewAttribute(types.ContractEnrolledContractId, strconv.FormatUint(msg.ContractId, 10)),
				sdk.NewAttribute(types.ContractEnrolledLesseeOrg, strconv.FormatUint(msg.LesseeOrg, 10)),
				sdk.NewAttribute(types.ContractEnrolledPriceAmount, contract.Price.Amount.String()),
				sdk.NewAttribute(types.ContractEnrolledPriceDenom, contract.Price.Denom),
			),
		)
	}
	// If first bid is not automatically accepter - use the bidding system
	if !contract.AcceptFirstBidder {
		// Bid interval validations
		if msg.BidExpirationHours == 0 || msg.BidExpirationHours > 120 {
			return nil, errorsmod.Wrapf(
				types.ErrInvalidBidExpirationInterval,
				"bid expiration interval %d should be between 1 and 120",
				msg.BidExpirationHours,
			)
		}

		// Intended use string validation
		if msg.IntendedUse == "" || len(msg.IntendedUse) < 3 || len(msg.IntendedUse) > 100 {
			return nil, errorsmod.Wrap(
				types.ErrIncorrectIntendedUseLength,
				"intended use string must be between 3 and 100 characters",
			)
		}

		// Check if bid already exists
		if _, found := k.GetBid(ctx, msg.ContractId, msg.LesseeOrg); found {
			return nil, errorsmod.Wrapf(types.ErrBidAlreadyExists, "bid for contract %d and org %d already exists", msg.ContractId, msg.LesseeOrg)
		}

		// Perform transaction
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleAccountName, sdk.NewCoins(contract.Price)); err != nil {
			k.Logger().Error("failed to send coins to escrow", "error", err)
			return nil, errorsmod.Wrapf(types.ErrFundTransfer, "failed to transfer funds to escrow: %s", err)
		}
		// Save payment
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_BID_SUBMITTED, paymentstypes.ModuleID_MODULE_ID_LEASE, msg.ContractId, contract.Price, msg.LesseeOrg, contract.OwnerOrg, msg.Creator, "module:lease")

		// Create the bid
		bid := types.Bid{
			ContractId:      msg.ContractId,
			BidderOrgId:     msg.LesseeOrg,
			OwnerOrgId:      contract.OwnerOrg,
			IntendedUse:     msg.IntendedUse,
			Status:          types.BidStatus_BID_STATUS_SUBMITTED,
			CreatedAt:       sdk.UnwrapSDKContext(ctx).BlockTime(),
			ExpirationHours: msg.BidExpirationHours,
			Amount:          contract.Price,
		}

		// Store the bid
		k.SetBid(ctx, bid)

		// Emit bid submitted event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.BidSubmittedEventType,
				sdk.NewAttribute(types.BidSubmittedCreator, msg.Creator),
				sdk.NewAttribute(types.BidSubmittedContractId, strconv.FormatUint(msg.ContractId, 10)),
				sdk.NewAttribute(types.BidSubmittedLesseeOrg, strconv.FormatUint(msg.LesseeOrg, 10)),
				sdk.NewAttribute(types.BidSubmittedPriceAmount, contract.Price.Amount.String()),
				sdk.NewAttribute(types.BidSubmittedPriceDenom, contract.Price.Denom),
				sdk.NewAttribute(types.BidSubmittedExpirationHours, strconv.FormatUint(msg.BidExpirationHours, 10)),
			),
		)
	}

	return &types.MsgEnrollContractResponse{}, nil
}
