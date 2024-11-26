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
