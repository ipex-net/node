package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// BidKeyPrefix is the prefix to retrieve all Bid
	BidKeyPrefix = "Bid/value/"
)

// BidKey returns the store key to retrieve a Bid from the index fields
func BidKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
