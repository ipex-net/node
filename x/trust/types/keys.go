package types

const (
	// ModuleName defines the module name
	ModuleName = "trust"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_trust"
)

var (
	ParamsKey = []byte("p_trust")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
