package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddFollowup{}

func NewMsgAddFollowup(creator string, complaintId uint64, content string, evidenceHash []string) *MsgAddFollowup {
	return &MsgAddFollowup{
		Creator:      creator,
		ComplaintId:  complaintId,
		Content:      content,
		EvidenceHash: evidenceHash,
	}
}

func (msg *MsgAddFollowup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
