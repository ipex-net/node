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

const (
	BlacklistLogKey         = "BlacklistLog/value/"
	BlacklistLogCountKey    = "BlacklistLog/count/"
	BlacklistLogBySubnetKey = "BlacklistLog/by_subnet/"
	BlacklistLogByLessorKey = "BlacklistLog/by_lessor/"
	BlacklistLogByLesseeKey = "BlacklistLog/by_lessee/"
	BlacklistLogByLeaseKey  = "BlacklistLog/by_lease/"
	BlacklistLogSequenceKey = "BlacklistLog/sequence/"
)
