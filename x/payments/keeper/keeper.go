package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/payments/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		// Collections for Transaction
		Schema               collections.Schema
		Transactions         collections.Map[uint64, types.Transaction]
		TransactionsByEntity collections.Map[collections.Pair[uint64, uint64], bool]
		TransactionsByOrg    collections.Map[collections.Pair[uint64, uint64], bool]
		TransactionSequence  collections.Sequence
		ParamsStore          collections.Item[types.Params]
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
		Transactions:         collections.NewMap(sb, collections.NewPrefix(types.TransactionKey), "transactions", collections.Uint64Key, codec.CollValue[types.Transaction](cdc)),
		TransactionsByEntity: collections.NewMap(sb, collections.NewPrefix(types.TransactionByEntityKey), "transactions_by_entity", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		TransactionsByOrg:    collections.NewMap(sb, collections.NewPrefix(types.TransactionByOrgKey), "transactions_by_org", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.BoolValue),
		TransactionSequence:  collections.NewSequence(sb, collections.NewPrefix(types.TransactionSequenceKey), "transaction_sequence"),
		ParamsStore:          collections.NewItem(sb, collections.NewPrefix(types.ParamsKey), "params", codec.CollValue[types.Params](cdc)),
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
