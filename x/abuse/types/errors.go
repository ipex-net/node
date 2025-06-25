package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/abuse module sentinel errors
var (
	ErrInvalidRequest      = errorsmod.Register(ModuleName, 1100, "invalid request")
	ErrInvalidAbuseType    = errorsmod.Register(ModuleName, 1101, "invalid abuse type")
	ErrComplaintNotFound   = errorsmod.Register(ModuleName, 1102, "complaint not found")
	ErrUnauthorized        = errorsmod.Register(ModuleName, 1103, "unauthorized")
	ErrInvalidDeadline     = errorsmod.Register(ModuleName, 1104, "invalid deadline")
	ErrComplaintClosed     = errorsmod.Register(ModuleName, 1105, "complaint already closed")
	ErrInvalidState        = errorsmod.Register(ModuleName, 1106, "invalid state transition")
	ErrInvalidSigner       = errorsmod.Register(ModuleName, 1107, "expected gov account as only signer for proposal message")
	ErrInvalidIncidentTime = errorsmod.Register(ModuleName, 1108, "invalid incident time")
	ErrInvalidLeaseId      = errorsmod.Register(ModuleName, 1109, "invalid lease id")
	ErrInvalidSubnetId     = errorsmod.Register(ModuleName, 1110, "invalid subnet id")
	ErrInvalidSummary      = errorsmod.Register(ModuleName, 1111, "invalid summary")
	ErrInvalidIP           = errorsmod.Register(ModuleName, 1112, "invalid IP address")
	ErrLeaseNotActive      = errorsmod.Register(ModuleName, 1113, "lease is not active")
	ErrInvalidTicketRef    = errorsmod.Register(ModuleName, 1114, "invalid ticket reference")
	ErrInvalidEvidenceHash = errorsmod.Register(ModuleName, 1115, "invalid evidence hash")
)
