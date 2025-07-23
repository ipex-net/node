package keeper

import (
	"context"
	"strconv"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateContract(goCtx context.Context, msg *types.MsgCreateContract) (*types.MsgCreateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Minimum revocation period
	minimumRevocationPeriod := uint64(1)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// l2BlacklistCountThreshold can not be zero
	if msg.L2BlacklistCountThreshold == 0 {
		return nil, types.ErrInvalidBlacklistThreshold
	}

	// Checking that the price is not nil
	if msg.Price.Amount.IsZero() {
		return nil, types.ErrInvalidPrice
	}
	// Add this check after verifying the subnet availability
	if msg.SafetyDeposit.Denom != msg.Price.Denom {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDeposit,
			"safety deposit denom (%s) must match price denom (%s)",
			msg.SafetyDeposit.Denom,
			msg.Price.Denom,
		)
	}

	if msg.SafetyDeposit.Amount.LT(msg.Price.Amount) {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidDeposit,
			"safety deposit amount (%s) must be greater than or equal to price (%s)",
			msg.SafetyDeposit.Amount.String(),
			msg.Price.Amount.String(),
		)
	}

	if msg.RevocationPeriodMonth < minimumRevocationPeriod {
		return nil, errorsmod.Wrapf(
			types.ErrInvalidRevocationPeriod,
			"revocation period (%d) must be at least (%d) month(s) ",
			msg.RevocationPeriodMonth,
			minimumRevocationPeriod,
		)
	}

	// Check if the org id exists
	if msg.OwnerOrg != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.OwnerOrg) {
			return nil, errorsmod.Wrapf(types.ErrInvalidOrgId, "org with ID %d not found", msg.OwnerOrg)
		}
	} else {
		return nil, errorsmod.Wrap(types.ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.OwnerOrg)
	}

	// Verify creator if the admin of the owner of subnet
	if !k.subnetKeeper.VerifySubnetOwner(goCtx, msg.SubnetId, msg.Creator) {
		return nil, errorsmod.Wrapf(types.ErrNotSubnetOwnerAdmin, "address %s is not an admin of the owner org %d of the subnet", msg.Creator, msg.OwnerOrg)
	}

	// Check that the contract creator has treasury account setup
	_, found := k.orgKeeper.GetOrgTreasury(ctx, msg.OwnerOrg)
	if !found {
		return nil, errorsmod.Wrap(types.ErrOrgTreasuryNotFound, "owner org doesn't have treasury")
	}

	//Verify subnet is available
	if !k.subnetKeeper.CheckSubnetLeaseAvailability(goCtx, msg.SubnetId) {
		return nil, types.ErrSubnetNotReady
	}

	var contract = types.Contract{
		SubnetId:                   msg.SubnetId,
		OwnerOrg:                   msg.OwnerOrg,
		LesseeOrg:                  0,
		Price:                      msg.Price,
		SafetyDeposit:              msg.SafetyDeposit,
		RevocationPeriodMonth:      msg.RevocationPeriodMonth,
		RevocationEffectiveAt:      nil,
		Status:                     types.ContractStatus_CONTRACT_STATUS_FORSALE,
		CreatedAt:                  ctx.BlockTime(),
		UpdatedAt:                  ctx.BlockTime(),
		GeoChangeAllowed:           msg.GeoChangeAllowed,
		L2BlacklistCountThreshold:  msg.L2BlacklistCountThreshold,
		AcceptFirstBidder:          msg.AcceptFirstBidder,
		AllowASNDROPOriginTransfer: msg.AllowASNDROPOriginTransfer,
		AllowUnverifiedBidders:     msg.AllowUnverifiedBidders,
	}
	id := k.AppendContract(
		ctx,
		contract,
	)

	// Emit the event with all attributes
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.ContractCreatedEventType,
			sdk.NewAttribute(types.ContractCreatedEventCreator, msg.Creator),
			sdk.NewAttribute(types.ContractCreatedEventContractId, strconv.FormatUint(id, 10)),
			sdk.NewAttribute(types.ContractCreatedEventSubnetId, strconv.FormatUint(msg.SubnetId, 10)),
			sdk.NewAttribute(types.ContractCreatedEventPrice, msg.Price.Amount.String()),
			sdk.NewAttribute(types.ContractCreatedEventPriceDenom, msg.Price.Denom),
			sdk.NewAttribute(types.ContractCreatedEventDepositAmount, msg.SafetyDeposit.Amount.String()),
			sdk.NewAttribute(types.ContractCreatedEventDepositDenom, msg.SafetyDeposit.Denom),
			sdk.NewAttribute(types.ContractCreatedEventRevocationPeriodDuration, strconv.FormatUint(msg.RevocationPeriodMonth, 10)),
		),
	)

	return &types.MsgCreateContractResponse{
		Id: id,
	}, nil
}
