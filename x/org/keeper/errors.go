package keeper

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)
const ModuleName = "org"
// x/org module sentinel errors
var (
	ErrInvalidSigner             = sdkerrors.Register(ModuleName, 1200, "expected gov account as only signer for proposal message")
	ErrInvalidCountryCode        = sdkerrors.Register(ModuleName, 1201, "country code must be exactly 2 capital letters")
	ErrInvalidWebsiteUrl         = sdkerrors.Register(ModuleName, 1202, "website URL is invalid")
	ErrInvalidEmail              = sdkerrors.Register(ModuleName, 1203, "email format is invalid")
	ErrInvalidName               = sdkerrors.Register(ModuleName, 1204, "invalid name")
	ErrInvalidLegalName          = sdkerrors.Register(ModuleName, 1205, "invalid legal name")
	ErrInvalidAddress            = sdkerrors.Register(ModuleName, 1206, "invalid address")
	ErrInvalidRegistrationNumber = sdkerrors.Register(ModuleName, 1207, "invalid registration number")
	ErrNothingToUpdate           = sdkerrors.Register(ModuleName, 1208, "no fields provided for update")
	ErrRequiredFields            = sdkerrors.Register(ModuleName, 1209, "all required fields must be provided")
	ErrOrgNotFound               = sdkerrors.Register(ModuleName, 1210, "organization not found")
)
