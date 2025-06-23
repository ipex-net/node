package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// ComplaintKeyPrefix is the prefix to retrieve all Complaint
	ComplaintKeyPrefix = "Complaint/value/"
)

// ComplaintKey returns the store key to retrieve a Complaint from the index fields
func ComplaintKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
