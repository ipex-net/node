package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/org/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		trustKeeper types.TrustKeeper

		// Collections for Org
		Schema        collections.Schema
		Orgs          collections.Map[uint64, types.Org]
		OrgsByAdmin   collections.Map[string, uint64]
		OrgsByCreator collections.Map[string, uint64]
		OrgSequence   collections.Sequence
		ParamsStore   collections.Item[types.Params]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

	trustKeeper types.TrustKeeper,

) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
		trustKeeper: trustKeeper,

		// Initialize collections
		Orgs:          collections.NewMap(sb, collections.NewPrefix(types.OrgKey), "orgs", collections.Uint64Key, codec.CollValue[types.Org](cdc)),
		OrgsByAdmin:   collections.NewMap(sb, collections.NewPrefix(types.OrgByAdminKey), "orgs_by_admin", collections.StringKey, collections.Uint64Value),
		OrgsByCreator: collections.NewMap(sb, collections.NewPrefix(types.OrgByCreatorKey), "orgs_by_creator", collections.StringKey, collections.Uint64Value),
		OrgSequence:   collections.NewSequence(sb, collections.NewPrefix(types.OrgSequenceKey), "org_sequence"),
		ParamsStore:   collections.NewItem(sb, collections.NewPrefix(types.ParamsKey), "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
