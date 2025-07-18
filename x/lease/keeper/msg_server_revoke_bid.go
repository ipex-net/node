package keeper

import (
	"context"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) RevokeBid(goCtx context.Context, msg *types.MsgRevokeBid) (*types.MsgRevokeBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if the org id exists
	if msg.BidderOrgId != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.BidderOrgId) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "org with ID %d not found", msg.BidderOrgId)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is BidderOrgId admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.BidderOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.BidderOrgId)
	}

	bid, found := k.GetBid(ctx, msg.ContractId, msg.BidderOrgId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrBidNotFound, "bid not found for contract id %d and bidder org id %d", msg.ContractId, msg.BidderOrgId)
	}
	// Bid must be in submitted status
	if bid.Status != types.BidStatus_BID_STATUS_SUBMITTED {
		return nil, errorsmod.Wrapf(types.ErrIncorrectBidStatus, "bid must be in submitted status")
	}

	// Refunding

	return &types.MsgRevokeBidResponse{}, nil
}
