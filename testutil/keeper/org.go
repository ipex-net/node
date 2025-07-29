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

	"ipex/x/org/keeper"
	"ipex/x/org/types"
)

// Mock keepers for testing
type mockTrustKeeperOrg struct{}

func (m mockTrustKeeperOrg) InitializeLesseeReputation(ctx context.Context, lesseeId uint64) {
}

func (m mockTrustKeeperOrg) RemoveLesseeReputation(ctx context.Context, index uint64) {
}

func (m mockTrustKeeperOrg) InitializeLessorReputation(ctx context.Context, lessorId uint64) {
}

func (m mockTrustKeeperOrg) RemoveLessorReputation(ctx context.Context, index uint64) {
}

func OrgKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
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
		mockTrustKeeperOrg{},
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(err)
	}

	// Initialize org sequence to start from 1
	if err := k.OrgSequence.Set(ctx, 1); err != nil {
		panic(err)
	}

	return k, ctx
}
