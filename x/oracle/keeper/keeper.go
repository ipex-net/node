package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"ipex/x/oracle/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		orgKeeper    types.OrgKeeper
		subnetKeeper types.SubnetKeeper
		leaseKeeper  types.LeaseKeeper
	}
)
const LastBlockTimeKey = "last_block_time"

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

	orgKeeper types.OrgKeeper,
	subnetKeeper types.SubnetKeeper,
	leaseKeeper types.LeaseKeeper,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,

		orgKeeper:    orgKeeper,
		subnetKeeper: subnetKeeper,
		leaseKeeper:  leaseKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}


// SetLastBlockTime stores only the most recent block time
func (k Keeper) SetLastBlockTime(ctx sdk.Context, blockTime time.Time) {
    store := k.storeService.OpenKVStore(ctx)
    
    timeBytes, err := blockTime.MarshalBinary()
    if err != nil {
        panic(err)
    }
    
    store.Set([]byte(LastBlockTimeKey), timeBytes)
}

// GetLastBlockTime retrieves the most recent block time
func (k Keeper) GetLastBlockTime(ctx sdk.Context) (time.Time, bool) {
    store := k.storeService.OpenKVStore(ctx)
    
    timeBytes, err := store.Get([]byte(LastBlockTimeKey))
    if err != nil || timeBytes == nil {
        return time.Time{}, false
    }
    
    var blockTime time.Time
    err = blockTime.UnmarshalBinary(timeBytes)
    if err != nil {
        return time.Time{}, false
    }
    
    return blockTime, true
}