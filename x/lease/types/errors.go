package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/lease module sentinel errors
var (
	ErrInvalidSigner       = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidCreator      = sdkerrors.Register(ModuleName, 1101, "need to be a contract creator to perform this operation")
	ErrInvalidIPFormat     = sdkerrors.Register(ModuleName, 1102, "invalid IPv4 address format")
	ErrReservedIPRange     = sdkerrors.Register(ModuleName, 1103, "IP address is in a reserved range")
	ErrInvalidPrefixLength = sdkerrors.Register(ModuleName, 1104, "prefix length must be between 8 and 24")
	ErrEmptyRIR            = sdkerrors.Register(ModuleName, 1105, "RIR cannot be empty")
	ErrUnknownRIR          = sdkerrors.Register(ModuleName, 1106, "Unknown RIR name")
	ErrInvalidCountryCode  = sdkerrors.Register(ModuleName, 1107, "country code must be exactly 2 capital letters")
	ErrInvalidPrice        = sdkerrors.Register(ModuleName, 1108, "price must be positive")
	ErrInvalidNoticePeriod = sdkerrors.Register(ModuleName, 1109, "notice period must be positive")
	ErrInvalidGeofeedURL   = sdkerrors.Register(ModuleName, 1110, "invalid geofeed URL format or requirements")
	ErrInvalidContractStatus   = sdkerrors.Register(ModuleName, 1111, "invalid contract status")
)
