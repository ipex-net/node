package keeper

import (
	"context"
	"net"
	"net/url"
	"strings"

	"ipex/x/subnet/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// UintToNetIP converts uint64 representation of IPv4 subnet address to net.IP type
func UintToNetIP(ipUint uint64) net.IP {
	// Extract octets - adjusting the shift order
	first := byte(ipUint >> 24)         // Get first byte (A9)
	second := byte(ipUint >> 16 & 0xFF) // Get second byte (FE)
	third := byte(ipUint >> 8 & 0xFF)   // Get third byte (00)
	fourth := byte(ipUint & 0xFF)       // Get fourth byte (00)

	return net.IPv4(first, second, third, fourth)
}

func isReservedIPRange(ip net.IP) bool {
	// Convert IP to 32-bit integer for easy range checking
	ipInt := uint32(ip[12])<<24 | uint32(ip[13])<<16 | uint32(ip[14])<<8 | uint32(ip[15])

	// Reserved ranges to check
	reservedRanges := []struct {
		start uint32
		end   uint32
	}{
		{0x00000000, 0x00FFFFFF}, // 0.0.0.0/8        - Current network
		{0x0A000000, 0x0AFFFFFF}, // 10.0.0.0/8       - Private network
		{0x64400000, 0x647FFFFF}, // 100.64.0.0/10    - CGN Shared space
		{0x7F000000, 0x7FFFFFFF}, // 127.0.0.0/8      - Loopback
		{0xA9FE0000, 0xA9FEFFFF}, // 169.254.0.0/16   - Link-local
		{0xAC100000, 0xAC1FFFFF}, // 172.16.0.0/12    - Private network
		{0xC0000000, 0xC00000FF}, // 192.0.0.0/24     - IETF Protocol
		{0xC0000200, 0xC00002FF}, // 192.0.2.0/24     - TEST-NET-1
		{0xC0586300, 0xC05863FF}, // 192.88.99.0/24   - IPv6 to IPv4 Relay
		{0xC0A80000, 0xC0A8FFFF}, // 192.168.0.0/16   - Private network
		{0xC6120000, 0xC613FFFF}, // 198.18.0.0/15    - Benchmark testing
		{0xC6336400, 0xC63364FF}, // 198.51.100.0/24  - TEST-NET-2
		{0xCB007100, 0xCB0071FF}, // 203.0.113.0/24   - TEST-NET-3
		{0xE0000000, 0xEFFFFFFF}, // 224.0.0.0/4      - Multicast
		{0xE9FC0000, 0xE9FC00FF}, // 233.252.0.0/24 	 - MCAST-TEST-NET
		{0xF0000000, 0xFFFFFFFF}, // 240.0.0.0/4 	 - Reserved for future use
	}

	for _, r := range reservedRanges {
		if ipInt >= r.start && ipInt <= r.end {
			return true
		}
	}
	return false
}

func (k msgServer) CreateSubnet(goCtx context.Context, msg *types.MsgCreateSubnet) (*types.MsgCreateSubnetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//Checking if the address is correct
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate IP address format using regex
	parsedIp := UintToNetIP(msg.NetworkAddress)
	if parsedIp == nil {
		return nil, errorsmod.Wrapf(ErrInvalidIPFormat, "invalid IPv4 address format (%s)", parsedIp.String())
	}

	//Validate that the last octet is equal to zero, as the smallest subnet size is /24
	if parsedIp[3] != 0 {
		return nil, errorsmod.Wrapf(ErrInvalidIPFormat, "subnet fourth octet must be zero for address (%s)", parsedIp.String())
	}

	// Check if IP is in a reserved range
	if isReservedIPRange(parsedIp) {
		return nil, errorsmod.Wrapf(ErrReservedIPRange, "IP address is in a reserved range (%s)", parsedIp.String())
	}

	// Validate prefix length for IPv4 (8-24)
	if msg.Mask < 8 || msg.Mask > 24 {
		return nil, errorsmod.Wrap(ErrInvalidMask, "prefix length must be between 8 and 24")
	}

	if msg.Rir != "ARIN" && msg.Rir !="RIPE" {
		return nil, errorsmod.Wrap(ErrUnknownRIR, "Unsopprted RIR")
	}

	//Validate Geofeed
	if msg.GeofeedUrl == "" {
		return nil, errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL not provided")
	}

	// Check if URL starts with http:// or https://
	if strings.HasPrefix(strings.ToLower(msg.GeofeedUrl), "http://") ||
		strings.HasPrefix(strings.ToLower(msg.GeofeedUrl), "https://") {
		return nil, errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL should not include http:// or https://")
	}

	//Check if URL parses
	parsedUrl, err := url.Parse("https://" + msg.GeofeedUrl)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrInvalidGeofeedURL, "invalid geofeed URL format (%s)", msg.GeofeedUrl)
	}

	// Check if host is present
	if parsedUrl.Host == "" {
		return nil, errorsmod.Wrapf(ErrInvalidGeofeedURL, "geofeed URL must have a valid host (%s)", msg.GeofeedUrl)
	}

	// Check if the org id exists
	if msg.OwnerOrgId != 0 {
		if !k.orgKeeper.HasOrg(ctx, msg.OwnerOrgId) {
			return nil, errorsmod.Wrapf(ErrInvalidOrgId, "org with ID %d not found", msg.OwnerOrgId)
		}
	} else {
		return nil, errorsmod.Wrap(ErrInvalidOrgId, "org ID not provided")
	}

	// Verify creator is org admin
	if !k.orgKeeper.VerifyOrgAdmin(goCtx, msg.OwnerOrgId, msg.Creator) {
		return nil, errorsmod.Wrapf(ErrNotOrgAdmin, "address %s is not admin of org %d", msg.Creator, msg.OwnerOrgId)
	}

	// Check for existing network with the same network address
	subnets := k.GetAllSubnet(ctx)
	for _, existingSubnet := range subnets {
		if existingSubnet.NetworkAddress == msg.NetworkAddress {
			return nil, ErrSubnetExists
		}
	}

	//Creating new subnet
	var subnet = types.Subnet{
		NetworkAddress: msg.NetworkAddress,
		Mask:           msg.Mask,
		Status:         types.SubnetStatus_SUBNET_STATUS_INITIAL,
		Rir:            msg.Rir,
		GeofeedUrl:     msg.GeofeedUrl,
		OwnerOrgId:     msg.OwnerOrgId,
		CreatedAt:      ctx.BlockTime(),
		UpdatedAt:      ctx.BlockTime(),
	}
	id := k.AppendSubnet(
		ctx,
		subnet,
	)

	return &types.MsgCreateSubnetResponse{Id: id}, nil
}
