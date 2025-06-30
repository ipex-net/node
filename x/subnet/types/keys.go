package types

const (
	// ModuleName defines the module name
	ModuleName = "subnet"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_subnet"
)

var (
	ParamsKey = []byte("p_subnet")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	SubnetKey      = "Subnet/value/"
	SubnetCountKey = "Subnet/count/"
	
	// Collection keys for secondary indexes
	SubnetByOwnerOrgKey       = "Subnet/by_owner_org/"
	SubnetByUserOrgKey        = "Subnet/by_user_org/"
	SubnetByNetworkAddressKey = "Subnet/by_network_address/"
	SubnetSequenceKey         = "Subnet/sequence/"
)
