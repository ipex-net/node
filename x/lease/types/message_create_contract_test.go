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
	
	tests := []struct {
		name string
		msg  MsgCreateContract
		err  error
	}{
		{
			name: "valid message",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     "192.168.1.1",
				PrefixLength:       24,
				Rir:               "ARIN",
				CountryCode:        "US",
				GeofeedUrl:        "https://example.com/geofeed.csv",
				Price:             validCoin,
				Deposit:           validCoin,
				NoticePeriodBlocks: 100,
			},
			err: nil,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateContract{
				Creator: "invalid_address",
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid network address format",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "invalid.ip.address",
				PrefixLength: 24,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "prefix length too low",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 7,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "prefix length too high",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 25,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty RIR",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "unknown RIR",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "RIPES",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid country code format",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "USA",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "zero price",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "US",
				Price: sdk.NewInt64Coin("stake", 0),
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "zero notice period",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "US",
				Price: validCoin,
				Deposit: validCoin,
				NoticePeriodBlocks: 0,
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "empty geofeed URL",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "US",
				Price: validCoin,
				Deposit: validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl: "",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid geofeed URL format",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "US",
				Price: validCoin,
				Deposit: validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl: "not-a-url",
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "non-HTTPS geofeed URL",
			msg: MsgCreateContract{
				Creator: validAddr,
				NetworkAddress: "192.168.1.1",
				PrefixLength: 24,
				Rir: "ARIN",
				CountryCode: "US",
				Price: validCoin,
				Deposit: validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl: "http://example.com/geofeed.csv",
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