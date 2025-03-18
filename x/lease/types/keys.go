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

	// ModuleAccountName is the name of the module account
	ModuleAccountName = ModuleName
)

var (
	ParamsKey = []byte("p_lease")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	ContractCreatedEventType                     = "new-contract-created"
	ContractCreatedEventCreator                  = "creator"
	ContractCreatedEventContractId               = "contract-id"
	ContractCreatedEventSubnetId                 = "subnet-id"
	ContractCreatedEventPrice                    = "price"             //Subnet lease price amount
	ContractCreatedEventPriceDenom               = "price-denom"       //Subnet lease price denom
	ContractCreatedEventDepositAmount            = "deposit-amount"    //Deposit amount
	ContractCreatedEventDepositDenom             = "deposit-denom"     //Deposit denom
	ContractCreatedEventRevocationPeriodDuration = "revocation-period" //Duration notify before owner-initiated termination
)

const (
	ContractUpdatedEventType       = "contract-updated"
	ContractUpdatedEventCreator    = "creator"
	ContractUpdatedEventContractId = "contract-id"
	ContractUpdatedEventPrice      = "price"       //Subnet lease price amount
	ContractUpdatedEventPriceDenom = "price-denom" //Subnet lease price denom
)

const (
	ContractDeletedEventType       = "contract-deleted"
	ContractDeletedEventCreator    = "creator"
	ContractDeletedEventContractId = "contract-id"
)

const (
	ContractCanceledEventType  = "contract-canceled"
	ContractCanceledCreator    = "creator"
	ContractCanceledContractId = "contract-id"
)

const (
	ContractRevokedEventType             = "contract-revoked"
	ContractRevokedCreator               = "creator"
	ContractRevokedContractId            = "contract-id"
	ContractRevokedRevocationEffectiveAt = "revocation-effective-at"
)

const (
	ContractEnrolledEventType                = "contract-enrolled"
	ContractEnrolledCreator                  = "creator"
	ContractEnrolledContractId               = "contract-id"
	ContractEnrolledLesseeOrg                = "lessee-org"
	ContractEnrolledPriceAmount              = "price-amount"
	ContractEnrolledPriceDenom               = "price-denom"
	ContractSafetyDepositCollectedAmount     = "deposit-collected-amount"
	ContractSafetyDepositCollectedDenom      = "deposit-collected-denom"
)

const (
	EnrollementRevertedEventType                = "contract-enrolled"
	EnrollementRevertedContractId               = "contract-id"
	EnrollementRevertedLesseeOrg                = "lessee-org"
	EnrollementRevertedPriceAmount              = "price-amount"
	EnrollementRevertedPriceDenom               = "price-denom"
)

const (
    EscrowDepositEventType   = "escrow-deposit"
    EscrowDepositEventSender        = "sender"
    EscrowDepositEventAmount        = "amount"
)
