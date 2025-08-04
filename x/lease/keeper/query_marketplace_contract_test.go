package keeper_test

import (
	"context"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/x/lease/keeper"
	"ipex/x/lease/types"
)

// createMarketplaceContracts creates test contracts with diverse properties for marketplace testing
func createMarketplaceContracts(keeper keeper.Keeper, ctx context.Context) []types.Contract {
	contracts := []types.Contract{
		// Contract 1: FORSALE, /24, ARIN, US, 8.8.8.0, $100
		{
			Id:             1,
			SubnetId:       101,
			OwnerOrg:       201,
			Price:          sdk.NewInt64Coin("ipex", 100),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 50),
			Mask:           24,
			NetworkAddress: 0x08080800, // 8.8.8.0
			Rir:            "ARIN",
			Country:        "US",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
		// Contract 2: FORSALE, /16, RIPE, DE, 10.0.0.0, $200
		{
			Id:             2,
			SubnetId:       102,
			OwnerOrg:       202,
			Price:          sdk.NewInt64Coin("ipex", 200),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 100),
			Mask:           16,
			NetworkAddress: 0x0A000000, // 10.0.0.0
			Rir:            "RIPE",
			Country:        "DE",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
		// Contract 3: LEASED (should not appear in marketplace)
		{
			Id:             3,
			SubnetId:       103,
			OwnerOrg:       203,
			LesseeOrg:      303,
			Price:          sdk.NewInt64Coin("ipex", 150),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 75),
			Mask:           24,
			NetworkAddress: 0x08080900, // 8.8.9.0
			Rir:            "ARIN",
			Country:        "US",
			Status:         types.ContractStatus_CONTRACT_STATUS_LEASED,
		},
		// Contract 4: FORSALE, /8, APNIC, JP, 192.168.1.0, $300
		{
			Id:             4,
			SubnetId:       104,
			OwnerOrg:       204,
			Price:          sdk.NewInt64Coin("ipex", 300),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 150),
			Mask:           8,
			NetworkAddress: 0xC0A80100, // 192.168.1.0
			Rir:            "APNIC",
			Country:        "JP",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
		// Contract 5: FORSALE, /24, LACNIC, BR, 172.16.0.0, $50
		{
			Id:             5,
			SubnetId:       105,
			OwnerOrg:       205,
			Price:          sdk.NewInt64Coin("ipex", 50),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 25),
			Mask:           24,
			NetworkAddress: 0xAC100000, // 172.16.0.0
			Rir:            "LACNIC",
			Country:        "BR",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
		// Contract 6: FORSALE, /24, AFRINIC, ZA, 41.0.0.0, $80
		{
			Id:             6,
			SubnetId:       106,
			OwnerOrg:       206,
			Price:          sdk.NewInt64Coin("ipex", 80),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 40),
			Mask:           24,
			NetworkAddress: 0x29000000, // 41.0.0.0
			Rir:            "AFRINIC",
			Country:        "ZA",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
		// Contract 7: CANCELED (should not appear in marketplace)
		{
			Id:             7,
			SubnetId:       107,
			OwnerOrg:       207,
			Price:          sdk.NewInt64Coin("ipex", 120),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 60),
			Mask:           24,
			NetworkAddress: 0x08080A00, // 8.8.10.0
			Rir:            "ARIN",
			Country:        "US",
			Status:         types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER,
		},
		// Contract 8: FORSALE, /24, ARIN, CA, 8.8.11.0, $90 (same first octet as contract 1)
		{
			Id:             8,
			SubnetId:       108,
			OwnerOrg:       208,
			Price:          sdk.NewInt64Coin("ipex", 90),
			SafetyDeposit:  sdk.NewInt64Coin("ipex", 45),
			Mask:           24,
			NetworkAddress: 0x08080B00, // 8.8.11.0
			Rir:            "ARIN",
			Country:        "CA",
			Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
		},
	}

	for _, contract := range contracts {
		keeper.SetContract(ctx, contract)
	}

	return contracts
}

