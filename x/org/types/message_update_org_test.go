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
				Creator: "invalid_address",
				Id:      1,
				Name:    "Test Org",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid message",
			msg: MsgUpdateOrg{
				Creator:            validAddr,
				Id:                 1,
				Name:               "Test Org",
				LegalName:          "Test Organization LLC",
				Address:            "123 Test St",
				RegistrationNumber: "REG123",
				CountryCode:        "US",
				TaxId:              "TAX123",
				Website:            "https://testorg.com",
				AbuseEmail:         "abuse@testorg.com",
			},
		}, {
			name: "minimal valid message",
			msg: MsgUpdateOrg{
				Creator: validAddr,
				Id:      1,
				Name:    "Test Org",
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
