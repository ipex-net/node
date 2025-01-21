package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgUpdateSubnetGeolocation{}

func NewMsgUpdateSubnetGeolocation(creator string, id uint64, country string, region string, city string, postalCode string) *MsgUpdateSubnetGeolocation {
	return &MsgUpdateSubnetGeolocation{
		Creator:    creator,
		Id:         id,
		Country:    country,
		Region:     region,
		City:       city,
		PostalCode: postalCode,
	}
}

func (msg *MsgUpdateSubnetGeolocation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