func TestMarketplaceContract(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)
	contracts := createMarketplaceContracts(keeper, ctx)

	// Helper function to create requests
	request := func(mask uint64, rir, country string, minprice, maxprice, networkAddress, firstOctet uint64) *types.QueryMarketplaceContractRequest {
		return &types.QueryMarketplaceContractRequest{
			Mask:           mask,
			Rir:            rir,
			Country:        country,
			Minprice:       minprice,
			Maxprice:       maxprice,
			NetworkAddress: networkAddress,
			FirstOctet:     firstOctet,
		}
	}

	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.MarketplaceContract(ctx, nil)
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("MaskValidation", func(t *testing.T) {
		testCases := []struct {
			name        string
			mask        uint64
			expectError bool
			errorMsg    string
		}{
			{
				name:        "valid mask 24",
				mask:        24,
				expectError: false,
			},
			{
				name:        "valid mask 8",
				mask:        8,
				expectError: false,
			},
			{
				name:        "valid mask 16",
				mask:        16,
				expectError: false,
			},
			{
				name:        "invalid mask too small",
				mask:        7,
				expectError: true,
				errorMsg:    "mask must be between 8 and 24",
			},
			{
				name:        "invalid mask too large",
				mask:        25,
				expectError: true,
				errorMsg:    "mask must be between 8 and 24",
			},
			{
				name:        "invalid mask zero",
				mask:        0,
				expectError: true,
				errorMsg:    "mask must be between 8 and 24",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(tc.mask, "", "", 0, 0, 0, 0))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("RIRValidation", func(t *testing.T) {
		testCases := []struct {
			name        string
			rir         string
			expectError bool
			errorMsg    string
		}{
			{
				name:        "valid RIR ARIN",
				rir:         "ARIN",
				expectError: false,
			},
			{
				name:        "valid RIR RIPE",
				rir:         "RIPE",
				expectError: false,
			},
			{
				name:        "valid RIR APNIC",
				rir:         "APNIC",
				expectError: false,
			},
			{
				name:        "valid RIR LACNIC",
				rir:         "LACNIC",
				expectError: false,
			},
			{
				name:        "valid RIR AFRINIC",
				rir:         "AFRINIC",
				expectError: false,
			},
			{
				name:        "valid RIR case insensitive",
				rir:         "arin",
				expectError: false,
			},
			{
				name:        "invalid RIR",
				rir:         "INVALID",
				expectError: true,
				errorMsg:    "invalid RIR: INVALID",
			},
			{
				name:        "empty RIR (should be valid - means no filter)",
				rir:         "",
				expectError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(24, tc.rir, "", 0, 0, 0, 0))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("CountryValidation", func(t *testing.T) {
		testCases := []struct {
			name        string
			country     string
			expectError bool
			errorMsg    string
		}{
			{
				name:        "valid country US",
				country:     "US",
				expectError: false,
			},
			{
				name:        "valid country DE",
				country:     "DE",
				expectError: false,
			},
			{
				name:        "valid country lowercase",
				country:     "us",
				expectError: false,
			},
			{
				name:        "invalid country code too short",
				country:     "U",
				expectError: true,
				errorMsg:    "country code must be 2 letters",
			},
			{
				name:        "invalid country code too long",
				country:     "USA",
				expectError: true,
				errorMsg:    "country code must be 2 letters",
			},
			{
				name:        "invalid country code",
				country:     "ZZ",
				expectError: true,
				errorMsg:    "invalid ISO-3166-1 alpha-2 country code",
			},
			{
				name:        "empty country (should be valid - means no filter)",
				country:     "",
				expectError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(24, "", tc.country, 0, 0, 0, 0))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("NetworkAddressValidation", func(t *testing.T) {
		testCases := []struct {
			name           string
			networkAddress uint64
			expectError    bool
			errorMsg       string
		}{
			{
				name:           "valid network address",
				networkAddress: 0x08080800, // 8.8.8.0
				expectError:    false,
			},
			{
				name:           "valid network address with zero last octet",
				networkAddress: 0x0A000000, // 10.0.0.0
				expectError:    false,
			},
			{
				name:           "invalid network address non-zero last octet",
				networkAddress: 0x08080801, // 8.8.8.1
				expectError:    true,
				errorMsg:       "fourth octet must be zero for subnet",
			},
			{
				name:           "zero network address (should be valid - means no filter)",
				networkAddress: 0,
				expectError:    false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, tc.networkAddress, 0))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("FirstOctetValidation", func(t *testing.T) {
		testCases := []struct {
			name        string
			firstOctet  uint64
			expectError bool
			errorMsg    string
		}{
			{
				name:        "valid first octet 8",
				firstOctet:  8,
				expectError: false,
			},
			{
				name:        "valid first octet 255",
				firstOctet:  255,
				expectError: false,
			},
			{
				name:        "invalid first octet too large",
				firstOctet:  256,
				expectError: true,
				errorMsg:    "first octet must be between 0 and 255",
			},
			{
				name:        "zero first octet (should be valid - means no filter)",
				firstOctet:  0,
				expectError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, 0, tc.firstOctet))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("PriceRangeValidation", func(t *testing.T) {
		testCases := []struct {
			name        string
			minprice    uint64
			maxprice    uint64
			expectError bool
			errorMsg    string
		}{
			{
				name:        "valid price range",
				minprice:    50,
				maxprice:    200,
				expectError: false,
			},
			{
				name:        "valid price range equal",
				minprice:    100,
				maxprice:    100,
				expectError: false,
			},
			{
				name:        "invalid price range min > max",
				minprice:    200,
				maxprice:    100,
				expectError: true,
				errorMsg:    "minimum price cannot be greater than maximum price",
			},
			{
				name:        "only minprice (should be valid)",
				minprice:    50,
				maxprice:    0,
				expectError: false,
			},
			{
				name:        "only maxprice (should be valid)",
				minprice:    0,
				maxprice:    200,
				expectError: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", tc.minprice, tc.maxprice, 0, 0))
				if tc.expectError {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.errorMsg)
					require.Nil(t, resp)
				} else {
					require.NoError(t, err)
					require.NotNil(t, resp)
				}
			})
		}
	})

	t.Run("BasicFiltering", func(t *testing.T) {
		// Test that only FORSALE contracts are returned
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Should return contracts 1, 5, 6, 8 (all /24 and FORSALE)
		require.Equal(t, 4, len(resp.Contract))

		// Verify all returned contracts have status FORSALE
		for _, contract := range resp.Contract {
			require.Equal(t, types.ContractStatus_CONTRACT_STATUS_FORSALE, contract.Status)
			require.Equal(t, uint64(24), contract.Mask)
		}
	})

	t.Run("MaskFiltering", func(t *testing.T) {
		// Test /16 filtering
		resp, err := keeper.MarketplaceContract(ctx, request(16, "", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract)) // Only contract 2
		require.Equal(t, uint64(2), resp.Contract[0].Id)

		// Test /8 filtering
		resp, err = keeper.MarketplaceContract(ctx, request(8, "", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract)) // Only contract 4
		require.Equal(t, uint64(4), resp.Contract[0].Id)
	})

	t.Run("RIRFiltering", func(t *testing.T) {
		// Test ARIN filtering
		resp, err := keeper.MarketplaceContract(ctx, request(24, "ARIN", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract)) // Contracts 1, 8
		for _, contract := range resp.Contract {
			require.Equal(t, "ARIN", contract.Rir)
		}

		// Test RIPE filtering (no /24 RIPE contracts)
		resp, err = keeper.MarketplaceContract(ctx, request(24, "RIPE", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 0, len(resp.Contract))

		// Test case insensitive
		resp, err = keeper.MarketplaceContract(ctx, request(24, "arin", "", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract))
	})

	t.Run("CountryFiltering", func(t *testing.T) {
		// Test US filtering
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "US", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract)) // Contract 1
		require.Equal(t, uint64(1), resp.Contract[0].Id)

		// Test case insensitive
		resp, err = keeper.MarketplaceContract(ctx, request(24, "", "us", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract))
	})

	t.Run("NetworkAddressFiltering", func(t *testing.T) {
		// Test specific network address filtering
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, 0x08080800, 0)) // 8.8.8.0
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract)) // Contract 1
		require.Equal(t, uint64(1), resp.Contract[0].Id)
	})

	t.Run("FirstOctetFiltering", func(t *testing.T) {
		// Test first octet 8 filtering
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, 0, 8))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract)) // Contracts 1, 8
		for _, contract := range resp.Contract {
			firstOctet := (contract.NetworkAddress >> 24) & 0xFF
			require.Equal(t, uint64(8), firstOctet)
		}

		// Test first octet 172 filtering
		resp, err = keeper.MarketplaceContract(ctx, request(24, "", "", 0, 0, 0, 172))
		require.NoError(t, err)
		require.Equal(t, 1, len(resp.Contract)) // Contract 5
		require.Equal(t, uint64(5), resp.Contract[0].Id)
	})

	t.Run("PriceFiltering", func(t *testing.T) {
		// Test minimum price filtering
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 90, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract)) // Contracts 1 ($100), 8 ($90)

		// Test maximum price filtering
		resp, err = keeper.MarketplaceContract(ctx, request(24, "", "", 0, 90, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Contract)) // Contracts 5 ($50), 6 ($80), 8 ($90)

		// Test price range filtering
		resp, err = keeper.MarketplaceContract(ctx, request(24, "", "", 80, 100, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 3, len(resp.Contract)) // Contracts 1 ($100), 6 ($80), 8 ($90)
	})

	t.Run("CombinedFiltering", func(t *testing.T) {
		// Test multiple filters combined: /24 + ARIN + first octet 8
		resp, err := keeper.MarketplaceContract(ctx, request(24, "ARIN", "", 0, 0, 0, 8))
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract)) // Contracts 1, 8

		// Test multiple filters that should return no results
		resp, err = keeper.MarketplaceContract(ctx, request(24, "RIPE", "US", 0, 0, 0, 0))
		require.NoError(t, err)
		require.Equal(t, 0, len(resp.Contract)) // No RIPE contracts in US
	})

	// Get expected FORSALE contracts for pagination testing
	expectedForsaleIds := []uint64{}
	for _, contract := range contracts {
		if contract.Status == types.ContractStatus_CONTRACT_STATUS_FORSALE {
			expectedForsaleIds = append(expectedForsaleIds, contract.Id)
		}
	}

	t.Run("Pagination", func(t *testing.T) {
		// Test with limit
		resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
			Mask: 24, // This should return 4 contracts
			Pagination: &query.PageRequest{
				Limit: 2,
			},
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(resp.Contract))
		require.NotNil(t, resp.Pagination.NextKey)

		// Test second page
		resp2, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
			Mask: 24,
			Pagination: &query.PageRequest{
				Key:   resp.Pagination.NextKey,
				Limit: 2,
			},
		})
		require.NoError(t, err)
		require.Equal(t, 2, len(resp2.Contract))

		// Verify no duplicate contracts between pages
		page1Ids := make(map[uint64]bool)
		for _, contract := range resp.Contract {
			page1Ids[contract.Id] = true
		}
		for _, contract := range resp2.Contract {
			require.False(t, page1Ids[contract.Id], "Contract ID %d appeared in both pages", contract.Id)
		}
	})

	t.Run("EmptyResults", func(t *testing.T) {
		// Test filters that should return no results
		resp, err := keeper.MarketplaceContract(ctx, request(24, "", "", 1000, 2000, 0, 0)) // Very high price range
		require.NoError(t, err)
		require.Equal(t, 0, len(resp.Contract))
		require.Nil(t, resp.Pagination.NextKey)
		require.Equal(t, uint64(0), resp.Pagination.Total)
	})
}

