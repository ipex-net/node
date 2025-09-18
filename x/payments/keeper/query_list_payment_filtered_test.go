package keeper_test

import (
	"testing"
	"time"

	keepertest "ipex/testutil/keeper"
	"ipex/testutil/sample"
	"ipex/x/payments/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

// TestListPaymentFiltered tests the universal filtered payment query
func TestListPaymentFiltered(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(100)

	// Test data setup
	orgId1 := uint64(1001)
	orgId2 := uint64(2001)
	orgId3 := uint64(3001)

	baseTime := ctx.BlockTime()
	fromTimestamp := uint64(baseTime.Unix())
	toTimestamp := uint64(baseTime.Add(time.Hour).Unix())

	// Create test transactions with different characteristics
	testTxs := []struct {
		txType    types.TransactionType
		moduleId  types.ModuleID
		entityId  uint64
		amount    int64
		fromOrgId uint64
		toOrgId   uint64
		timestamp time.Time
	}{
		{
			txType:    types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			moduleId:  types.ModuleID_MODULE_ID_LEASE,
			entityId:  10001,
			amount:    1000,
			fromOrgId: orgId1,
			toOrgId:   orgId2,
			timestamp: baseTime,
		},
		{
			txType:    types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			moduleId:  types.ModuleID_MODULE_ID_LEASE,
			entityId:  10002,
			amount:    5000,
			fromOrgId: orgId1,
			toOrgId:   orgId2,
			timestamp: baseTime.Add(time.Minute * 30),
		},
		{
			txType:    types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE,
			moduleId:  types.ModuleID_MODULE_ID_SUBNET_SALE,
			entityId:  20001,
			amount:    10000,
			fromOrgId: orgId2,
			toOrgId:   orgId3,
			timestamp: baseTime.Add(time.Minute * 45),
		},
		{
			txType:    types.TransactionType_TRANSACTION_TYPE_ASN_SALE,
			moduleId:  types.ModuleID_MODULE_ID_ASN_SALE,
			entityId:  30001,
			amount:    2000,
			fromOrgId: orgId1,
			toOrgId:   orgId3,
			timestamp: baseTime.Add(time.Hour * 2), // Outside time range
		},
	}

	var txIds []uint64
	for _, testTx := range testTxs {
		txId := keeper.RecordTransaction(
			ctx.WithBlockTime(testTx.timestamp),
			testTx.txType,
			testTx.moduleId,
			testTx.entityId,
			sdk.NewCoin("token", math.NewInt(testTx.amount)),
			testTx.fromOrgId,
			testTx.toOrgId,
			sample.AccAddress(),
			sample.AccAddress(),
		)
		txIds = append(txIds, txId)
	}

	t.Run("Basic Filter by Organization", func(t *testing.T) {
		// Query transactions for orgId1 (should find 3 transactions: as sender in tx1,tx2,tx4)
		req := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId1,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Transaction, 3, "should find 3 transactions for orgId1")

		// Verify all transactions involve orgId1
		for _, tx := range resp.Transaction {
			require.True(t, tx.FromOrgId == orgId1 || tx.ToOrgId == orgId1,
				"transaction should involve orgId1")
		}
	})

	t.Run("Filter by Organization Required", func(t *testing.T) {
		// Test that organization ID is required
		req := &types.QueryListPaymentFilteredRequest{
			OrgId: 0, // Missing org ID
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "organization ID is required")
	})

	t.Run("Filter by Transaction Type", func(t *testing.T) {
		// Query for LEASE_PAYMENT transactions for orgId1
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:           orgId1,
			TransactionType: uint64(types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT),
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 1, "should find 1 LEASE_PAYMENT transaction for orgId1")
		require.Equal(t, types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT, resp.Transaction[0].TransactionType)
	})

	t.Run("Filter by Module ID", func(t *testing.T) {
		// Query for LEASE module transactions for orgId1
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:    orgId1,
			ModuleId: uint64(types.ModuleID_MODULE_ID_LEASE),
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 2, "should find 2 LEASE module transactions for orgId1")

		for _, tx := range resp.Transaction {
			require.Equal(t, types.ModuleID_MODULE_ID_LEASE, tx.RelatedModuleId)
		}
	})

	t.Run("Filter by Timestamp Range", func(t *testing.T) {
		// Query for transactions within the first hour for orgId1
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:         orgId1,
			FromTimestamp: fromTimestamp,
			ToTimestamp:   toTimestamp,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 2, "should find 2 transactions within time range for orgId1")

		// Verify timestamps are within range
		for _, tx := range resp.Transaction {
			require.NotNil(t, tx.Timestamp)
			txTimestamp := uint64(tx.Timestamp.Unix())
			require.GreaterOrEqual(t, txTimestamp, fromTimestamp)
			require.LessOrEqual(t, txTimestamp, toTimestamp)
		}
	})

	t.Run("Filter by Amount Range", func(t *testing.T) {
		// Query for transactions with amount between 1000 and 6000 for orgId1
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:     orgId1,
			MinAmount: 1000,
			MaxAmount: 6000,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 3, "should find 3 transactions in amount range for orgId1")

		// Verify amounts are within range
		for _, tx := range resp.Transaction {
			amount := tx.Amount.Amount.Uint64()
			require.GreaterOrEqual(t, amount, uint64(1000))
			require.LessOrEqual(t, amount, uint64(6000))
		}
	})

	t.Run("Multiple Filters Combined", func(t *testing.T) {
		// Query for LEASE module transactions within time range with amount <= 2000 for orgId1
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:         orgId1,
			ModuleId:      uint64(types.ModuleID_MODULE_ID_LEASE),
			FromTimestamp: fromTimestamp,
			ToTimestamp:   toTimestamp,
			MaxAmount:     2000,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 1, "should find 1 transaction matching all filters")

		tx := resp.Transaction[0]
		require.Equal(t, types.ModuleID_MODULE_ID_LEASE, tx.RelatedModuleId)
		require.LessOrEqual(t, tx.Amount.Amount.Uint64(), uint64(2000))

		txTimestamp := uint64(tx.Timestamp.Unix())
		require.GreaterOrEqual(t, txTimestamp, fromTimestamp)
		require.LessOrEqual(t, txTimestamp, toTimestamp)
	})

	t.Run("No Results Filter", func(t *testing.T) {
		// Query for transactions that don't exist
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:           orgId1,
			TransactionType: uint64(types.TransactionType_TRANSACTION_TYPE_BID_SUBMITTED), // Not used in test data
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 0, "should find no transactions for unused transaction type")
		require.Equal(t, uint64(0), resp.Pagination.Total)
	})

	t.Run("Pagination", func(t *testing.T) {
		// First page
		req1 := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId1,
			Pagination: &query.PageRequest{
				Limit: 2,
			},
		}

		resp1, err := keeper.ListPaymentFiltered(ctx, req1)
		require.NoError(t, err)
		require.Len(t, resp1.Transaction, 2, "should get 2 transactions on first page")
		require.Equal(t, uint64(3), resp1.Pagination.Total, "should have correct total")
		require.NotEmpty(t, resp1.Pagination.NextKey, "should have next key for pagination")

		// Second page
		req2 := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId1,
			Pagination: &query.PageRequest{
				Limit: 2,
				Key:   resp1.Pagination.NextKey,
			},
		}

		resp2, err := keeper.ListPaymentFiltered(ctx, req2)
		require.NoError(t, err)
		require.Len(t, resp2.Transaction, 1, "should get 1 transaction on second page")
		require.Empty(t, resp2.Pagination.NextKey, "should not have next key on last page")

		// Verify no duplicate transactions between pages
		page1Ids := make(map[uint64]bool)
		for _, tx := range resp1.Transaction {
			page1Ids[tx.Id] = true
		}

		for _, tx := range resp2.Transaction {
			require.False(t, page1Ids[tx.Id], "should not have duplicate transactions across pages")
		}
	})

	t.Run("Filter by Entity ID", func(t *testing.T) {
		// Query for transactions with specific entity ID
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:    orgId1,
			EntityId: 10001, // First test transaction
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 1, "should find 1 transaction with entity ID 10001")
		require.Equal(t, uint64(10001), resp.Transaction[0].EntityId, "should have correct entity ID")
	})

	t.Run("Filter by FromOrgId", func(t *testing.T) {
		// Query for transactions where orgId1 is the sender
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:     orgId1,
			FromOrgId: orgId1,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 3, "should find 3 transactions where orgId1 is sender")

		for _, tx := range resp.Transaction {
			require.Equal(t, orgId1, tx.FromOrgId, "all transactions should have orgId1 as sender")
		}
	})

	t.Run("Filter by ToOrgId", func(t *testing.T) {
		// Query for transactions where orgId2 is the recipient
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:   orgId1,
			ToOrgId: orgId2,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 2, "should find 2 transactions where orgId2 is recipient")

		for _, tx := range resp.Transaction {
			require.Equal(t, orgId2, tx.ToOrgId, "all transactions should have orgId2 as recipient")
		}
	})

	t.Run("Multiple New Filters Combined", func(t *testing.T) {
		// Query with entity ID and fromOrgId filters
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:     orgId1,
			EntityId:  10002,
			FromOrgId: orgId1,
			ToOrgId:   orgId2,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 1, "should find 1 transaction matching all criteria")

		tx := resp.Transaction[0]
		require.Equal(t, uint64(10002), tx.EntityId, "should have correct entity ID")
		require.Equal(t, orgId1, tx.FromOrgId, "should have correct from org ID")
		require.Equal(t, orgId2, tx.ToOrgId, "should have correct to org ID")
	})

	t.Run("Edge Cases", func(t *testing.T) {
		// Query with all zero filters (except orgId) - should return all transactions for org
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:           orgId1,
			TransactionType: 0, // No filter
			ModuleId:        0, // No filter
			FromTimestamp:   0, // No filter
			ToTimestamp:     0, // No filter
			MinAmount:       0, // No filter
			MaxAmount:       0, // No filter
			EntityId:        0, // No filter
			FromOrgId:       0, // No filter
			ToOrgId:         0, // No filter
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 3, "should return all transactions for org when no filters applied")
	})

	t.Run("Organization as Recipient", func(t *testing.T) {
		// Query for orgId2 (appears as recipient in tx1,tx2 and sender in tx3)
		req := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId2,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 3, "should find 3 transactions for orgId2")

		// Verify orgId2 is involved in all transactions
		for _, tx := range resp.Transaction {
			require.True(t, tx.FromOrgId == orgId2 || tx.ToOrgId == orgId2,
				"transaction should involve orgId2")
		}
	})

	t.Run("Latest Entries First", func(t *testing.T) {
		// Query all transactions for orgId1 to check ordering
		req := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId1,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Transaction, 3, "should find 3 transactions for orgId1")

		// Verify transactions are ordered by ID descending (latest first)
		for i := 1; i < len(resp.Transaction); i++ {
			require.Greater(t, resp.Transaction[i-1].Id, resp.Transaction[i].Id,
				"transactions should be ordered by ID descending (latest first)")
		}

		// The first transaction should be the one with the highest ID (latest)
		require.Equal(t, txIds[3], resp.Transaction[0].Id, "first transaction should be the ASN sale (latest)")
	})

	t.Run("Invalid Request", func(t *testing.T) {
		resp, err := keeper.ListPaymentFiltered(ctx, nil)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "invalid request")
	})
}

