package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgChangeOrgAdmin{}

func NewMsgChangeOrgAdmin(creator string, orgId uint64, newAdmin string) *MsgChangeOrgAdmin {
	return &MsgChangeOrgAdmin{
		Creator:  creator,
		OrgId:    orgId,
		NewAdmin: newAdmin,
	}
}

func (msg *MsgChangeOrgAdmin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
