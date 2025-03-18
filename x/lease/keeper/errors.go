package keeper

// DONTCOVER
import (
	sdkerrors "cosmossdk.io/errors"
)

const ModuleName = "lease"

// x/subnet module sentinel errors
var (
	ErrInvalidOrgId        		= sdkerrors.Register(ModuleName, 1208, "invalid org id")
	ErrNotOrgAdmin         		= sdkerrors.Register(ModuleName, 1209, "not organization admin")
	ErrSubnetNotFound      		= sdkerrors.Register(ModuleName, 1210, "invalid subnet id")
	ErrNotSubnetOwnerAdmin 		= sdkerrors.Register(ModuleName, 1211, "not subnet owner admin")
	ErrSubnetNotReady      		= sdkerrors.Register(ModuleName, 1212, "subnet is not ready for lease")
	ErrInsufficientFunds   		= sdkerrors.Register(ModuleName, 1213, "insufficient funds for the operation")
	ErrFundTransfer        		= sdkerrors.Register(ModuleName, 1214, "error during fund transfer")
	ErrInvalidDeposit 	   		= sdkerrors.Register(ModuleName, 1215, "invalid safety deposit")
	ErrInvalidRevocationPeriod 	= sdkerrors.Register(ModuleName, 1217, "invalid revocation period")
	ErrInvalidCurrency 			= sdkerrors.Register(ModuleName, 1218, "invalid currency denomination")

)
