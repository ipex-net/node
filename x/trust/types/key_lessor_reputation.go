package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// LessorReputationKeyPrefix is the prefix to retrieve all LessorReputation
	LessorReputationKeyPrefix = "LessorReputation/value/"
)

// LessorReputationKey returns the store key to retrieve a LessorReputation from the index fields
func LessorReputationKey(
	index uint64,
) []byte {
	var key []byte

	indexBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBytes, index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
