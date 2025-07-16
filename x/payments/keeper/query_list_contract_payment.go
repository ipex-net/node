package keeper

import (
	"context"

	"ipex/x/payments/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ListContractPayment(goCtx context.Context, req *types.QueryListContractPaymentRequest) (*types.QueryListContractPaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var transactions []types.Transaction

	// Use entity index to efficiently get transactions for this entity
	rng := collections.NewPrefixedPairRange[uint64, uint64](req.Id)
	err := k.TransactionsByEntity.Walk(goCtx, rng, func(key collections.Pair[uint64, uint64], value bool) (bool, error) {
		// Get the transaction ID from the key
		transactionId := key.K2()

		// Fetch the full transaction
		transaction, found := k.GetTransaction(goCtx, transactionId)
		if !found {
			// Transaction might have been deleted, skip
			return false, nil
		}

		transactions = append(transactions, transaction)
		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Apply pagination to the collected transactions
	offset := int(req.Pagination.GetOffset())
	limit := int(req.Pagination.GetLimit())
	total := uint64(len(transactions))

	// Apply offset
	if offset > len(transactions) {
		offset = len(transactions)
	}
	transactions = transactions[offset:]

	// Apply limit
	if limit > 0 && limit < len(transactions) {
		transactions = transactions[:limit]
	}

	pageResponse := &query.PageResponse{
		Total: total,
	}

	// Set next key if there are more results
	if req.Pagination.GetCountTotal() {
		pageResponse.Total = total
	}

	return &types.QueryListContractPaymentResponse{
		Transaction: transactions,
		Pagination:  pageResponse,
	}, nil
}
