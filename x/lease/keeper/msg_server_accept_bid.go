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

func (k msgServer) AcceptBid(goCtx context.Context, msg *types.MsgAcceptBid) (*types.MsgAcceptBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)


	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if the bidder org exists
	if msg.BidderOrgId != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.BidderOrgId) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "bidder org with ID %d not found", msg.BidderOrgId)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "bidder org ID not provided")
	}

	bid, found := k.GetBid(ctx, msg.ContractId, msg.BidderOrgId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrBidNotFound, "bid not found for contract id %d and bidder org id %d", msg.ContractId, msg.BidderOrgId)
	}

	// Bid must be in submitted status
	if bid.Status != types.BidStatus_BID_STATUS_SUBMITTED {
		return nil, errorsmod.Wrapf(types.ErrIncorrectBidStatus, "bid must be in submitted status")
	}

	// Check that bid haven't expired
	if ctx.BlockTime().After(bid.CreatedAt.Add(time.Duration(bid.ExpirationHours) * time.Hour)) {
		return nil, errorsmod.Wrapf(types.ErrBidExpired, "bid has expired")
 	
	}
	// Check if the owner org exists
	if bid.OwnerOrgId != 0 {
		if !k.orgKeeper.HasOrg(ctx, bid.OwnerOrgId) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "owner org with ID %d not found", bid.OwnerOrgId)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "owner org ID not provided")
	}

	// Verify creator is OwnerOrgId admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, bid.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, bid.OwnerOrgId)
	}

	// Enrolling in contract
	// Getting the contract
	contract, found := k.GetContract(ctx, msg.ContractId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "key %d doesn't exist", msg.ContractId)
	}

	// Checking the contract status
	if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "contract must be in forsale status to be enrolled in")
	}

	// Updating the contract
	blockTime := ctx.BlockTime()
	contract.Status = types.ContractStatus_CONTRACT_STATUS_LEASED
	contract.LesseeOrg = bid.BidderOrgId

	// Production:
	//prepaidUntil := blockTime.AddDate(0, 1, 0) // 0 years, 1 month, 0 days
	// Testing:
	prepaidUntil := blockTime.Add(time.Minute * 10)
	contract.UpdatedAt = blockTime
	contract.ActivatedAt = &blockTime
	depositRetained := bid.Amount
	contract.SafetyDepositRetained = &depositRetained
	contract.PrepaidUntil = &prepaidUntil

	// Store updated contract
	k.SetContract(ctx, contract)

	// Initializing lease reputation
	k.trustKeeper.InitializeLeaseReputation(ctx, contract.Id)

	// Updating the subnet user Org
	k.subnetKeeper.SetSubnetUserOrg(ctx, contract.SubnetId, bid.BidderOrgId)

	// Removing accepted bid
	k.RemoveBid(ctx, bid.ContractId, bid.BidderOrgId)


	// Refunding and deleting lost bids
	bids := k.GetBidsByContract(ctx, bid.ContractId)

	for _, lostbid := range bids {
		if lostbid.BidderOrgId == bid.BidderOrgId {
			continue
		}
		if lostbid.Status != types.BidStatus_BID_STATUS_SUBMITTED {
			// If it is already rejected or revoked, we just remove it
			k.RemoveBid(ctx, lostbid.ContractId, lostbid.BidderOrgId)
			continue
		}
		lostBidAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, lostbid.BidderOrgId)
		if !found {
			return nil, errorsmod.Wrapf(types.ErrOrgAdminNotFound, "can not refund bid for contract %d - admin for org %d not found", lostbid.ContractId, lostbid.BidderOrgId)
		}
		coins := sdk.NewCoins(lostbid.Amount)
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleAccountName,
			lostBidAdmin,
			coins,
		); err != nil {
			return nil, errorsmod.Wrapf(types.ErrFundTransfer, "can not refund bid for contract %d and bidder org id %d - transaction failed with error %s", lostbid.ContractId, lostbid.BidderOrgId, err)
		}
		k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_BID_REFUNDED, paymentstypes.ModuleID_MODULE_ID_LEASE, lostbid.ContractId, lostbid.Amount, lostbid.OwnerOrgId, lostbid.BidderOrgId, "module:lease", lostBidAdmin.String())

		k.RemoveBid(ctx, lostbid.ContractId, lostbid.BidderOrgId)
	}

	// Emit enrollment event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ContractEnrolledEventType,
			sdk.NewAttribute(types.ContractEnrolledCreator, msg.Creator),
			sdk.NewAttribute(types.ContractEnrolledContractId, strconv.FormatUint(bid.ContractId, 10)),
			sdk.NewAttribute(types.ContractEnrolledLesseeOrg, strconv.FormatUint(bid.BidderOrgId, 10)),
			sdk.NewAttribute(types.ContractEnrolledPriceAmount, bid.Amount.Amount.String()),
			sdk.NewAttribute(types.ContractEnrolledPriceDenom, bid.Amount.Denom),
		),
	)

	return &types.MsgAcceptBidResponse{}, nil
}
