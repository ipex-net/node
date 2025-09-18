package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/nullify"
	"ipex/x/payments/types"
)

func TestTransactionQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	msgs := createNTransaction(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetTransactionRequest
		response *types.QueryGetTransactionResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetTransactionRequest{Id: msgs[0].Id},
			response: &types.QueryGetTransactionResponse{Transaction: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetTransactionRequest{Id: msgs[1].Id},
			response: &types.QueryGetTransactionResponse{Transaction: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetTransactionRequest{Id: 100000},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Transaction(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}
