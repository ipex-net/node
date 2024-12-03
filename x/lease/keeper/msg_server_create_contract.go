package keeper

import (
	"context"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateContract(goCtx context.Context, msg *types.MsgCreateContract) (*types.MsgCreateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// First do basic validation
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Check for existing active contracts with the same network address
	contracts := k.GetAllContract(ctx)
	for _, existingContract := range contracts {
		if existingContract.NetworkAddress == msg.NetworkAddress && 
		   existingContract.Status != types.ContractStatus_CONTRACT_STATUS_TERMINATED {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"contract already exists for network address %s and is not terminated",
				msg.NetworkAddress,
			)
		}
	}

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
		NoticeSentBlock:       0,
		Status:                types.ContractStatus_CONTRACT_STATUS_FORSALE,
		StatusChangeBlock:     uint64(ctx.BlockHeight()),
		Lessee:                "",
		CreationBlock:         uint64(ctx.BlockHeight()),
		ActivationBlock:       0,
		PrepaidUntilBlock:     0,
		AsNumber:              "",
	}
	id := k.AppendContract(
		ctx,
		contract,
	)
	return &types.MsgCreateContractResponse{
		Id: id,
	}, nil
}
