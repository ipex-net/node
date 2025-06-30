package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/trust/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// Collections for BlacklistLog
		Schema                collections.Schema
		BlacklistLogs         collections.Map[uint64, types.BlacklistLog]
		BlacklistLogsBySubnet collections.Map[collections.Pair[uint64, uint64], bool]
		BlacklistLogsByLessor collections.Map[collections.Pair[uint64, uint64], bool]
		BlacklistLogsByLessee collections.Map[collections.Pair[uint64, uint64], bool]
		BlacklistLogsByLease  collections.Map[collections.Pair[uint64, uint64], bool]
		BlacklistLogSequence  collections.Sequence
		
		// Collections for Reputation Systems
		LeaseReputations      collections.Map[uint64, types.LeaseReputation]
		LesseeReputations     collections.Map[uint64, types.LesseeReputation]
		SubnetReputations     collections.Map[uint64, types.SubnetReputation]
		ParamsStore           collections.Item[types.Params]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

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

		// Initialize collections
		BlacklistLogs:         collections.NewMap(sb, collections.NewPrefix(types.BlacklistLogKey), "blacklist_logs", collections.Uint64Key, codec.CollValue[types.BlacklistLog](cdc)),
		BlacklistLogsBySubnet: collections.NewMap(sb, collections.NewPrefix(types.BlacklistLogBySubnetKey), "blacklist_logs_by_subnet", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		BlacklistLogsByLessor: collections.NewMap(sb, collections.NewPrefix(types.BlacklistLogByLessorKey), "blacklist_logs_by_lessor", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		BlacklistLogsByLessee: collections.NewMap(sb, collections.NewPrefix(types.BlacklistLogByLesseeKey), "blacklist_logs_by_lessee", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		BlacklistLogsByLease:  collections.NewMap(sb, collections.NewPrefix(types.BlacklistLogByLeaseKey), "blacklist_logs_by_lease", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		BlacklistLogSequence:  collections.NewSequence(sb, collections.NewPrefix(types.BlacklistLogSequenceKey), "blacklist_log_sequence"),
		
		// Initialize reputation collections
		LeaseReputations:      collections.NewMap(sb, collections.NewPrefix(types.LeaseReputationKeyPrefix), "lease_reputations", collections.Uint64Key, codec.CollValue[types.LeaseReputation](cdc)),
		LesseeReputations:     collections.NewMap(sb, collections.NewPrefix(types.LesseeReputationKeyPrefix), "lessee_reputations", collections.Uint64Key, codec.CollValue[types.LesseeReputation](cdc)),
		SubnetReputations:     collections.NewMap(sb, collections.NewPrefix(types.SubnetReputationKeyPrefix), "subnet_reputations", collections.Uint64Key, codec.CollValue[types.SubnetReputation](cdc)),
		ParamsStore:           collections.NewItem(sb, collections.NewPrefix(types.ParamsKey), "params", codec.CollValue[types.Params](cdc)),
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
