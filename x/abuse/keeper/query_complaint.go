package keeper

import (
	"context"

	"ipex/x/abuse/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ComplaintAll(ctx context.Context, req *types.QueryAllComplaintRequest) (*types.QueryAllComplaintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var complaints []types.Complaint
	var count uint64
	var skipped uint64

	// Set up pagination parameters
	limit := uint64(100) // Default limit
	offset := uint64(0)
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
	}

	// Use collections to iterate through complaints with pagination
	err := k.Complaints.Walk(ctx, nil, func(key uint64, value types.Complaint) (bool, error) {
		count++

		// Skip entries until we reach the offset
		if skipped < offset {
			skipped++
			return false, nil
		}

		// Stop if we've collected enough items
		if uint64(len(complaints)) >= limit {
			return true, nil
		}

		complaints = append(complaints, value)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Calculate next key for pagination
	var nextKey []byte
	if uint64(len(complaints)) == limit && count > offset+limit {
		// There are more items available
		nextKey = []byte("has_more")
	}

	return &types.QueryAllComplaintResponse{
		Complaint: complaints,
		Pagination: &query.PageResponse{
			NextKey: nextKey,
			Total:   count,
		},
	}, nil
}

func (k Keeper) Complaint(ctx context.Context, req *types.QueryGetComplaintRequest) (*types.QueryGetComplaintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetComplaint(
		ctx,
		req.Id,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetComplaintResponse{Complaint: val}, nil
}
