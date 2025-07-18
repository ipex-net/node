package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRevokeBid{}

func NewMsgRevokeBid(creator string, contractId uint64, bidderOrgId uint64) *MsgRevokeBid {
	return &MsgRevokeBid{
		Creator:     creator,
		ContractId:  contractId,
		BidderOrgId: bidderOrgId,
	}
}

func (msg *MsgRevokeBid) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
