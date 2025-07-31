package keeper_test

import (
	"testing"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/sample"
	"ipex/x/subnet/keeper"
	"ipex/x/subnet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgServer_CreateSubnet(t *testing.T) {
	k, ctx := keepertest.SubnetKeeper(t)
	srv := keeper.NewMsgServerImpl(k)
	wctx := sdk.WrapSDKContext(ctx)

	// Valid test data
	creator := sample.AccAddress()
	ownerOrgId := uint64(1001)
	networkAddress := uint64(0x08080800) // 8.8.8.0 (public IP range)
	mask := uint64(24)
	country := "US"

	t.Run("successful subnet creation", func(t *testing.T) {
		msg := &types.MsgCreateSubnet{
			Creator:        creator,
			OwnerOrgId:     ownerOrgId,
			NetworkAddress: networkAddress,
			Mask:           mask,
			Country:        country,
		}

		resp, err := srv.CreateSubnet(wctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Greater(t, resp.Id, uint64(0))

		// Verify subnet was created with country
		subnet, found := k.GetSubnet(ctx, resp.Id)
		require.True(t, found)
		require.Equal(t, networkAddress, subnet.NetworkAddress)
		require.Equal(t, mask, subnet.Mask)
		require.Equal(t, ownerOrgId, subnet.OwnerOrgId)
		require.Equal(t, country, subnet.Country)
		require.Equal(t, types.SubnetStatus_SUBNET_STATUS_INITIAL, subnet.Status)
	})

	t.Run("country validation tests", func(t *testing.T) {
		testCases := []struct {
			name        string
			country     string
			expectedErr string
		}{
			{
				name:        "empty country",
				country:     "",
				expectedErr: "country data can not be empty",
			},
			{
				name:        "single letter country",
				country:     "U",
				expectedErr: "country code must be 2 letters",
			},
			{
				name:        "three letter country",
				country:     "USA",
				expectedErr: "country code must be 2 letters",
			},
			{
				name:        "invalid country code",
				country:     "ZZ", // Use a definitely invalid code
				expectedErr: "invalid ISO-3166-1 alpha-2 country code",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				msg := &types.MsgCreateSubnet{
					Creator:        sample.AccAddress(),
					OwnerOrgId:     ownerOrgId,
					NetworkAddress: 0x09090900 + uint64(10*256), // 9.9.10.0, 9.9.20.0, etc. (valid subnet addresses)
					Mask:           mask,
					Country:        tc.country,
				}

				_, err := srv.CreateSubnet(wctx, msg)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})

	t.Run("other validation tests", func(t *testing.T) {
		testCases := []struct {
			name        string
			modifyMsg   func(*types.MsgCreateSubnet)
			expectedErr string
		}{
			{
				name: "invalid creator address",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.Creator = "invalid_address"
				},
				expectedErr: "invalid creator address",
			},
			{
				name: "zero owner org id",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.OwnerOrgId = 0
				},
				expectedErr: "org ID not provided",
			},
			{
				name: "reserved IP range - private 10.x",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.NetworkAddress = 0x0A000000 // 10.0.0.0
				},
				expectedErr: "IP address is in a reserved range",
			},
			{
				name: "reserved IP range - private 192.168.x",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.NetworkAddress = 0xC0A80000 // 192.168.0.0
				},
				expectedErr: "IP address is in a reserved range",
			},
			{
				name: "reserved IP range - loopback",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.NetworkAddress = 0x7F000000 // 127.0.0.0
				},
				expectedErr: "IP address is in a reserved range",
			},
			{
				name: "invalid mask - too small",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.Mask = 7
				},
				expectedErr: "prefix length must be between 8 and 24",
			},
			{
				name: "invalid mask - too large",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.Mask = 25
				},
				expectedErr: "prefix length must be between 8 and 24",
			},
			{
				name: "network address with non-zero last octet",
				modifyMsg: func(msg *types.MsgCreateSubnet) {
					msg.NetworkAddress = 0x08080801 // 8.8.8.1 (should be 8.8.8.0)
				},
				expectedErr: "fourth octet must be zero for subnet",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				msg := &types.MsgCreateSubnet{
					Creator:        creator,
					OwnerOrgId:     ownerOrgId,
					NetworkAddress: 0x03030300 + uint64(20*256), // 3.3.20.0, 3.3.40.0, etc. (valid subnet addresses)
					Mask:           mask,
					Country:        country,
				}

				tc.modifyMsg(msg)

				_, err := srv.CreateSubnet(wctx, msg)
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			})
		}
	})

	t.Run("duplicate network address validation", func(t *testing.T) {
		// First subnet creation should succeed
		msg1 := &types.MsgCreateSubnet{
			Creator:        creator,
			OwnerOrgId:     ownerOrgId,
			NetworkAddress: 0x04040400, // 4.4.4.0
			Mask:           mask,
			Country:        country,
		}

		resp1, err := srv.CreateSubnet(wctx, msg1)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Second subnet with same network address should fail
		msg2 := &types.MsgCreateSubnet{
			Creator:        sample.AccAddress(),
			OwnerOrgId:     ownerOrgId + 1,
			NetworkAddress: 0x04040400, // Same as above
			Mask:           mask,
			Country:        "CA",
		}

		_, err = srv.CreateSubnet(wctx, msg2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "subnet already exists")
	})

	t.Run("valid country codes acceptance", func(t *testing.T) {
		validCountries := []string{"US", "CA", "GB", "DE", "FR", "JP", "AU", "BR"}
		
		for i, countryCode := range validCountries {
			t.Run("country_"+countryCode, func(t *testing.T) {
				msg := &types.MsgCreateSubnet{
					Creator:        sample.AccAddress(),
					OwnerOrgId:     ownerOrgId,
					NetworkAddress: 0x05050500 + uint64(i*256), // 5.5.5.0, 5.5.6.0, etc.
					Mask:           mask,
					Country:        countryCode,
				}

				resp, err := srv.CreateSubnet(wctx, msg)
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Greater(t, resp.Id, uint64(0))

				// Verify country was stored correctly
				subnet, found := k.GetSubnet(ctx, resp.Id)
				require.True(t, found)
				require.Equal(t, countryCode, subnet.Country)
			})
		}
	})

	t.Run("lowercase country codes are accepted", func(t *testing.T) {
		msg := &types.MsgCreateSubnet{
			Creator:        sample.AccAddress(),
			OwnerOrgId:     ownerOrgId,
			NetworkAddress: 0x06060600, // 6.6.6.0
			Mask:           mask,
			Country:        "us", // lowercase
		}

		resp, err := srv.CreateSubnet(wctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Greater(t, resp.Id, uint64(0))

		// Verify country was stored correctly (as provided, not converted)
		subnet, found := k.GetSubnet(ctx, resp.Id)
		require.True(t, found)
		require.Equal(t, "us", subnet.Country) // Stored as provided
	})
}