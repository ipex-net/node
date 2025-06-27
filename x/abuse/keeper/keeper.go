package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/abuse/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		subnetKeeper types.SubnetKeeper
		leaseKeeper  types.LeaseKeeper
		trustKeeper  types.TrustKeeper

		// Collections for Complaint
		Schema                collections.Schema
		Complaints            collections.Map[uint64, types.Complaint]
		ComplaintsBySubnet    collections.Map[collections.Pair[uint64, uint64], bool]
		ComplaintsByLease     collections.Map[collections.Pair[uint64, uint64], bool]
		ComplaintsByLessee    collections.Map[collections.Pair[uint64, uint64], bool]
		ComplaintsByLessor    collections.Map[collections.Pair[uint64, uint64], bool]
		ComplaintSequence     collections.Sequence
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	subnetKeeper types.SubnetKeeper,
	leaseKeeper types.LeaseKeeper,
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
		subnetKeeper: subnetKeeper,
		leaseKeeper:  leaseKeeper,
		trustKeeper:  trustKeeper,

		// Initialize collections
		Complaints:         collections.NewMap(sb, collections.NewPrefix(types.ComplaintKeyPrefix), "complaints", collections.Uint64Key, codec.CollValue[types.Complaint](cdc)),
		ComplaintsBySubnet: collections.NewMap(sb, collections.NewPrefix(types.ComplaintBySubnetKey), "complaints_by_subnet", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ComplaintsByLease:  collections.NewMap(sb, collections.NewPrefix(types.ComplaintByLeaseKey), "complaints_by_lease", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ComplaintsByLessee: collections.NewMap(sb, collections.NewPrefix(types.ComplaintByLesseeKey), "complaints_by_lessee", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ComplaintsByLessor: collections.NewMap(sb, collections.NewPrefix(types.ComplaintByLessorKey), "complaints_by_lessor", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ComplaintSequence:  collections.NewSequence(sb, collections.NewPrefix(types.ComplaintSequenceKey), "complaint_sequence"),
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
