package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateContract{}

func NewMsgUpdateContract(creator string, price sdk.Coin, deposit sdk.Coin, noticePeriodBlocks uint64, id uint64) *MsgUpdateContract {
	return &MsgUpdateContract{
		Creator:            creator,
		Price:              price,
		Deposit:            deposit,
		NoticePeriodBlocks: noticePeriodBlocks,
		Id:                 id,
	}
}

func (msg *MsgUpdateContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !msg.Price.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	if msg.NoticePeriodBlocks == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "notice period must be positive")
	}
	return nil
}
