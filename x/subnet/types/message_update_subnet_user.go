package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateSubnetUser{}

func NewMsgUpdateSubnetUser(creator string, id uint64, userOrgId uint64) *MsgUpdateSubnetUser {
	return &MsgUpdateSubnetUser{
		Creator:   creator,
		Id:        id,
		UserOrgId: userOrgId,
	}
}

func (msg *MsgUpdateSubnetUser) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
