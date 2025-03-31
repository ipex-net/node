package types

const (
	// ModuleName defines the module name
	ModuleName = "payments"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_payments"
)

var (
	ParamsKey = []byte("p_payments")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	TransactionKey      = "Transaction/value/"
	TransactionCountKey = "Transaction/count/"
)

const (
    EventTypeTransactionRecorded = "transaction_recorded"

    AttributeKeyTransactionID   = "transaction_id"
    AttributeKeyTransactionType = "transaction_type"
    AttributeKeyModuleID        = "module_id"
    AttributeKeyEntityID        = "entity_id"
    AttributeKeyAmount          = "amount"
    AttributeKeyFromOrgID       = "from_org_id"
    AttributeKeyToOrgID         = "to_org_id"
    AttributeKeyFromAddress     = "from_address"
    AttributeKeyToAddress       = "to_address"
    AttributeKeyBlockHeight     = "block_height"
    AttributeKeyReferenceID     = "reference_id"
)