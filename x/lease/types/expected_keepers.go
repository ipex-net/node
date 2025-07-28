package types

import (
	"context"
	paymentstypes "ipex/x/payments/types"

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
	SendCoinsFromAccountToModule(context.Context, sdk.AccAddress, string, sdk.Coins) error
	SendCoinsFromModuleToAccount(context.Context, string, sdk.AccAddress, sdk.Coins) error
	SendCoins(context.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}

type OrgKeeper interface {
	HasOrg(ctx context.Context, id uint64) bool
	VerifyOrgAdmin(ctx context.Context, orgid uint64, admin string) bool
	GetOrgTreasury(ctx context.Context, orgid uint64) (sdk.AccAddress, bool)
	GetOrgAdmin(ctx context.Context, orgid uint64) (sdk.AccAddress, bool)
	IsOrgVerified(ctx context.Context, id uint64) bool
}

type SubnetKeeper interface {
	VerifySubnetOwner(ctx context.Context, subnetId uint64, user string) bool
	CheckSubnetLeaseAvailability(ctx context.Context, subnetId uint64) bool
	SetSubnetUserOrg(goCtx context.Context, subnetId uint64, userOrgId uint64, allowUserGeoChange bool, allowUserASNDROPOriginTransfer bool)
	RemoveSubnetUserOrg(goCtx context.Context, subnetId uint64)
	IsUserRequestedASNDROPOrigin(ctx context.Context, subnetId uint64) bool
	IsUserRefundAllowed(ctx context.Context, subnetId uint64) bool
	GetSubnetIpCount(ctx context.Context, subnetId uint64) uint64
}

type PaymentsKeeper interface {
	RecordTransaction(
		ctx context.Context,
		transactionType paymentstypes.TransactionType,
		relatedModuleId paymentstypes.ModuleID,
		entityId uint64,
		amount sdk.Coin,
		fromOrgId uint64,
		toOrgId uint64,
		fromAddress string,
		toAddress string,
	) uint64
}

type TrustKeeper interface {
	InitializeLeaseReputation(ctx context.Context, leaseId uint64)
	RemoveLeaseReputation(ctx context.Context, leaseId uint64)
	GetShortLeaseReputation(ctx context.Context, index uint64) (l1Listed bool, l2Count int32, failedAbuse uint64, found bool)
	LogLesseeContractTermination(ctx context.Context, lesseeId uint64, violated bool, ipCount uint64) error
	LogLesseeContractRenewal(ctx context.Context, lesseeId uint64, ipCount uint64) error
	LogLesseeContractActivation(ctx context.Context, lesseeId uint64, ipCount uint64) error
}
