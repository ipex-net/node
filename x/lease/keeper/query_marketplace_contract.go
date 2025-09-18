package keeper

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"ipex/x/lease/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/biter777/countries"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validateRir validates if RIR is one of the supported values
func validateRir(rir string) error {
	supportedRirs := []string{"ARIN", "RIPE", "APNIC", "LACNIC", "AFRINIC"}
	for _, supportedRir := range supportedRirs {
		if strings.EqualFold(rir, supportedRir) {
			return nil
		}
	}
	return fmt.Errorf("invalid RIR: %s, supported values: %s", rir, strings.Join(supportedRirs, ", "))
}

// validateCountryCode validates if country code is a valid ISO-3166-1 alpha-2 code
func validateCountryCode(alpha2 string) error {
	if len(alpha2) != 2 {
		return fmt.Errorf("country code must be 2 letters")
	}
	country := countries.ByName(strings.ToUpper(alpha2))
	if !country.IsValid() {
		return fmt.Errorf("invalid ISO-3166-1 alpha-2 country code: %s", alpha2)
	}
	return nil
}

// UintToNetIP converts uint64 representation of IPv4 subnet address to net.IP type
func UintToNetIP(ipUint uint64) net.IP {
	first := byte(ipUint >> 24)         // Get first byte
	second := byte(ipUint >> 16 & 0xFF) // Get second byte
	third := byte(ipUint >> 8 & 0xFF)   // Get third byte
	fourth := byte(ipUint & 0xFF)       // Get fourth byte
	return net.IPv4(first, second, third, fourth)
}

// validateNetworkAddress validates the network address format
func validateNetworkAddress(networkAddress uint64) error {
	parsedIp := UintToNetIP(networkAddress)
	if parsedIp == nil {
		return fmt.Errorf("invalid IPv4 address format")
	}

	// Validate that the last octet is equal to zero, as the smallest subnet size is /24
	if networkAddress&0xFF != 0 {
		return fmt.Errorf("fourth octet must be zero for subnet (%s)", parsedIp.String())
	}

	return nil
}

// validateOctet validates if octet value is within valid range
func validateOctet(octet uint64, octetName string) error {
	if octet > 255 {
		return fmt.Errorf("%s must be between 0 and 255", octetName)
	}
	return nil
}

func (k Keeper) MarketplaceContract(goCtx context.Context, req *types.QueryMarketplaceContractRequest) (*types.QueryMarketplaceContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate required mask field
	if req.Mask != 0 && (req.Mask < 8 || req.Mask > 24) {
		return nil, errorsmod.Wrap(types.ErrInvalidPrefixLength, "mask must be between 8 and 24")
	}

	// Validate optional fields if provided (strict validation - no trimming)
	if req.Rir != "" && req.Rir != "none" && req.Rir != "any" {
		if err := validateRir(req.Rir); err != nil {
			return nil, errorsmod.Wrap(types.ErrUnknownRIR, err.Error())
		}
	} else {
		req.Rir = "" // Normalize placeholder values to empty string
	}

	if req.Country != "" && req.Country != "none" && req.Country != "any" {
		if err := validateCountryCode(req.Country); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidCountryCode, err.Error())
		}
	} else {
		req.Country = "" // Normalize placeholder values to empty string
	}

	if req.NetworkAddress != 0 {
		if err := validateNetworkAddress(req.NetworkAddress); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidIPFormat, err.Error())
		}
	}

	// Validate first octet field if provided
	if req.FirstOctet != 0 {
		if err := validateOctet(req.FirstOctet, "first octet"); err != nil {
			return nil, errorsmod.Wrap(types.ErrInvalidIPFormat, err.Error())
		}
	}

	// Validate price range
	if req.Minprice > 0 && req.Maxprice > 0 && req.Minprice > req.Maxprice {
		return nil, errorsmod.Wrap(types.ErrInvalidPrice, "minimum price cannot be greater than maximum price")
	}

	// Build the query to filter contracts following ListContract pattern
	var contracts []types.Contract
	var totalCount uint64
	var lastKey uint64

	// Set up pagination parameters
	limit := uint64(20) // Default limit
	offset := uint64(0)
	var startKey *uint64

	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		// Support key-based pagination
		if len(req.Pagination.Key) > 0 {
			// Parse the key as uint64 (contract ID)
			if len(req.Pagination.Key) == 8 {
				keyValue := binary.BigEndian.Uint64(req.Pagination.Key)
				startKey = &keyValue
				offset = 0 // When using key-based pagination, ignore offset
			}
		}
	}

	var skipped uint64
	var collectedCount uint64
	var collectingItems = true

	// Use collections to iterate through contracts with pagination
	err := k.Contracts.Walk(ctx, nil, func(key uint64, contract types.Contract) (bool, error) {
		// Only include contracts that are available for sale (status = FORSALE)
		if contract.Status != types.ContractStatus_CONTRACT_STATUS_FORSALE {
			return false, nil // Continue iteration, don't include this contract
		}

		// Apply mask filter if mask is not zero
		if req.Mask != 0 && contract.Mask != req.Mask {
			return false, nil
		}

		// Apply RIR filter
		if req.Rir != "" && !strings.EqualFold(contract.Rir, req.Rir) {
			return false, nil
		}

		// Apply country filter
		if req.Country != "" && !strings.EqualFold(contract.Country, req.Country) {
			return false, nil
		}

		// Apply network address filter
		if req.NetworkAddress != 0 && contract.NetworkAddress != req.NetworkAddress {
			return false, nil
		}

		// Apply first octet filter
		if req.FirstOctet != 0 {
			contractFirstOctet := (contract.NetworkAddress >> 24) & 0xFF
			if contractFirstOctet != req.FirstOctet {
				return false, nil
			}
		}

		// Apply price filters
		contractPrice := contract.Price.Amount.Uint64()
		if req.Minprice > 0 && contractPrice < req.Minprice {
			return false, nil
		}

		if req.Maxprice > 0 && contractPrice > req.Maxprice {
			return false, nil
		}

		// Contract matches all filters, now apply pagination logic
		totalCount++

		if collectingItems {
			// If using key-based pagination, skip until we reach the start key
			if startKey != nil && key <= *startKey {
				return false, nil
			}

			// Skip entries until we reach the offset (for offset-based pagination)
			if startKey == nil && skipped < offset {
				skipped++
				return false, nil
			}

			// Stop collecting items if we've collected enough
			if collectedCount >= limit {
				collectingItems = false
				return false, nil // Continue counting but stop collecting
			}

			contracts = append(contracts, contract)
			lastKey = key
			collectedCount++
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if collectedCount == limit {
		// Convert the last contract ID to bytes for next key
		nextKey = make([]byte, 8)
		binary.BigEndian.PutUint64(nextKey, lastKey)
	}

	return &types.QueryMarketplaceContractResponse{
		Contract: contracts,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   totalCount,
		},
	}, nil
}
