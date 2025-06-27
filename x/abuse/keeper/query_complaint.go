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

	// Use collections to iterate through all complaints
	err := k.Complaints.Walk(ctx, nil, func(key uint64, value types.Complaint) (bool, error) {
		complaints = append(complaints, value)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllComplaintResponse{
		Complaint:  complaints,
		Pagination: &query.PageResponse{Total: uint64(len(complaints))},
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
