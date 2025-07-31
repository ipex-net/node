package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"ipex/testutil/sample"
	"ipex/x/lease/keeper"
	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"
	subnettypes "ipex/x/subnet/types"
)

// Mock keepers for testing
type mockAccountKeeper struct{}

func (m mockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

type mockBankKeeper struct{}

func (m mockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

func (m mockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m mockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m mockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

type mockOrgKeeper struct{}

func (m mockOrgKeeper) HasOrg(ctx context.Context, id uint64) bool {
	return true
}

func (m mockOrgKeeper) VerifyOrgAdmin(ctx context.Context, orgid uint64, admin string) bool {
	return true
}

func (m mockOrgKeeper) GetOrgTreasury(ctx context.Context, orgid uint64) (sdk.AccAddress, bool) {
	// Return a mock treasury address for testing
	treasuryAddr, _ := sdk.AccAddressFromBech32(sample.AccAddress())
	return treasuryAddr, true
}

func (m mockOrgKeeper) GetOrgAdmin(ctx context.Context, orgid uint64) (sdk.AccAddress, bool) {
	return sdk.AccAddress{}, false
}

func (m mockOrgKeeper) IsOrgVerified(ctx context.Context, orgid uint64) bool {
	return true
}

type mockSubnetKeeperLease struct {
	// Configuration for different test scenarios
	GeoDbRecordVerified bool
	RirRecordValid      bool
	SubnetExists        bool
}

func (m mockSubnetKeeperLease) GetSubnet(ctx context.Context, id uint64) (subnettypes.Subnet, bool) {
	if !m.SubnetExists {
		return subnettypes.Subnet{}, false
	}
	
	// Mock subnet with configurable validation states
	subnet := subnettypes.Subnet{
		Id:                  id,
		NetworkAddress:      167772160, // 10.0.0.0
		Mask:                24,
		Country:             "US",
		Rir:                 "ARIN",
		Status:              subnettypes.SubnetStatus_SUBNET_STATUS_READY,
		RirRecordValid:      m.RirRecordValid,
		GeoDbRecordVerified: m.GeoDbRecordVerified,
	}
	return subnet, true
}

func (m mockSubnetKeeperLease) VerifySubnetOwner(ctx context.Context, subnetId uint64, user string) bool {
	return true
}


func (m mockSubnetKeeperLease) SetSubnetUserOrg(goCtx context.Context, subnetId uint64, userOrgId uint64, allowUserGeoChange bool, allowUserASNDROPOriginTransfer bool) {
}

func (m mockSubnetKeeperLease) RemoveSubnetUserOrg(goCtx context.Context, subnetId uint64) {
}

func (m mockSubnetKeeperLease) IsUserRequestedASNDROPOrigin(ctx context.Context, subnetId uint64) bool {
	return false
}

func (m mockSubnetKeeperLease) IsUserRefundAllowed(ctx context.Context, subnetId uint64) bool {
	return true
}

func (m mockSubnetKeeperLease) GetSubnetIpCount(ctx context.Context, subnetId uint64) uint64 {
	return 256 // Mock /24 subnet
}

type mockPaymentsKeeper struct{}

func (m mockPaymentsKeeper) RecordTransaction(
	ctx context.Context,
	transactionType paymentstypes.TransactionType,
	relatedModuleId paymentstypes.ModuleID,
	entityId uint64,
	amount sdk.Coin,
	fromOrgId uint64,
	toOrgId uint64,
	fromAddress string,
	toAddress string,
) uint64 {
	return 0
}

type mockTrustKeeperLease struct{}

func (m mockTrustKeeperLease) InitializeLeaseReputation(ctx context.Context, leaseId uint64) {
}

func (m mockTrustKeeperLease) RemoveLeaseReputation(ctx context.Context, leaseId uint64) {
}

func (m mockTrustKeeperLease) GetShortLeaseReputation(ctx context.Context, index uint64) (l1Listed bool, l2Count int32, failedAbuse uint64, found bool) {
	return false, 0, 0, false
}

func (m mockTrustKeeperLease) LogLesseeContractTermination(ctx context.Context, lesseeId uint64, lesseeContractViolation bool, ipCount uint64) error {
	return nil
}

func (m mockTrustKeeperLease) LogLesseeContractRenewal(ctx context.Context, lesseeId uint64, ipCount uint64) error {
	return nil
}

func (m mockTrustKeeperLease) LogLesseeContractActivation(ctx context.Context, lesseeId uint64, ipCount uint64) error {
	return nil
}

func (m mockTrustKeeperLease) LogLessorContractTermination(ctx context.Context, lessorId uint64, lessorContractViolation bool, ipCount uint64) error {
	return nil
}

func (m mockTrustKeeperLease) LogLessorContractRenewal(ctx context.Context, lessorId uint64, ipCount uint64) error {
	return nil
}

func (m mockTrustKeeperLease) LogLessorContractActivation(ctx context.Context, lessorId uint64, ipCount uint64) error {
	return nil
}

func LeaseKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
		mockAccountKeeper{},
		mockBankKeeper{},
		mockOrgKeeper{},
		mockSubnetKeeperLease{
			GeoDbRecordVerified: true,
			RirRecordValid:      true,
			SubnetExists:        true,
		},
		mockPaymentsKeeper{},
		mockTrustKeeperLease{},
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx
}

// LeaseKeeperWithMockSubnet creates a lease keeper with a custom mock subnet keeper
// for testing different subnet validation scenarios
func LeaseKeeperWithMockSubnet(t testing.TB, geoDbVerified, rirValid, subnetExists bool) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
		mockAccountKeeper{},
		mockBankKeeper{},
		mockOrgKeeper{},
		mockSubnetKeeperLease{
			GeoDbRecordVerified: geoDbVerified,
			RirRecordValid:      rirValid,
			SubnetExists:        subnetExists,
		},
		mockPaymentsKeeper{},
		mockTrustKeeperLease{},
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	return k, ctx
}
