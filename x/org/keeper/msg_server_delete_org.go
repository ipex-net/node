package keeper

import (
	"context"

	"ipex/x/org/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) DeleteOrg(goCtx context.Context, msg *types.MsgDeleteOrg) (*types.MsgDeleteOrgResponse, error) {

	// Message not implemented yet
	return  &types.MsgDeleteOrgResponse{}, errorsmod.Wrap(types.ErrNotImplemented, ("delete org is not implemented"))
	/*
	ctx := sdk.UnwrapSDKContext(goCtx)
	contract, found := k.GetOrg(ctx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(types.ErrOrgNotFound, fmt.Sprintf("org with id %d not found", msg.Id))
	}
	if msg.Creator != contract.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect org creator")
	}
	active, err := k.trustKeeper.LessorHasActiveContracts(ctx, msg.Id)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrOrgNotFound, "can not get count of active contracts")
	}
	if active {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "can not delete org with active contracts")
	}

	// TODO: check if there are subnets with users

	// TODO: check org balance before removing
	// TODO: last contract cooldown period before removing
	k.RemoveOrg(ctx, msg.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.OrgDeletedEventType,
			sdk.NewAttribute(types.OrgDeletedEventCreator, msg.Creator),
			sdk.NewAttribute(types.OrgDeletedEventContractId, strconv.FormatUint(msg.Id, 10)),
		),
	)

	return &types.MsgDeleteOrgResponse{}, nil
	*/
}
