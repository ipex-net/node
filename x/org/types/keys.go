package types

const (
	// ModuleName defines the module name
	ModuleName = "org"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_org"
)

var (
	ParamsKey = []byte("p_org")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	OrgKey      = "Org/value/"
	OrgCountKey = "Org/count/"
)
