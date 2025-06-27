package keeper

import (
	"context"
	"encoding/binary"

	"ipex/x/abuse/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// AppendComplaint appends a complaint to the store with an auto-incrementing id
func (k Keeper) AppendComplaint(ctx context.Context, complaint types.Complaint) uint64 {
	// Get next ID from sequence
	id, err := k.ComplaintSequence.Next(ctx)
	if err != nil {
		panic(err)
	}
	complaint.Id = id

	// Store the complaint
	err = k.Complaints.Set(ctx, id, complaint)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes
	k.updateComplaintIndexes(ctx, complaint)

	return id
}

// updateComplaintIndexes updates all secondary indexes for a complaint
func (k Keeper) updateComplaintIndexes(ctx context.Context, complaint types.Complaint) {
	// Index by subnet
	err := k.ComplaintsBySubnet.Set(ctx, collections.Join(complaint.SubnetId, complaint.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by lease
	err = k.ComplaintsByLease.Set(ctx, collections.Join(complaint.LeaseId, complaint.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by lessee
	err = k.ComplaintsByLessee.Set(ctx, collections.Join(complaint.LesseeOrgId, complaint.Id), true)
	if err != nil {
		panic(err)
	}

	// Index by lessor
	err = k.ComplaintsByLessor.Set(ctx, collections.Join(complaint.LessorOrgId, complaint.Id), true)
	if err != nil {
		panic(err)
	}
}

// GetComplaintCount returns the total number of complaints
func (k Keeper) GetComplaintCount(ctx context.Context) uint64 {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ComplaintCountKey)
	bz := store.Get(byteKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// SetComplaintCount sets the total number of complaints
func (k Keeper) SetComplaintCount(ctx context.Context, count uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	byteKey := types.KeyPrefix(types.ComplaintCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// SetComplaint set a specific complaint in the store from its id
func (k Keeper) SetComplaint(ctx context.Context, complaint types.Complaint) {
	// Store the complaint
	err := k.Complaints.Set(ctx, complaint.Id, complaint)
	if err != nil {
		panic(err)
	}

	// Update secondary indexes (in case any indexed fields changed)
	k.updateComplaintIndexes(ctx, complaint)
}

// GetComplaint returns a complaint from its id
func (k Keeper) GetComplaint(
	ctx context.Context,
	id uint64,
) (val types.Complaint, found bool) {
	complaint, err := k.Complaints.Get(ctx, id)
	if err != nil {
		if err.Error() == "collections: not found" {
			return val, false
		}
		panic(err)
	}
	return complaint, true
}

// RemoveComplaint removes a complaint from the store
func (k Keeper) RemoveComplaint(
	ctx context.Context,
	id uint64,
) {
	// Get the complaint first to access its indexed fields
	complaint, found := k.GetComplaint(ctx, id)
	if !found {
		return
	}

	// Remove from main collection
	err := k.Complaints.Remove(ctx, id)
	if err != nil {
		panic(err)
	}

	// Remove from secondary indexes
	k.removeComplaintFromIndexes(ctx, complaint)
}

// removeComplaintFromIndexes removes complaint from all secondary indexes
func (k Keeper) removeComplaintFromIndexes(ctx context.Context, complaint types.Complaint) {
	// Remove from subnet index
	k.ComplaintsBySubnet.Remove(ctx, collections.Join(complaint.SubnetId, complaint.Id))

	// Remove from lease index
	k.ComplaintsByLease.Remove(ctx, collections.Join(complaint.LeaseId, complaint.Id))

	// Remove from lessee index
	k.ComplaintsByLessee.Remove(ctx, collections.Join(complaint.LesseeOrgId, complaint.Id))

	// Remove from lessor index
	k.ComplaintsByLessor.Remove(ctx, collections.Join(complaint.LessorOrgId, complaint.Id))
}

// GetAllComplaint returns all complaint
func (k Keeper) GetAllComplaint(ctx context.Context) (list []types.Complaint) {
	err := k.Complaints.Walk(ctx, nil, func(key uint64, value types.Complaint) (bool, error) {
		list = append(list, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return
}


