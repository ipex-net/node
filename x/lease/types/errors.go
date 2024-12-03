package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/lease module sentinel errors
var (
	ErrInvalidSigner  = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidCreator = sdkerrors.Register(ModuleName, 1102, "need to be a contract creator to perform this operation")
	ErrInvalidIPFormat      = sdkerrors.Register(ModuleName, 1103, "invalid IPv4 address format")
    ErrReservedIPRange      = sdkerrors.Register(ModuleName, 1104, "IP address is in a reserved range")
    ErrInvalidPrefixLength  = sdkerrors.Register(ModuleName, 1105, "prefix length must be between 8 and 24")
    ErrEmptyRIR            = sdkerrors.Register(ModuleName, 1106, "RIR cannot be empty")
    ErrUnknownRIR          = sdkerrors.Register(ModuleName, 1107, "Unknown RIR name")
    ErrInvalidCountryCode   = sdkerrors.Register(ModuleName, 1108, "country code must be exactly 2 capital letters")
    ErrInvalidPrice        = sdkerrors.Register(ModuleName, 1109, "price must be positive")
    ErrInvalidNoticePeriod = sdkerrors.Register(ModuleName, 1110, "notice period must be positive")
    ErrInvalidGeofeedURL   = sdkerrors.Register(ModuleName, 1111, "invalid geofeed URL format or requirements")
)
