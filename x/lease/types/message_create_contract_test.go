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
	validSubnet := "212.1.1.0"
	invalidSubnet1 := "192.168.0.0"
	invalidSubnet2 := "212.1.1."

	tests := []struct {
		name string
		msg  MsgCreateContract
		err  error
	}{
		{
			name: "valid message",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     validSubnet,
				PrefixLength:       24,
				Rir:                "ARIN",
				CountryCode:        "US",
				GeofeedUrl:         "https://example.com/geofeed.csv",
				Price:              validCoin,
				Deposit:            validCoin,
				NoticePeriodBlocks: 100,
			},
			err: nil,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateContract{
				Creator:        "invalid_address",
				NetworkAddress: validSubnet,
				PrefixLength:   24,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "reserved IP range",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: invalidSubnet1,
				PrefixLength:   24,
			},
			err: ErrReservedIPRange,
		},
		{
			name: "invalid network address format",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: invalidSubnet2,
				PrefixLength:   24,
			},
			err: ErrInvalidIPFormat,
		},
		{
			name: "prefix length too low",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   7,
			},
			err: ErrInvalidPrefixLength,
		},
		{
			name: "prefix length too high",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   25,
			},
			err: ErrInvalidPrefixLength,
		},
		{
			name: "empty RIR",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   24,
				Rir:            "",
			},
			err: ErrEmptyRIR,
		},
		{
			name: "unknown RIR",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   24,
				Rir:            "RIPES",
			},
			err: ErrUnknownRIR,
		},
		{
			name: "invalid country code format",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   24,
				Rir:            "ARIN",
				CountryCode:    "USA",
			},
			err: ErrInvalidCountryCode,
		},
		{
			name: "zero price",
			msg: MsgCreateContract{
				Creator:        validAddr,
				NetworkAddress: validSubnet,
				PrefixLength:   24,
				Rir:            "ARIN",
				CountryCode:    "US",
				Price:          sdk.NewInt64Coin("stake", 0),
			},
			err: ErrInvalidPrice,
		},
		{
			name: "zero notice period",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     validSubnet,
				PrefixLength:       24,
				Rir:                "ARIN",
				CountryCode:        "US",
				Price:              validCoin,
				Deposit:            validCoin,
				NoticePeriodBlocks: 0,
			},
			err: ErrInvalidNoticePeriod,
		},
		{
			name: "empty geofeed URL",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     validSubnet,
				PrefixLength:       24,
				Rir:                "ARIN",
				CountryCode:        "US",
				Price:              validCoin,
				Deposit:            validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl:         "",
			},
			err: ErrInvalidGeofeedURL,
		},
		{
			name: "invalid geofeed URL format",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     validSubnet,
				PrefixLength:       24,
				Rir:                "ARIN",
				CountryCode:        "US",
				Price:              validCoin,
				Deposit:            validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl:         "not-a-url",
			},
			err: ErrInvalidGeofeedURL,
		},
		{
			name: "non-HTTPS geofeed URL",
			msg: MsgCreateContract{
				Creator:            validAddr,
				NetworkAddress:     validSubnet,
				PrefixLength:       24,
				Rir:                "ARIN",
				CountryCode:        "US",
				Price:              validCoin,
				Deposit:            validCoin,
				NoticePeriodBlocks: 100,
				GeofeedUrl:         "http://example.com/geofeed.csv",
			},
			err: ErrInvalidGeofeedURL,
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