func TestMarketplaceContractIntegration(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)

	t.Run("NoContractsScenario", func(t *testing.T) {
		// Test with empty database
		resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
			Mask: 24,
		})
		require.NoError(t, err)
		require.Equal(t, 0, len(resp.Contract))
		require.Nil(t, resp.Pagination.NextKey)
		require.Equal(t, uint64(0), resp.Pagination.Total)
	})

	t.Run("AllNonForsaleContractsScenario", func(t *testing.T) {
		// Create only non-FORSALE contracts
		nonForsaleContracts := []types.Contract{
			{
				Id:             1,
				SubnetId:       101,
				OwnerOrg:       201,
				LesseeOrg:      301,
				Price:          sdk.NewInt64Coin("ipex", 100),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 50),
				Mask:           24,
				NetworkAddress: 0x08080800,
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_LEASED,
			},
			{
				Id:             2,
				SubnetId:       102,
				OwnerOrg:       202,
				Price:          sdk.NewInt64Coin("ipex", 200),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 100),
				Mask:           24,
				NetworkAddress: 0x0A000000,
				Rir:            "RIPE",
				Country:        "DE",
				Status:         types.ContractStatus_CONTRACT_STATUS_CANCELED_OWNER,
			},
		}

		for _, contract := range nonForsaleContracts {
			keeper.SetContract(ctx, contract)
		}

		// Should return no contracts since none are FORSALE
		resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
			Mask: 24,
		})
		require.NoError(t, err)
		require.Equal(t, 0, len(resp.Contract))
	})

	t.Run("RealWorldScenario", func(t *testing.T) {
		// Reset context for fresh test
		keeper, ctx = keepertest.LeaseKeeper(t)

		// Create realistic marketplace scenario
		realWorldContracts := []types.Contract{
			// US-EAST: 8.8.8.0/24, $150/month, Google DNS range
			{
				Id:             1,
				SubnetId:       101,
				OwnerOrg:       201,
				Price:          sdk.NewInt64Coin("ipex", 150),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 75),
				Mask:           24,
				NetworkAddress: 0x08080800, // 8.8.8.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// EU-WEST: 185.199.108.0/24, €200/month equivalent, GitHub Pages range
			{
				Id:             2,
				SubnetId:       102,
				OwnerOrg:       202,
				Price:          sdk.NewInt64Coin("ipex", 200),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 100),
				Mask:           24,
				NetworkAddress: 0xB9C76C00, // 185.199.108.0
				Rir:            "RIPE",
				Country:        "NL",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// ASIA-PACIFIC: 103.21.244.0/24, $120/month, Cloudflare range
			{
				Id:             3,
				SubnetId:       103,
				OwnerOrg:       203,
				Price:          sdk.NewInt64Coin("ipex", 120),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 60),
				Mask:           24,
				NetworkAddress: 0x6715F400, // 103.21.244.0
				Rir:            "APNIC",
				Country:        "SG",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// BRAZIL: 200.160.0.0/16, $300/month, Large subnet
			{
				Id:             4,
				SubnetId:       104,
				OwnerOrg:       204,
				Price:          sdk.NewInt64Coin("ipex", 300),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 150),
				Mask:           16,
				NetworkAddress: 0xC8A00000, // 200.160.0.0
				Rir:            "LACNIC",
				Country:        "BR",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// Already leased contract (should not appear)
			{
				Id:             5,
				SubnetId:       105,
				OwnerOrg:       205,
				LesseeOrg:      305,
				Price:          sdk.NewInt64Coin("ipex", 180),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 90),
				Mask:           24,
				NetworkAddress: 0x08080900, // 8.8.9.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_LEASED,
			},
		}

		for _, contract := range realWorldContracts {
			keeper.SetContract(ctx, contract)
		}

		// Test real-world query scenarios
		t.Run("FindAffordableSubnets", func(t *testing.T) {
			// Find subnets under $200/month
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:     24,
				Maxprice: 200,
			})
			require.NoError(t, err)
			require.Equal(t, 3, len(resp.Contract)) // Contracts 1, 2, 3

			// Verify all are under $200
			for _, contract := range resp.Contract {
				require.LessOrEqual(t, contract.Price.Amount.Uint64(), uint64(200))
			}
		})

		t.Run("FindUSBasedNetworks", func(t *testing.T) {
			// Find US-based networks
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:    24,
				Country: "US",
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Only contract 1
			require.Equal(t, "US", resp.Contract[0].Country)
		})

		t.Run("FindLargeSubnets", func(t *testing.T) {
			// Find large subnets (/16 or smaller)
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask: 16,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Only contract 4
			require.Equal(t, uint64(16), resp.Contract[0].Mask)
		})

		t.Run("FindByRIR", func(t *testing.T) {
			// Find RIPE networks (European)
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask: 24,
				Rir:  "RIPE",
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Only contract 2
			require.Equal(t, "RIPE", resp.Contract[0].Rir)
		})

		t.Run("FindPremiumRanges", func(t *testing.T) {
			// Find first octet 8 (premium range - Google, Level3, etc.)
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:       24,
				FirstOctet: 8,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Only contract 1
			firstOctet := (resp.Contract[0].NetworkAddress >> 24) & 0xFF
			require.Equal(t, uint64(8), firstOctet)
		})
	})

	t.Run("EdgeCaseScenarios", func(t *testing.T) {
		keeper, ctx = keepertest.LeaseKeeper(t)

		// Test edge case contracts
		edgeCaseContracts := []types.Contract{
			// Zero first octet (rare but valid)
			{
				Id:             1,
				SubnetId:       101,
				OwnerOrg:       201,
				Price:          sdk.NewInt64Coin("ipex", 50),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 25),
				Mask:           8,
				NetworkAddress: 0x00000000, // 0.0.0.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// Maximum mask /24
			{
				Id:             2,
				SubnetId:       102,
				OwnerOrg:       202,
				Price:          sdk.NewInt64Coin("ipex", 100),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 50),
				Mask:           24,
				NetworkAddress: 0xFF000000, // 255.0.0.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
			// Minimum mask /8
			{
				Id:             3,
				SubnetId:       103,
				OwnerOrg:       203,
				Price:          sdk.NewInt64Coin("ipex", 1000),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", 500),
				Mask:           8,
				NetworkAddress: 0x0A000000, // 10.0.0.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			},
		}

		for _, contract := range edgeCaseContracts {
			keeper.SetContract(ctx, contract)
		}

		t.Run("FindZeroFirstOctet", func(t *testing.T) {
			// This should not find contract 1 because firstOctet=0 means "no filter"
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:       8,
				FirstOctet: 0,
			})
			require.NoError(t, err)
			// Should return all /8 contracts (both contract 1 and 3)
			require.Equal(t, 2, len(resp.Contract))
		})

		t.Run("FindHighFirstOctet", func(t *testing.T) {
			// Find first octet 255
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:       24,
				FirstOctet: 255,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Contract 2
			firstOctet := (resp.Contract[0].NetworkAddress >> 24) & 0xFF
			require.Equal(t, uint64(255), firstOctet)
		})

		t.Run("ExtremeValueFiltering", func(t *testing.T) {
			// Test with maximum valid values
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask:           24,
				Rir:            "ARIN",
				Country:        "US",
				Minprice:       1,
				Maxprice:       10000,
				NetworkAddress: 0xFF000000, // 255.0.0.0
				FirstOctet:     255,
			})
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Contract)) // Should find contract 2
		})
	})

	t.Run("StressTestPagination", func(t *testing.T) {
		keeper, ctx = keepertest.LeaseKeeper(t)

		// Create many contracts for pagination stress test
		const numContracts = 50
		for i := 1; i <= numContracts; i++ {
			contract := types.Contract{
				Id:             uint64(i),
				SubnetId:       uint64(100 + i),
				OwnerOrg:       uint64(200 + i),
				Price:          sdk.NewInt64Coin("ipex", int64(100+i)),
				SafetyDeposit:  sdk.NewInt64Coin("ipex", int64(50+i)),
				Mask:           24,
				NetworkAddress: uint64(0x08000000 + (i << 8)), // 8.0.i.0
				Rir:            "ARIN",
				Country:        "US",
				Status:         types.ContractStatus_CONTRACT_STATUS_FORSALE,
			}
			keeper.SetContract(ctx, contract)
		}

		// Test paginating through all contracts
		var allContracts []types.Contract
		var nextKey []byte
		pageSize := uint64(7) // Odd number to test edge cases

		for {
			resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
				Mask: 24,
				Pagination: &query.PageRequest{
					Key:   nextKey,
					Limit: pageSize,
				},
			})
			require.NoError(t, err)

			allContracts = append(allContracts, resp.Contract...)

			if resp.Pagination.NextKey == nil {
				break
			}
			nextKey = resp.Pagination.NextKey
		}

		// Should have retrieved all contracts
		require.Equal(t, numContracts, len(allContracts))

		// Verify no duplicates
		seenIds := make(map[uint64]bool)
		for _, contract := range allContracts {
			require.False(t, seenIds[contract.Id], "Duplicate contract ID %d found", contract.Id)
			seenIds[contract.Id] = true
		}
	})
}

