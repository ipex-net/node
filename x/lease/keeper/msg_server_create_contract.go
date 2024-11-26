package keeper

import (
	"context"

	"ipex/x/lease/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateContract(goCtx context.Context, msg *types.MsgCreateContract) (*types.MsgCreateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var contract = types.Contract{
		Id:                    0,
		NetworkAddress:        msg.NetworkAddress,
		PrefixLength:          msg.PrefixLength,
		Rir:                   msg.Rir,
		CountryCode:           msg.CountryCode,
		GeofeedUrl:            msg.GeofeedUrl,
		Owner:                 msg.Creator,
		Price:                 msg.Price,
		Deposit:               msg.Deposit,
		NoticePeriodBlocks:    msg.NoticePeriodBlocks,
		NoticeSentBlock:      0,
		Status:                0,
		StatusChangeBlock: 	 uint64(ctx.BlockHeight()),
		Lessee:                "",
		CreationBlock:         uint64(ctx.BlockHeight()),
		ActivationBlock:       0,
		PrepaidUntilBlock:     0,
		UserAbuseContactEmail: "",
		UserAbuseContactPhone: "",
	}
	id := k.AppendContract(
		ctx,
		contract,
	)
	return &types.MsgCreateContractResponse{
		Id: id,
	}, nil
}
