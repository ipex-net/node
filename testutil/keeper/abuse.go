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

	"ipex/x/abuse/keeper"
	"ipex/x/abuse/types"
	leasetypes "ipex/x/lease/types"
	subnettypes "ipex/x/subnet/types"
)

// Mock keepers for testing
type mockSubnetKeeper struct{}

func (m mockSubnetKeeper) VerifySubnetOwner(ctx context.Context, subnetId uint64, owner string) bool {
	return true
}

func (m mockSubnetKeeper) VerifySubnetUser(ctx context.Context, subnetId uint64, user string) bool {
	return true
}

func (m mockSubnetKeeper) GetSubnet(ctx context.Context, id uint64) (val subnettypes.Subnet, found bool) {
	return subnettypes.Subnet{}, false
}

type mockLeaseKeeper struct{}

func (m mockLeaseKeeper) GetFirstLeasedContractBySubnet(ctx context.Context, subnetId uint64) (uint64, bool) {
	return 0, false
}

func (m mockLeaseKeeper) GetContract(ctx context.Context, id uint64) (val leasetypes.Contract, found bool) {
	return leasetypes.Contract{}, false
}

type mockTrustKeeper struct{}

func (m mockTrustKeeper) UpdateAbuseStats(ctx context.Context, leaseId uint64, lesseeOrgId uint64, abuseCategory uint32, oldStatus string, newStatus string) error {
	return nil
}

func AbuseKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
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
		mockSubnetKeeper{},
		mockLeaseKeeper{},
		mockTrustKeeper{},
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	// Initialize complaint sequence to start from 1
	if err := k.ComplaintSequence.Set(ctx, 1); err != nil {
		panic(err)
	}

	return k, ctx
}
