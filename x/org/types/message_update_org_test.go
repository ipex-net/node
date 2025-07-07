package types

import (
	"testing"

	"ipex/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateOrg_ValidateBasic(t *testing.T) {
	validAddr := sample.AccAddress()
	
	tests := []struct {
		name string
		msg  MsgUpdateOrg
		err  error
	}{
		{
			name: "invalid creator address",
			msg: MsgUpdateOrg{
				Creator:  "invalid_address",
				Id:       1,
				Name:     "Test Org",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid message",
			msg: MsgUpdateOrg{
				Creator:  validAddr,
				Id:       1,
				Name:     "Test Org",
				Admin:    validAddr,
				Treasury: validAddr,
			},
		}, {
			name: "invalid admin address",
			msg: MsgUpdateOrg{
				Creator: validAddr,
				Id:      1,
				Name:    "Test Org",
				Admin:   "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid treasury address",
			msg: MsgUpdateOrg{
				Creator:  validAddr,
				Id:       1,
				Name:     "Test Org",
				Treasury: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "empty admin and treasury are valid",
			msg: MsgUpdateOrg{
				Creator: validAddr,
				Id:      1,
				Name:    "Test Org",
				Admin:   "",
				Treasury: "",
			},
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
