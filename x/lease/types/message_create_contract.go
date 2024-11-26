package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateContract{}

func NewMsgCreateContract(creator string, networkAddress string, prefixLength uint64, rir string, countryCode string, geofeedUrl string, price sdk.Coin, deposit sdk.Coin, noticePeriodBlocks uint64) *MsgCreateContract {
	return &MsgCreateContract{
		Creator:            creator,
		NetworkAddress:     networkAddress,
		PrefixLength:       prefixLength,
		Rir:                rir,
		CountryCode:        countryCode,
		GeofeedUrl:         geofeedUrl,
		Price:              price,
		Deposit:            deposit,
		NoticePeriodBlocks: noticePeriodBlocks,
	}
}

func (msg *MsgCreateContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
