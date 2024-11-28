package types

import (
	"net/url"
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateContract{}

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
	ipv4Regex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	if !ipv4Regex.MatchString(msg.NetworkAddress) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid IPv4 address format")
	}

	// Validate prefix length for IPv4 (8-24)
	if msg.PrefixLength < 8 || msg.PrefixLength > 24 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "prefix length must be between 8 and 24")
	}

	if msg.Rir == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "RIR cannot be empty")
	}

	if msg.Rir != "RIPE" && msg.Rir != "ARIN" && msg.Rir != "APNIC" && msg.Rir != "AFRINIC" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Unknown RIR name")
	}
	countryCodeRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !countryCodeRegex.MatchString(msg.CountryCode) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "country code must be exactly 2 capital letters")
	}

	if !msg.Price.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}

	if msg.NoticePeriodBlocks == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "notice period must be positive")
	}

	//Validate Geofeed
	if msg.GeofeedUrl == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "geofeed URL not provided")
	} else { 
		parsedUrl, err := url.Parse(msg.GeofeedUrl)
		if err != nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid geofeed URL format")
		}

		// Check if scheme is HTTPS
		if parsedUrl.Scheme != "https" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "geofeed URL must use HTTPS")
		}

		// Check if host is present
		if parsedUrl.Host == "" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "geofeed URL must have a valid host")
		}
	}
	return nil
}
