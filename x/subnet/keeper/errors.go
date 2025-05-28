package keeper

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

const ModuleName = "subnet"

// x/subnet module sentinel errors
var (
	ErrSubnetExists           = sdkerrors.Register(ModuleName, 1201, "subnet already exists")
	ErrInvalidSubnet          = sdkerrors.Register(ModuleName, 1202, "invalid subnet")
	ErrInvalidIPFormat        = sdkerrors.Register(ModuleName, 1203, "invalid ip format")
	ErrReservedIPRange        = sdkerrors.Register(ModuleName, 1204, "reserved ip range")
	ErrInvalidMask            = sdkerrors.Register(ModuleName, 1205, "invalid mask value")
	ErrUnknownRIR             = sdkerrors.Register(ModuleName, 1206, "unknown rir")
	ErrInvalidGeofeedURL      = sdkerrors.Register(ModuleName, 1207, "invalid geofeed url")
	ErrInvalidOrgId           = sdkerrors.Register(ModuleName, 1208, "invalid org id")
	ErrNotOrgAdmin            = sdkerrors.Register(ModuleName, 1209, "not organization admin")
	ErrSubnetNotFound         = sdkerrors.Register(ModuleName, 1210, "invalid subnet id")
	ErrReservedASNumber       = sdkerrors.Register(ModuleName, 1211, "reserved as number")
	ErrASNumberOutOfRange     = sdkerrors.Register(ModuleName, 1212, "as number out of range")
	ErrInvalidGeolocationData = sdkerrors.Register(ModuleName, 1213, "invalid geolocation data")
	ErrGeoCooldownPeriod      = sdkerrors.Register(ModuleName, 1214, "geo location cooldown period in effect")
)
