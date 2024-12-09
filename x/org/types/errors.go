package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/org module sentinel errors
var (
	ErrInvalidSigner             = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidCountryCode        = sdkerrors.Register(ModuleName, 1101, "country code must be exactly 2 capital letters")
	ErrInvalidWebsiteUrl         = sdkerrors.Register(ModuleName, 1102, "website URL is invalid")
	ErrInvalidEmail              = sdkerrors.Register(ModuleName, 1103, "email format is invalid")
	ErrInvalidName               = sdkerrors.Register(ModuleName, 1104, "invalid name")
	ErrInvalidLegalName          = sdkerrors.Register(ModuleName, 1105, "invalid legal name")
	ErrInvalidAddress            = sdkerrors.Register(ModuleName, 1106, "invalid address")
	ErrInvalidRegistrationNumber = sdkerrors.Register(ModuleName, 1107, "invalid registration number")
	ErrNothingToUpdate           = sdkerrors.Register(ModuleName, 1008, "no fields provided for update")
	ErrRequiredFields            = sdkerrors.Register(ModuleName, 1009, "all required fields must be provided")
	ErrOrgNotFound               = sdkerrors.Register(ModuleName, 1010, "organization not found")
)