func TestMarketplaceContractValidationEdgeCases(t *testing.T) {
	keeper, ctx := keepertest.LeaseKeeper(t)

	t.Run("AllValidRIRsAccepted", func(t *testing.T) {
		validRirs := []string{"ARIN", "RIPE", "APNIC", "LACNIC", "AFRINIC"}

		for _, rir := range validRirs {
			t.Run("RIR_"+rir, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
					Mask: 24,
					Rir:  rir,
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
			})

			// Test case insensitive
			t.Run("RIR_"+rir+"_lowercase", func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
					Mask: 24,
					Rir:  strings.ToLower(rir),
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
			})
		}
	})

	t.Run("ValidCountryCodesFromDifferentRegions", func(t *testing.T) {
		validCountries := []string{
			"US", "CA", "MX", // North America
			"BR", "AR", "CL", // South America
			"GB", "DE", "FR", "IT", "ES", "NL", // Europe
			"JP", "CN", "KR", "IN", "SG", "AU", // Asia-Pacific
			"ZA", "NG", "EG", "KE", // Africa
		}

		for _, country := range validCountries {
			t.Run("Country_"+country, func(t *testing.T) {
				resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
					Mask:    24,
					Country: country,
				})
				require.NoError(t, err)
				require.NotNil(t, resp)
			})
		}
	})

	t.Run("BoundaryValueTesting", func(t *testing.T) {
		testCases := []struct {
			name  string
			mask  uint64
			valid bool
		}{
			{"mask_7", 7, false},
			{"mask_8", 8, true},
			{"mask_16", 16, true},
			{"mask_24", 24, true},
			{"mask_25", 25, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
					Mask: tc.mask,
				})
				if tc.valid {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})

	t.Run("LargeNumberHandling", func(t *testing.T) {
		// Test with large but valid values
		resp, err := keeper.MarketplaceContract(ctx, &types.QueryMarketplaceContractRequest{
			Mask:           24,
			Minprice:       18446744073709551615, // Max uint64
			Maxprice:       18446744073709551615, // Max uint64
			NetworkAddress: 0xFFFFFF00,           // 255.255.255.0
			FirstOctet:     255,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		// Should return no results but not error
		require.Equal(t, 0, len(resp.Contract))
	})
}
