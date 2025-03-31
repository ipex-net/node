package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateContract{}

func NewMsgCreateContract(creator string, ownerOrg uint64, price sdk.Coin, safetyDeposit sdk.Coin, revocationPeriodMonth uint64, subnetId uint64) *MsgCreateContract {
	return &MsgCreateContract{
		Creator:               creator,
		OwnerOrg:              ownerOrg,
		Price:                 price,
		SafetyDeposit:         safetyDeposit,
		RevocationPeriodMonth: revocationPeriodMonth,
		SubnetId:              subnetId,
	}
}

func (msg *MsgCreateContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
