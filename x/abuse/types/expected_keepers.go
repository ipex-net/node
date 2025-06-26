package types

import (
	"context"

	leasetypes "ipex/x/lease/types"
	subnettypes "ipex/x/subnet/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type SubnetKeeper interface {
	VerifySubnetOwner(ctx context.Context, subnetId uint64, owner string) bool
	VerifySubnetUser(ctx context.Context, subnetId uint64, user string) bool
	GetSubnet(ctx context.Context, id uint64) (val subnettypes.Subnet, found bool)
}

type LeaseKeeper interface {
	GetFirstLeasedContractBySubnet(ctx context.Context, subnetId uint64) (uint64, bool)
	GetContract(ctx context.Context, id uint64) (val leasetypes.Contract, found bool)
}

type TrustKeeper interface {
	UpdateAbuseStats(ctx context.Context, leaseId uint64, lesseeOrgId uint64, abuseCategory uint32, oldStatus string, newStatus string) error
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
