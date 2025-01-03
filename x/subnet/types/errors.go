package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/subnet module sentinel errors
var (
	ErrInvalidSigner = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSubnetExists        = sdkerrors.Register(ModuleName, 1101, "subnet already exists")
	ErrInvalidSubnet        = sdkerrors.Register(ModuleName, 1102, "invalid subnet")
)