// TestListPaymentFilteredPerformance tests performance characteristics
func TestListPaymentFilteredPerformance(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(100)

	// Create many transactions for performance testing
	orgId := uint64(9999)
	baseTime := ctx.BlockTime()

	// Create 100 transactions
	for i := 0; i < 100; i++ {
		keeper.RecordTransaction(
			ctx.WithBlockTime(baseTime.Add(time.Duration(i)*time.Minute)),
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			uint64(50000+i),
			sdk.NewCoin("token", math.NewInt(int64(1000+i*10))),
			orgId,
			uint64(8000+i),
			sample.AccAddress(),
			sample.AccAddress(),
		)
	}

	t.Run("Large Dataset Query", func(t *testing.T) {
		start := time.Now()

		req := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId,
			Pagination: &query.PageRequest{
				Limit: 50,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.Len(t, resp.Transaction, 50, "should retrieve 50 transactions")
		require.Equal(t, uint64(100), resp.Pagination.Total, "should have correct total")

		t.Logf("Queried 50 transactions from 100 total in %v", duration)
		require.Less(t, duration, time.Millisecond*100, "query should complete quickly")
	})

	t.Run("Filter Performance", func(t *testing.T) {
		start := time.Now()

		// Apply multiple filters
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:         orgId,
			ModuleId:      uint64(types.ModuleID_MODULE_ID_LEASE),
			FromTimestamp: uint64(baseTime.Add(time.Minute * 10).Unix()),
			ToTimestamp:   uint64(baseTime.Add(time.Minute * 50).Unix()),
			MinAmount:     1100,
			MaxAmount:     1500,
			Pagination: &query.PageRequest{
				Limit: 20,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Filtered query completed in %v, found %d transactions", duration, len(resp.Transaction))
		require.Less(t, duration, time.Millisecond*50, "filtered query should complete quickly")
	})

	t.Run("Reverse Order Performance", func(t *testing.T) {
		start := time.Now()

		// Test that reverse ordering doesn't impact performance
		req := &types.QueryListPaymentFilteredRequest{
			OrgId: orgId,
			Pagination: &query.PageRequest{
				Limit: 20,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		duration := time.Since(start)

		require.NoError(t, err)
		require.Len(t, resp.Transaction, 20, "should retrieve 20 transactions")

		// Verify transactions are in descending order (latest first)
		for i := 1; i < len(resp.Transaction); i++ {
			require.Greater(t, resp.Transaction[i-1].Id, resp.Transaction[i].Id,
				"transactions should be in descending order")
		}

		t.Logf("Reverse order query completed in %v", duration)
		require.Less(t, duration, time.Millisecond*50, "reverse order query should be fast")
	})

	t.Run("Extended Filter Performance", func(t *testing.T) {
		start := time.Now()

		// Test performance with all new filters applied
		req := &types.QueryListPaymentFilteredRequest{
			OrgId:         orgId,
			ModuleId:      uint64(types.ModuleID_MODULE_ID_LEASE),
			FromTimestamp: uint64(baseTime.Add(time.Minute * 10).Unix()),
			ToTimestamp:   uint64(baseTime.Add(time.Minute * 50).Unix()),
			MinAmount:     1100,
			MaxAmount:     1500,
			EntityId:      0,     // No specific entity filter
			FromOrgId:     orgId, // Only as sender
			ToOrgId:       0,     // Any recipient
			Pagination: &query.PageRequest{
				Limit: 20,
			},
		}

		resp, err := keeper.ListPaymentFiltered(ctx, req)
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Extended filter query completed in %v, found %d transactions", duration, len(resp.Transaction))
		require.Less(t, duration, time.Millisecond*50, "extended filter query should be fast")

		// Verify all returned transactions match the filters
		for _, tx := range resp.Transaction {
			require.Equal(t, orgId, tx.FromOrgId, "should match fromOrgId filter")
			require.Equal(t, types.ModuleID_MODULE_ID_LEASE, tx.RelatedModuleId, "should match module filter")
		}
	})
}
