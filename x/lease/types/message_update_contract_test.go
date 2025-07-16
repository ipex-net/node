package types

import (
	"testing"

	"ipex/testutil/sample"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateContract_ValidateBasic(t *testing.T) {
	validCoin := sdk.NewInt64Coin("stake", 100)
	zeroCoin := sdk.NewInt64Coin("stake", 0)

	tests := []struct {
		name string
		msg  MsgUpdateContract
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateContract{
				Creator: "invalid_address",
				Price:   validCoin,
				Id:      1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateContract{
				Creator: sample.AccAddress(),
				Price:   validCoin,
				Id:      1,
			},
		}, {
			name: "zero price",
			msg: MsgUpdateContract{
				Creator: sample.AccAddress(),
				Price:   zeroCoin,
				Id:      1,
			},
			err: sdkerrors.ErrInvalidRequest,
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
