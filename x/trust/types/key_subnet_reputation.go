package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// SubnetReputationKeyPrefix is the prefix to retrieve all SubnetReputation
	SubnetReputationKeyPrefix = "SubnetReputation/value/"
)

// SubnetReputationKey returns the store key to retrieve a SubnetReputation from the index fields
func SubnetReputationKey(
	index uint64,
) []byte {
	var key []byte

	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
