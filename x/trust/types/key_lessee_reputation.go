package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LesseeReputationKeyPrefix is the prefix to retrieve all LesseeReputation
	LesseeReputationKeyPrefix = "LesseeReputation/value/"
)

// LesseeReputationKey returns the store key to retrieve a LesseeReputation from the index fields
func LesseeReputationKey(
	index uint64,
) []byte {
	var key []byte

	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
