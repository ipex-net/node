package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/subnet/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		orgKeeper   types.OrgKeeper
		trustKeeper types.TrustKeeper

		// Collections for Subnet
		Schema                   collections.Schema
		Subnets                  collections.Map[uint64, types.Subnet]
		SubnetsByOwnerOrg        collections.Map[collections.Pair[uint64, uint64], bool]
		SubnetsByUserOrg         collections.Map[collections.Pair[uint64, uint64], bool]
		SubnetsByNetworkAddress  collections.Map[collections.Pair[uint64, uint64], bool]
		SubnetSequence           collections.Sequence
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	orgKeeper types.OrgKeeper,
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
		orgKeeper:    orgKeeper,
		trustKeeper:  trustKeeper,
		
		// Initialize collections
		Subnets:                 collections.NewMap(sb, collections.NewPrefix(types.SubnetKey), "subnets", collections.Uint64Key, codec.CollValue[types.Subnet](cdc)),
		SubnetsByOwnerOrg:       collections.NewMap(sb, collections.NewPrefix(types.SubnetByOwnerOrgKey), "subnets_by_owner_org", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		SubnetsByUserOrg:        collections.NewMap(sb, collections.NewPrefix(types.SubnetByUserOrgKey), "subnets_by_user_org", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		SubnetsByNetworkAddress: collections.NewMap(sb, collections.NewPrefix(types.SubnetByNetworkAddressKey), "subnets_by_network_address", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		SubnetSequence:          collections.NewSequence(sb, collections.NewPrefix(types.SubnetSequenceKey), "subnet_sequence"),
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
