package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateContract(goCtx context.Context, msg *types.MsgCreateContract) (*types.MsgCreateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//TODO: ALL VALIDATIONS
	// Check for existing active contracts with the same network address

	//TODO: PRICE ADJUSTEMENT

	var contract = types.Contract{
		SubnetId:     msg.SubnetId,
		OwnerOrg:       msg.OwnerOrg,
		LesseeOrg:          0,
		Price:              msg.Price,
		PriceAdjustementType: msg.PriceAdjustementType,
		NextPricedAdjustement: nil,
		SafetyDeposit: msg.SafetyDeposit,
		CancellationNoticePeriod: msg.CancellationNoticePeriod,
		CancellationNoticeSentAt: nil,
		Status: types.ContractStatus_CONTRACT_STATUS_FORSALE,
		CreatedAt: ctx.BlockTime(),
		UpdatedAt: ctx.BlockTime(),
	}
	id := k.AppendContract(
		ctx,
		contract,
	)


	//TODO: EMIT EVENTS
	return &types.MsgCreateContractResponse{
		Id: id,
	}, nil
}
