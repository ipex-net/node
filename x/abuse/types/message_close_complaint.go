package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCloseComplaint{}

func NewMsgCloseComplaint(creator string, complaintId uint64) *MsgCloseComplaint {
	return &MsgCloseComplaint{
		Creator:     creator,
		ComplaintId: complaintId,
	}
}

func (msg *MsgCloseComplaint) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
