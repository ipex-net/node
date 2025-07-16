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

	// Collection keys for secondary indexes
	ContractByStatusKey = "Contract/by_status/"
	ContractBySubnetKey = "Contract/by_subnet/"
	ContractByOwnerKey  = "Contract/by_owner/"
	ContractByLesseeKey = "Contract/by_lessee/"
	ContractSequenceKey = "Contract/sequence/"

	// ModuleAccountName is the name of the module account
	ModuleAccountName = ModuleName
	
	// Bid collection keys
	BidKeyPrefix     = "Bid/value/"
	BidByContractKey = "Bid/by_contract/"
	BidByBidderKey   = "Bid/by_bidder/"
)

var (
	ParamsKey = []byte("p_lease")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	ContractCreatedEventType                     = "new_contract_created"
	ContractCreatedEventCreator                  = "creator"
	ContractCreatedEventContractId               = "contract_id"
	ContractCreatedEventSubnetId                 = "subnet_id"
	ContractCreatedEventPrice                    = "price"             //Subnet lease price amount
	ContractCreatedEventPriceDenom               = "price_denom"       //Subnet lease price denom
	ContractCreatedEventDepositAmount            = "deposit_amount"    //Deposit amount
	ContractCreatedEventDepositDenom             = "deposit_denom"     //Deposit denom
	ContractCreatedEventRevocationPeriodDuration = "revocation_period" //Duration notify before owner-initiated termination
)

const (
	ContractUpdatedEventType       = "contract_updated"
	ContractUpdatedEventCreator    = "creator"
	ContractUpdatedEventContractId = "contract_id"
	ContractUpdatedEventPrice      = "price"       //Subnet lease price amount
	ContractUpdatedEventPriceDenom = "price_denom" //Subnet lease price denom
)

const (
	ContractDeletedEventType       = "contract_deleted"
	ContractDeletedEventCreator    = "creator"
	ContractDeletedEventContractId = "contract_id"
)

const (
	ContractCanceledEventType  = "contract_canceled"
	ContractCanceledCreator    = "creator"
	ContractCanceledContractId = "contract_id"
)
const (
	ContractCancellationRevokedEventType  = "contract_cancellation_revoked"
	ContractCancellationRevokedCreator    = "creator"
	ContractCancellationRevokedContractId = "contract_id"
)

const (
	ContractRevokedEventType             = "contract_revoked"
	ContractRevokedCreator               = "creator"
	ContractRevokedContractId            = "contract_id"
	ContractRevokedRevocationEffectiveAt = "revocation_effective_at"
)

const (
	ContractEnrolledEventType            = "contract_enrolled"
	ContractEnrolledCreator              = "creator"
	ContractEnrolledContractId           = "contract_id"
	ContractEnrolledLesseeOrg            = "lessee_org"
	ContractEnrolledPriceAmount          = "price_amount"
	ContractEnrolledPriceDenom           = "price_denom"
	ContractSafetyDepositCollectedAmount = "deposit_collected_amount"
	ContractSafetyDepositCollectedDenom  = "deposit_collected_denom"
)
const (
	ContractTerminatedByOwnerEventType  = "contract_terminated_owner"
	ContractTerminatedByOwnerCreator    = "creator"
	ContractTerminatedByOwnerContractId = "contract_id"
	ContractTerminatedByOwnerType       = "termination_type"
	ContractTerminatedByOwnerAt         = "terminated_at"
)
const (
	ContractRenewedEventType       = "contract_renewed"
	ContractRenewedCreator         = "creator"
	ContractRenewedContractId      = "contract_id"
	ContractRenewedLesseeOrg       = "lessee_org"
	ContractRenewedPriceAmount     = "price_amount"
	ContractRenewedPriceDenom      = "price_denom"
	ContractRenewedPrepaidUntil    = "prepaid_until"
	ContractRenewedDepositRetained = "deposit_retained"
)
