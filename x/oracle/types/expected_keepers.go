package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	leasetypes "ipex/x/lease/types"
	orgtypes "ipex/x/org/types"
	subnettypes "ipex/x/subnet/types"
	trusttypes "ipex/x/trust/types"
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
	GetOrgsForDomainValidation(ctx context.Context, referenceTime time.Time) []orgtypes.Org
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
	GetSubnetsForRirSync(ctx context.Context, referenceTime time.Time) []subnettypes.Subnet
	UpdateSubnetRirDetails(ctx context.Context, subnetID uint64, info subnettypes.SubnetInfo) error
	GetSubnetsForRPKIValidation(ctx context.Context, referenceTime time.Time) (list []subnettypes.Subnet)
	UpdateSubnetRpkiValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error
	GetSubnetsForGeoValidation(ctx context.Context, referenceTime time.Time) (list []subnettypes.Subnet)
	UpdateSubnetGeoValidationStatus(ctx context.Context, subnetID uint64, isValid bool) error
	GetSubnetsForOwnershipVerification(ctx context.Context, referenceTime time.Time) (list []subnettypes.Subnet)
	UpdateSubnetOwnershipVerification(ctx context.Context, subnetId uint64, isValid bool) error
	IsRirRecordValid(ctx context.Context, subnetId uint64) bool
	ClearSubnetOriginCheckPending(ctx context.Context, subnetId uint64) error
	GetOriginCheckPendingSubnets(ctx context.Context, limit int) ([]uint64)
	FlagUserRequestedASNDROPOrigin(ctx context.Context, subnetId uint64) error
}

// LeaseKeeper defines the expected interface for the lease module.
type LeaseKeeper interface {
	// GetContract retrieves a Contract by ID.
	GetContract(ctx context.Context, id uint64) (leasetypes.Contract, bool)
	// SetContract updates the given Contract.
	SetContract(ctx context.Context, contract leasetypes.Contract)
	// Returns the contract ID and true if found, or 0 and false if not found
	GetFirstLeasedContractBySubnet(ctx context.Context, subnetId uint64) (uint64, bool)
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

type NetIntelService interface {
	// Service status
	IsReady() bool
	GetLastUpdate() time.Time
	GetLoadedLevels() []int
	GetLoadedCountries() []string

	// IP reputation checking
	CheckReputation(ip string, level int) int
	ScanSubnetReputation(subnet string, level int) int

	// Geolocation checking
	CheckCountry(ip string, countryCode string) bool
}

type TrustKeeper interface {
	GetSubnetsForReputationScanning(ctx context.Context, referenceTime time.Time) (list []uint64)
	UpdateSubnetReputation(ctx context.Context, subnetId uint64, leaseId uint64, lessorOrgId uint64, lesseeOrgId uint64, reputation trusttypes.SubnetBlacklistCount) error
}
