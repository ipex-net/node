package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateSubnetOrigin{}

func NewMsgUpdateSubnetOrigin(creator string, id uint64, as uint64) *MsgUpdateSubnetOrigin {
	return &MsgUpdateSubnetOrigin{
		Creator: creator,
		Id:      id,
		As:      as,
	}
}

func (msg *MsgUpdateSubnetOrigin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
