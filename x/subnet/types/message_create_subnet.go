package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateSubnet{}

func NewMsgCreateSubnet(creator string, ownerOrgId uint64, networkAddress uint64, mask uint64, country string) *MsgCreateSubnet {
	return &MsgCreateSubnet{
		Creator:        creator,
		OwnerOrgId:     ownerOrgId,
		NetworkAddress: networkAddress,
		Mask:           mask,
		Country:        country,
	}
}

func (msg *MsgCreateSubnet) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
