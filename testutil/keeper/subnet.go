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

	"ipex/x/subnet/keeper"
	"ipex/x/subnet/types"
)

// Mock keepers for testing
type mockOrgKeeperSubnet struct{}

func (m mockOrgKeeperSubnet) HasOrg(ctx context.Context, id uint64) bool {
	return true
}

func (m mockOrgKeeperSubnet) VerifyOrgAdmin(ctx context.Context, orgid uint64, admin string) bool {
	return true
}

type mockTrustKeeperSubnet struct{}

func (m mockTrustKeeperSubnet) InitializeSubnetReputation(ctx context.Context, subnetID uint64) {
}

func (m mockTrustKeeperSubnet) RemoveSubnetReputation(ctx context.Context, index uint64) {
}

func SubnetKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
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
		mockOrgKeeperSubnet{},
		mockTrustKeeperSubnet{},
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	// Initialize subnet sequence to start from 1
	if err := k.SubnetSequence.Set(ctx, 1); err != nil {
		panic(err)
	}

	return k, ctx
}
