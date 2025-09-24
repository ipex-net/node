package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateContract{}

func NewMsgUpdateContract(creator string, price uint64, id uint64) *MsgUpdateContract {
	return &MsgUpdateContract{
		Creator: creator,
		Price:   price,
		Id:      id,
	}
}

func (msg *MsgUpdateContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
