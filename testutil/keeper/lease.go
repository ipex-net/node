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

	"ipex/x/lease/keeper"
	"ipex/x/lease/types"
	paymentstypes "ipex/x/payments/types"
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
	return sdk.AccAddress{}, false
}

func (m mockOrgKeeper) GetOrgAdmin(ctx context.Context, orgid uint64) (sdk.AccAddress, bool) {
	return sdk.AccAddress{}, false
}

func (m mockOrgKeeper) IsOrgVerified(ctx context.Context, orgid uint64) bool {
	return true
}

type mockSubnetKeeperLease struct{}

func (m mockSubnetKeeperLease) VerifySubnetOwner(ctx context.Context, subnetId uint64, user string) bool {
	return true
}

func (m mockSubnetKeeperLease) CheckSubnetLeaseAvailability(ctx context.Context, subnetId uint64) bool {
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
		mockSubnetKeeperLease{},
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
