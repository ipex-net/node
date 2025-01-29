package keeper

// DONTCOVER
import (
	sdkerrors "cosmossdk.io/errors"
)

const ModuleName = "lease"

// x/subnet module sentinel errors
var (
	ErrInvalidOrgId       = sdkerrors.Register(ModuleName, 1208, "invalid org id")
	ErrNotOrgAdmin        = sdkerrors.Register(ModuleName, 1209, "not organization admin")
	ErrSubnetNotFound     = sdkerrors.Register(ModuleName, 1210, "invalid subnet id")
)
