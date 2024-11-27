package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/lease module sentinel errors
var (
	ErrInvalidSigner  = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample         = sdkerrors.Register(ModuleName, 1101, "sample error")
	ErrInvalidCreator = sdkerrors.Register(ModuleName, 1102, "need to be a contract creator to perform this operation")
)
