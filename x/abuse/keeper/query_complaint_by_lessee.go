package keeper

import (
	"context"

	"ipex/x/abuse/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ComplaintByLessee(ctx context.Context, req *types.QueryComplaintByLesseeRequest) (*types.QueryComplaintByLesseeResponse, error) {
	if req == nil || req.LesseeOrgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var complaints []types.Complaint

	// Iterate over all complaints with this lessee org ID using prefix iteration
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.LesseeOrgId)
	err := k.ComplaintsByLessee.Walk(ctx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// key.K2() is the complaint ID
		complaintId := key.K2()
		complaint, found := k.GetComplaint(ctx, complaintId)
		if found {
			complaints = append(complaints, complaint)
		}
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryComplaintByLesseeResponse{
		Complaint:  complaints,
		Pagination: &query.PageResponse{Total: uint64(len(complaints))},
	}, nil
}
