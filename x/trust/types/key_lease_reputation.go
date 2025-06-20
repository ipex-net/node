package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LeaseReputationKeyPrefix is the prefix to retrieve all LeaseReputation
	LeaseReputationKeyPrefix = "LeaseReputation/value/"
)

// LeaseReputationKey returns the store key to retrieve a LeaseReputation from the index fields
func LeaseReputationKey(
	index uint64,
) []byte {
	var key []byte

	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
