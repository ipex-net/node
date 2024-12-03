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
		{0xC0A80000, 0xC0A8FFFF},    // 192.168.0.0/16   - Private network
		{0xC6120000, 0xC613FFFF},    // 198.18.0.0/15    - Benchmark testing
		{0xC6336400, 0xC63364FF},    // 198.51.100.0/24  - TEST-NET-2
		{0xCB007100, 0xCB0071FF},    // 203.0.113.0/24   - TEST-NET-3
		{0xE0000000, 0xFFFFFFFF},    // 224.0.0.0/4      - Multicast
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
		return ErrInvalidIPFormat
	}

	// Parse the IP address for reserved range checking
	ip := net.ParseIP(msg.NetworkAddress)
	if ip == nil {
		return ErrInvalidIPFormat
	}

	// Check if IP is in a reserved range
	if isReservedIPRange(ip) {
		return ErrReservedIPRange
	}
		
	// Validate prefix length for IPv4 (8-24)
	if msg.PrefixLength < 8 || msg.PrefixLength > 24 {
		return ErrInvalidPrefixLength
	}

	if msg.Rir == "" {
		return ErrEmptyRIR
	}

	if msg.Rir != "RIPE" && msg.Rir != "ARIN" && msg.Rir != "APNIC" && msg.Rir != "AFRINIC" {
		return ErrUnknownRIR
	}
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !countryCodeRegex.MatchString(msg.CountryCode) {
		return ErrInvalidCountryCode
	}

	if !msg.Price.IsPositive() {
		return ErrInvalidPrice
	}

	if msg.NoticePeriodBlocks == 0 {
		return ErrInvalidNoticePeriod
	}

	//Validate Geofeed
	if msg.GeofeedUrl == "" {
		return ErrInvalidGeofeedURL
	} else { 
		parsedUrl, err := url.Parse(msg.GeofeedUrl)
		if err != nil {
			return ErrInvalidGeofeedURL
		}

		// Check if scheme is HTTPS
		if parsedUrl.Scheme != "https" {
			return ErrInvalidGeofeedURL
		}

		// Check if host is present
		if parsedUrl.Host == "" {
			return ErrInvalidGeofeedURL
		}
	}

	return nil
}
