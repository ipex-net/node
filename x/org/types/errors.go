package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/org module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidCountryCode    = sdkerrors.Register(ModuleName, 1101, "country code must be exactly 2 capital letters")
	ErrInvalidWebsiteUrl	= sdkerrors.Register(ModuleName, 1102, "website URL is invalid")
	ErrInvalidEmail	= sdkerrors.Register(ModuleName, 1103, "email format is invalid")
)
