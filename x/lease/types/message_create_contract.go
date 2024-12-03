package types

import (
	"net"
	"net/url"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateContract{}

func isReservedIPRange(ip net.IP) bool {
	// Convert IP to 32-bit integer for easy range checking
	ipInt := uint32(ip[12])<<24 | uint32(ip[13])<<16 | uint32(ip[14])<<8 | uint32(ip[15])

	// Reserved ranges to check
	reservedRanges := []struct {
		start uint32
		end   uint32
	}{
		{0x00000000, 0x00FFFFFF},    // 0.0.0.0/8        - Current network
		{0x0A000000, 0x0AFFFFFF},    // 10.0.0.0/8       - Private network
		{0x64400000, 0x647FFFFF},    // 100.64.0.0/10    - CGN Shared space
		{0x7F000000, 0x7FFFFFFF},    // 127.0.0.0/8      - Loopback
		{0xA9FE0000, 0xA9FEFFFF},    // 169.254.0.0/16   - Link-local
		{0xAC100000, 0xAC1FFFFF},    // 172.16.0.0/12    - Private network
		{0xC0000000, 0xC00000FF},    // 192.0.0.0/24     - IETF Protocol
		{0xC0000200, 0xC00002FF},    // 192.0.2.0/24     - TEST-NET-1
		{0xC0586300, 0xC05863FF},    // 192.88.99.0/24   - IPv6 to IPv4 Relay
		{0xC0A80000, 0xC0A8FFFF},    // 192.168.0.0/16   - Private network
		{0xC6120000, 0xC613FFFF},    // 198.18.0.0/15    - Benchmark testing
		{0xC6336400, 0xC63364FF},    // 198.51.100.0/24  - TEST-NET-2
		{0xCB007100, 0xCB0071FF},    // 203.0.113.0/24   - TEST-NET-3
		{0xE0000000, 0xEFFFFFFF},    // 224.0.0.0/4      - Multicast
		{0xE9FC0000, 0xE9FC00FF},    // 233.252.0.0/24 	 - MCAST-TEST-NET
		{0xF0000000, 0xFFFFFFFF},    // 240.0.0.0/4 	 - Reserved for future use
	}

	for _, r := range reservedRanges {
		if ipInt >= r.start && ipInt <= r.end {
			return true
		}
	}
	return false
}

func NewMsgCreateContract(creator string, networkAddress string, prefixLength uint64, rir string, countryCode string, geofeedUrl string, price sdk.Coin, deposit sdk.Coin, noticePeriodBlocks uint64) *MsgCreateContract {
	return &MsgCreateContract{
		Creator:            creator,
		NetworkAddress:     networkAddress,
		PrefixLength:       prefixLength,
		Rir:                rir,
		CountryCode:        countryCode,
		GeofeedUrl:         geofeedUrl,
		Price:              price,
		Deposit:            deposit,
		NoticePeriodBlocks: noticePeriodBlocks,
	}
}

func (msg *MsgCreateContract) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate IP address format using regex
	ipv4Regex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}0$`)
	if !ipv4Regex.MatchString(msg.NetworkAddress) {
		return errorsmod.Wrapf(ErrInvalidIPFormat, "invalid IPv4 address format (%s)", msg.NetworkAddress)
	}

	// Parse the IP address for reserved range checking
	ip := net.ParseIP(msg.NetworkAddress)
	if ip == nil {
		return errorsmod.Wrapf(ErrInvalidIPFormat, "invalid IPv4 address format (%s)", msg.NetworkAddress)
	}

	// Check if IP is in a reserved range
	if isReservedIPRange(ip) {
		return errorsmod.Wrapf(ErrReservedIPRange, "IP address is in a reserved range (%s)", msg.NetworkAddress)
	}
		
	// Validate prefix length for IPv4 (8-24)
	if msg.PrefixLength < 8 || msg.PrefixLength > 24 {
		return errorsmod.Wrap(ErrInvalidPrefixLength, "prefix length must be between 8 and 24")
	}

	if msg.Rir == "" {
		return errorsmod.Wrap(ErrEmptyRIR, "RIR cannot be empty")
	}

	if msg.Rir != "RIPE" && msg.Rir != "ARIN" && msg.Rir != "APNIC" && msg.Rir != "AFRINIC" {
		return errorsmod.Wrapf(ErrUnknownRIR, "Unknown RIR name (%s)", msg.Rir)
	}
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !countryCodeRegex.MatchString(msg.CountryCode) {
		return errorsmod.Wrap(ErrInvalidCountryCode, "country code must be exactly 2 capital letters")
	}

	if !msg.Price.IsPositive() {
		return errorsmod.Wrap(ErrInvalidPrice, "price must be positive")
	}

	if msg.NoticePeriodBlocks == 0 {
		return errorsmod.Wrap(ErrInvalidNoticePeriod, "notice period must be positive")
	}

	//Validate Geofeed
	if msg.GeofeedUrl == "" {
		return errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL not provided")
	} else { 
		parsedUrl, err := url.Parse(msg.GeofeedUrl)
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidGeofeedURL, "invalid geofeed URL format (%s)", msg.GeofeedUrl)
		}

		// Check if scheme is HTTPS
		if parsedUrl.Scheme != "https" {
			return errorsmod.Wrap(ErrInvalidGeofeedURL, "geofeed URL must use HTTPS")
		}

		// Check if host is present
		if parsedUrl.Host == "" {
			return errorsmod.Wrapf(ErrInvalidGeofeedURL, "geofeed URL must have a valid host (%s)", msg.GeofeedUrl)
		}
	}

	return nil
}
