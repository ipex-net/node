package types

import (
	"testing"

	"ipex/testutil/sample"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateContract_ValidateBasic(t *testing.T) {
	validAddr := sample.AccAddress()
	validCoin := sdk.NewInt64Coin("stake", 100)
	validOwnerOrg := uint64(1)
	validSubnetId := uint64(1)
	validRevocationPeriod := uint64(6)

	tests := []struct {
		name string
		msg  MsgCreateContract
		err  error
	}{
		{
			name: "valid message",
			msg: MsgCreateContract{
				Creator:               validAddr,
				OwnerOrg:              validOwnerOrg,
				Price:                 validCoin,
				SafetyDeposit:         validCoin,
				RevocationPeriodMonth: validRevocationPeriod,
				SubnetId:              validSubnetId,
			},
			err: nil,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateContract{
				Creator:               "invalid_address",
				OwnerOrg:              validOwnerOrg,
				Price:                 validCoin,
				SafetyDeposit:         validCoin,
				RevocationPeriodMonth: validRevocationPeriod,
				SubnetId:              validSubnetId,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
