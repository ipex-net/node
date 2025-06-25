package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// ComplaintKeyPrefix is the prefix to retrieve all Complaint
	ComplaintKeyPrefix = "Complaint/value/"
	// ComplaintCountKey is the key for the complaint count
	ComplaintCountKey = "Complaint/count/"
)

// ComplaintKey returns the store key to retrieve a Complaint from the id
func ComplaintKey(
	id uint64,
) []byte {
	var key []byte

	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, id)
	key = append(key, idBytes...)

	return key
}

// GetComplaintIDBytes returns the byte representation of the complaint ID
func GetComplaintIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}
