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
