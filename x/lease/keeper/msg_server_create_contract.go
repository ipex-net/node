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

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Check if the org id exists
	if msg.OwnerOrg != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.OwnerOrg) {
			return nil, errorsmod.Wrapf(ErrInvalidOrgId, "org with ID %d not found", msg.OwnerOrg)
		}
	} else {
		return nil, errorsmod.Wrap(ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.OwnerOrg, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.OwnerOrg)
	}

	// Verify creator if the admin of the owner of subnet
	if !k.subnetKeeper.VerifySubnetOwner(goCtx, msg.SubnetId, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotSubnetOwnerAdmin, "address %s is not an admin of the owner org %d of the subnet", msg.Creator, msg.OwnerOrg)
	}

	//Verify subnet is available
	if !k.subnetKeeper.CheckSubnetLeaseAvailability(goCtx, msg.SubnetId) {
		return nil, errorsmod.Wrap(ErrSubnetNotReady, "subnet is not ready for lease")
	}
	deposit := msg.SafetyDeposit
	var contract = types.Contract{
		SubnetId:              msg.SubnetId,
		OwnerOrg:              msg.OwnerOrg,
		LesseeOrg:             0,
		Price:                 msg.Price,
		SafetyDeposit:         &deposit,
		RevocationPeriod:      msg.RevocationPeriod,
		RevocationEffectiveAt: nil,
		Status:                types.ContractStatus_CONTRACT_STATUS_FORSALE,
		CreatedAt:             ctx.BlockTime(),
		UpdatedAt:             ctx.BlockTime(),
		EnrolledAt:			   ctx.BlockTime(),
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
			sdk.NewAttribute(types.ContractCreatedEventRevocationPeriodDuration, msg.RevocationPeriod.String()),
		),
	)

	return &types.MsgCreateContractResponse{
		Id: id,
	}, nil
}
