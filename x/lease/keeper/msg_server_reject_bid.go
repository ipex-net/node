package keeper

import (
	"context"

	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RejectBid(goCtx context.Context, msg *types.MsgRejectBid) (*types.MsgRejectBidResponse, error) {
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

	blocktime := ctx.BlockTime()

	// Refunding the bid
	bidderAdmin, found := k.orgKeeper.GetOrgAdmin(ctx, bid.BidderOrgId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrOrgAdminNotFound, "can not refund bid for contract %d - admin for org %d not found", bid.ContractId, bid.BidderOrgId)
	}
	coins := sdk.NewCoins(bid.Amount)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleAccountName,
		bidderAdmin,
		coins,
	); err != nil {
		return nil, errorsmod.Wrapf(types.ErrFundTransfer, "can not refund bid for contract %d and bidder org id %d - transaction failed with error %s", bid.ContractId, bid.BidderOrgId, err)
	}
	k.paymentsKeeper.RecordTransaction(ctx, paymentstypes.TransactionType_TRANSACTION_TYPE_BID_REFUNDED, paymentstypes.ModuleID_MODULE_ID_LEASE, bid.ContractId, bid.Amount, bid.OwnerOrgId, bid.BidderOrgId, "module:lease", bidderAdmin.String())

	bid.Status = types.BidStatus_BID_STATUS_REJECTED
	bid.UpdatedAt = &blocktime
	k.SetBid(ctx, bid)

	return &types.MsgRejectBidResponse{}, nil
}
