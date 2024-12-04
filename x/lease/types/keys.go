package types

const (
	// ModuleName defines the module name
	ModuleName = "lease"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_lease"

	// ContractKey is used to uniquely identify contracts within the system.
	ContractKey = "Contract/value/"

	// ContractCountKey this key will be used to keep track of the ID of the latest contract added to the store.
	ContractCountKey = "Contract/count/"
)

var (
	ParamsKey = []byte("p_lease")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
const (
	ContractCreatedEventType      = "new-contract-created"
	ContractCreatedEventCreator   = "creator"
	ContractCreatedEventContractId = "contract-id"
	ContractCreatedEventNetworkAddress 	= "network-address"
	ContractCreatedEventPrefixLength 	= "prefix-length"
	ContractCreatedEventRir 	= "rir" 					// Subnet's RIR
	ContractCreatedEventCountryCode 	= "country-code" 	//Subnet's country code in geofeed
	ContractCreatedEventGeofeedUrl 	= "geofeed-url"			//Geofeed URL
	ContractCreatedEventPrice 	= "price"					//Subnet lease price amount
	ContractCreatedEventPriceDenom 	= "price-denom"			//Subnet lease price denom
	ContractCreatedEventDepositAmount 	= "deposit-amount"	//Deposit amount
	ContractCreatedEventDepositDenom = "deposit-denom"				//Deposit denom
	ContractCreatedEventNoticePeriodBlocks 	= "notice-period-blocks" // Number of blocks to notify before owner-initiated termination
)

const (
	ContractUpdatedEventType      = "contract-updated"
	ContractUpdatedEventCreator   = "creator"
	ContractUpdatedEventContractId = "contract-id"
	ContractUpdatedEventPrice 	= "price"					//Subnet lease price amount
	ContractUpdatedEventPriceDenom 	= "price-denom"			//Subnet lease price denom
	ContractUpdatedEventDepositAmount 	= "deposit-amount"	//Deposit amount
	ContractUpdatedEventDepositDenom = "deposit-denom"				//Deposit denom
	ContractUpdatedEventNoticePeriodBlocks 	= "notice-period-blocks" // Number of blocks to notify before owner-initiated termination
)

const (
	ContractDeletedEventType      = "contract-deleted"
	ContractDeletedEventCreator   = "creator"
	ContractDeletedEventContractId = "contract-id"
)