package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	leasetypes "ipex/x/lease/types"
	orgtypes "ipex/x/org/types"
	subnettypes "ipex/x/subnet/types"
)

// OrgKeeper defines the expected interface for the org module.
type OrgKeeper interface {
   // HasOrg returns true if an org with the given ID exists.
   HasOrg(ctx context.Context, id uint64) bool
   // VerifyOrgAdmin checks whether the given admin address is the admin of the org.
   GetOrg(ctx context.Context, id uint64) (orgtypes.Org, bool)
   // SetOrg updates the given Org.
   SetOrg(ctx context.Context, org orgtypes.Org)
   // GetOrgsForDomainValidation returns all orgs eligible for domain validation.
   GetOrgsForDomainValidation(ctx context.Context) []orgtypes.Org
   UpdateOrgDomainValidationStatus(ctx context.Context, orgID uint64, isValid bool) error
}

// SubnetKeeper defines the expected interface for the subnet module.
type SubnetKeeper interface {
   // CheckSubnetLeaseAvailability returns true if the subnet is available for leasing.
   CheckSubnetLeaseAvailability(ctx context.Context, subnetID uint64) bool
   // GetSubnet retrieves a Subnet by ID.
   GetSubnet(ctx context.Context, id uint64) (subnettypes.Subnet, bool)
   // SetSubnet updates the given Subnet.
   SetSubnet(ctx context.Context, subnet subnettypes.Subnet)
   GetSubnetsForRirSync(ctx context.Context) []subnettypes.Subnet
   UpdateSubnetRirDetails(ctx context.Context, subnetID uint64, info subnettypes.SubnetInfo) error 
   GetSubnetsForRPKIValidation(ctx context.Context) (list []subnettypes.Subnet)
   UpdateSubnetRpkiValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error
   GetSubnetsForGeoValidation(ctx context.Context) (list []subnettypes.Subnet)
   UpdateSubnetGeoValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error
}

// LeaseKeeper defines the expected interface for the lease module.
type LeaseKeeper interface {
   // GetContract retrieves a Contract by ID.
   GetContract(ctx context.Context, id uint64) (leasetypes.Contract, bool)
   // SetContract updates the given Contract.
   SetContract(ctx context.Context, contract leasetypes.Contract)
}

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

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
