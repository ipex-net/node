package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/lease/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		accountKeeper  types.AccountKeeper
		bankKeeper     types.BankKeeper
		orgKeeper      types.OrgKeeper
		subnetKeeper   types.SubnetKeeper
		paymentsKeeper types.PaymentsKeeper
		trustKeeper    types.TrustKeeper

		// Collections for Contract
		Schema            collections.Schema
		Contracts         collections.Map[uint64, types.Contract]
		ContractsByStatus collections.Map[collections.Pair[int32, uint64], bool]
		ContractsBySubnet collections.Map[collections.Pair[uint64, uint64], bool]
		ContractsByOwner  collections.Map[collections.Pair[uint64, uint64], bool]
		ContractsByLessee collections.Map[collections.Pair[uint64, uint64], bool]
		ContractSequence  collections.Sequence
		ParamsStore       collections.Item[types.Params]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	orgKeeper types.OrgKeeper,
	subnetKeeper types.SubnetKeeper,
	paymentsKeeper types.PaymentsKeeper,
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

		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		orgKeeper:      orgKeeper,
		subnetKeeper:   subnetKeeper,
		paymentsKeeper: paymentsKeeper,
		trustKeeper:    trustKeeper,

		// Initialize collections
		Contracts:         collections.NewMap(sb, collections.NewPrefix(types.ContractKey), "contracts", collections.Uint64Key, codec.CollValue[types.Contract](cdc)),
		ContractsByStatus: collections.NewMap(sb, collections.NewPrefix(types.ContractByStatusKey), "contracts_by_status", collections.PairKeyCodec(collections.Int32Key, collections.Uint64Key), collections.BoolValue),
		ContractsBySubnet: collections.NewMap(sb, collections.NewPrefix(types.ContractBySubnetKey), "contracts_by_subnet", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ContractsByOwner:  collections.NewMap(sb, collections.NewPrefix(types.ContractByOwnerKey), "contracts_by_owner", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ContractsByLessee: collections.NewMap(sb, collections.NewPrefix(types.ContractByLesseeKey), "contracts_by_lessee", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		ContractSequence:  collections.NewSequence(sb, collections.NewPrefix(types.ContractSequenceKey), "contract_sequence"),
		ParamsStore:       collections.NewItem(sb, collections.NewPrefix(types.ParamsKey), "params", codec.CollValue[types.Params](cdc)),
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
