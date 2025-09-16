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

// TestTransactionLifecycleIntegration tests the complete transaction lifecycle
func TestTransactionLifecycleIntegration(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(100)

	// Test data
	baseTime := ctx.BlockTime()
	fromAddress := sample.AccAddress()
	toAddress := sample.AccAddress()
	amount := sdk.NewCoin("token", math.NewInt(1000))

	t.Run("Record Lease Payment Transaction", func(t *testing.T) {
		// Record a lease payment transaction
		txId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			12345, // entityId (lease contract ID)
			amount,
			1001, // fromOrgId (lessee)
			2001, // toOrgId (lessor)
			fromAddress,
			toAddress,
		)

		require.Greater(t, txId, uint64(0), "should create transaction with valid ID")

		// Verify transaction was recorded correctly
		retrievedTx, found := keeper.GetTransaction(ctx, txId)
		require.True(t, found, "should find recorded transaction")
		require.Equal(t, txId, retrievedTx.Id, "should have correct ID")
		require.Equal(t, types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT, retrievedTx.TransactionType, "should have correct transaction type")
		require.Equal(t, types.ModuleID_MODULE_ID_LEASE, retrievedTx.RelatedModuleId, "should have correct module ID")
		require.Equal(t, uint64(12345), retrievedTx.EntityId, "should have correct entity ID")
		require.Equal(t, amount, retrievedTx.Amount, "should have correct amount")
		require.Equal(t, uint64(1001), retrievedTx.FromOrgId, "should have correct from org ID")
		require.Equal(t, uint64(2001), retrievedTx.ToOrgId, "should have correct to org ID")
		require.Equal(t, fromAddress, retrievedTx.FromAddress, "should have correct from address")
		require.Equal(t, toAddress, retrievedTx.ToAddress, "should have correct to address")
		require.Equal(t, int64(100), retrievedTx.BlockHeight, "should have correct block height")
		require.NotNil(t, retrievedTx.Timestamp, "should have timestamp")
		require.Equal(t, baseTime, *retrievedTx.Timestamp, "should have correct timestamp")
	})

	t.Run("Record Safety Deposit Transaction", func(t *testing.T) {
		// Record a safety deposit collection
		depositAmount := sdk.NewCoin("token", math.NewInt(5000))
		txId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			types.ModuleID_MODULE_ID_LEASE,
			12346, // different entity
			depositAmount,
			1002, // different fromOrg
			2001, // same toOrg
			fromAddress,
			toAddress,
		)

		require.Greater(t, txId, uint64(0), "should create deposit transaction")

		// Verify deposit transaction
		retrievedTx, found := keeper.GetTransaction(ctx, txId)
		require.True(t, found, "should find deposit transaction")
		require.Equal(t, types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT, retrievedTx.TransactionType, "should be deposit collection")
		require.Equal(t, depositAmount, retrievedTx.Amount, "should have correct deposit amount")
	})

	t.Run("Record Multiple Transaction Types", func(t *testing.T) {
		// Test all transaction types
		transactionTypes := []types.TransactionType{
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COMPENSATION,
			types.TransactionType_TRANSACTION_TYPE_REFUND,
			types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE,
			types.TransactionType_TRANSACTION_TYPE_ASN_SALE,
		}

		moduleIds := []types.ModuleID{
			types.ModuleID_MODULE_ID_LEASE,
			types.ModuleID_MODULE_ID_LEASE,
			types.ModuleID_MODULE_ID_LEASE,
			types.ModuleID_MODULE_ID_SUBNET_SALE,
			types.ModuleID_MODULE_ID_ASN_SALE,
		}

		var txIds []uint64
		for i, txType := range transactionTypes {
			txId := keeper.RecordTransaction(
				ctx,
				txType,
				moduleIds[i],
				uint64(50000+i), // unique entity IDs
				sdk.NewCoin("token", math.NewInt(int64(1000*(i+1)))),
				uint64(3000+i), // unique from orgs
				uint64(4000+i), // unique to orgs
				sample.AccAddress(),
				sample.AccAddress(),
			)

			txIds = append(txIds, txId)
			require.Greater(t, txId, uint64(0), "should create transaction for type %s", txType.String())
		}

		// Verify all transactions were created with correct types
		for i, txId := range txIds {
			tx, found := keeper.GetTransaction(ctx, txId)
			require.True(t, found, "should find transaction %d", i)
			require.Equal(t, transactionTypes[i], tx.TransactionType, "should have correct transaction type %d", i)
			require.Equal(t, moduleIds[i], tx.RelatedModuleId, "should have correct module ID %d", i)
		}
	})

	t.Run("Transaction Update and State Management", func(t *testing.T) {
		// Create a transaction
		originalTx := types.Transaction{
			TransactionType: types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			RelatedModuleId: types.ModuleID_MODULE_ID_LEASE,
			EntityId:        99999,
			Amount:          sdk.NewCoin("token", math.NewInt(2000)),
			FromOrgId:       5001,
			ToOrgId:         6001,
			FromAddress:     sample.AccAddress(),
			ToAddress:       sample.AccAddress(),
			BlockHeight:     ctx.BlockHeight(),
			Timestamp:       &baseTime,
		}

		txId := keeper.AppendTransaction(ctx, originalTx)

		// Update transaction (simulating a correction or additional data)
		updatedTx, found := keeper.GetTransaction(ctx, txId)
		require.True(t, found, "should find transaction for update")

		// Update some fields (note: in practice, transactions might be immutable)
		updatedTx.BlockHeight = ctx.BlockHeight() + 1
		keeper.SetTransaction(ctx, updatedTx)

		// Verify update
		finalTx, found := keeper.GetTransaction(ctx, txId)
		require.True(t, found, "should find updated transaction")
		require.Equal(t, ctx.BlockHeight()+1, finalTx.BlockHeight, "should have updated block height")
		require.Equal(t, originalTx.Amount, finalTx.Amount, "should preserve original amount")
	})

	t.Run("Transaction Removal", func(t *testing.T) {
		// Create a transaction for removal testing
		tx := types.Transaction{
			TransactionType: types.TransactionType_TRANSACTION_TYPE_REFUND,
			RelatedModuleId: types.ModuleID_MODULE_ID_LEASE,
			EntityId:        88888,
			Amount:          sdk.NewCoin("token", math.NewInt(500)),
			FromOrgId:       7001,
			ToOrgId:         8001,
			FromAddress:     sample.AccAddress(),
			ToAddress:       sample.AccAddress(),
			BlockHeight:     ctx.BlockHeight(),
			Timestamp:       &baseTime,
		}

		txId := keeper.AppendTransaction(ctx, tx)

		// Verify transaction exists
		_, found := keeper.GetTransaction(ctx, txId)
		require.True(t, found, "should find transaction before removal")

		// Remove transaction
		keeper.RemoveTransaction(ctx, txId)

		// Verify transaction was removed
		_, found = keeper.GetTransaction(ctx, txId)
		require.False(t, found, "should not find transaction after removal")

		// Verify transaction is not in the list
		allTransactions := keeper.GetAllTransaction(ctx)
		for _, tx := range allTransactions {
			require.NotEqual(t, txId, tx.Id, "removed transaction should not appear in list")
		}
	})
}

// TestTransactionIndexingIntegration tests the indexing system for efficient queries
func TestTransactionIndexingIntegration(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(200)

	t.Run("Index by Entity", func(t *testing.T) {
		entityId1 := uint64(10001)
		entityId2 := uint64(10002)

		// Create transactions for different entities
		tx1Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			entityId1,
			sdk.NewCoin("token", math.NewInt(1000)),
			1001,
			2001,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		tx2Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			types.ModuleID_MODULE_ID_LEASE,
			entityId1, // Same entity
			sdk.NewCoin("token", math.NewInt(3000)),
			1001,
			2001,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		tx3Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE,
			types.ModuleID_MODULE_ID_SUBNET_SALE,
			entityId2, // Different entity
			sdk.NewCoin("token", math.NewInt(10000)),
			3001,
			4001,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Test ListContractPayment query (queries by entity)
		req := &types.QueryListContractPaymentRequest{
			Id: entityId1,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp, err := keeper.ListContractPayment(ctx, req)
		require.NoError(t, err, "should successfully query transactions by entity")
		require.NotNil(t, resp, "should have response")
		require.Len(t, resp.Transaction, 2, "should find 2 transactions for entity1")

		// Verify the transactions belong to the correct entity
		foundTx1 := false
		foundTx2 := false
		for _, tx := range resp.Transaction {
			require.Equal(t, entityId1, tx.EntityId, "all transactions should belong to entity1")
			if tx.Id == tx1Id {
				foundTx1 = true
			}
			if tx.Id == tx2Id {
				foundTx2 = true
			}
		}
		require.True(t, foundTx1, "should find first transaction")
		require.True(t, foundTx2, "should find second transaction")

		// Test with entity2 (should find only one transaction)
		req2 := &types.QueryListContractPaymentRequest{
			Id: entityId2,
			Pagination: &query.PageRequest{
				Limit: 10,
			},
		}

		resp2, err := keeper.ListContractPayment(ctx, req2)
		require.NoError(t, err, "should successfully query transactions by entity2")
		require.Len(t, resp2.Transaction, 1, "should find 1 transaction for entity2")
		require.Equal(t, tx3Id, resp2.Transaction[0].Id, "should find correct transaction for entity2")
	})

	t.Run("Index by Organization", func(t *testing.T) {
		orgId1 := uint64(20001)
		orgId2 := uint64(20002)
		orgId3 := uint64(20003)

		// Create transactions involving different organizations
		// Transaction 1: org1 -> org2
		tx1Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			30001,
			sdk.NewCoin("token", math.NewInt(1500)),
			orgId1, // from
			orgId2, // to
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Transaction 2: org2 -> org3
		tx2Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND,
			types.ModuleID_MODULE_ID_LEASE,
			30002,
			sdk.NewCoin("token", math.NewInt(2500)),
			orgId2, // from
			orgId3, // to
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Transaction 3: org3 -> org1
		tx3Id := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_REFUND,
			types.ModuleID_MODULE_ID_LEASE,
			30003,
			sdk.NewCoin("token", math.NewInt(750)),
			orgId3, // from
			orgId1, // to
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Test filtering transactions for org1 (involved in tx1 as from, tx3 as to)
		allTransactionsResp, err := keeper.TransactionAll(ctx, &types.QueryAllTransactionRequest{})
		require.NoError(t, err, "should get all transactions")

		var org1Transactions []types.Transaction
		for _, tx := range allTransactionsResp.Transaction {
			if tx.FromOrgId == orgId1 || tx.ToOrgId == orgId1 {
				org1Transactions = append(org1Transactions, tx)
			}
		}
		require.Len(t, org1Transactions, 2, "org1 should be involved in 2 transactions")

		var foundTx1, foundTx3 bool
		for _, tx := range org1Transactions {
			if tx.Id == tx1Id {
				foundTx1 = true
				require.Equal(t, orgId1, tx.FromOrgId, "org1 should be sender in tx1")
			}
			if tx.Id == tx3Id {
				foundTx3 = true
				require.Equal(t, orgId1, tx.ToOrgId, "org1 should be recipient in tx3")
			}
		}
		require.True(t, foundTx1, "should find tx1 for org1")
		require.True(t, foundTx3, "should find tx3 for org1")

		// Test filtering transactions for org2 (involved in tx1 as to, tx2 as from)
		var org2Transactions []types.Transaction
		for _, tx := range allTransactionsResp.Transaction {
			if tx.FromOrgId == orgId2 || tx.ToOrgId == orgId2 {
				org2Transactions = append(org2Transactions, tx)
			}
		}
		require.Len(t, org2Transactions, 2, "org2 should be involved in 2 transactions")

		var foundTx1ForOrg2, foundTx2 bool
		for _, tx := range org2Transactions {
			if tx.Id == tx1Id {
				foundTx1ForOrg2 = true
				require.Equal(t, orgId2, tx.ToOrgId, "org2 should be recipient in tx1")
			}
			if tx.Id == tx2Id {
				foundTx2 = true
				require.Equal(t, orgId2, tx.FromOrgId, "org2 should be sender in tx2")
			}
		}
		require.True(t, foundTx1ForOrg2, "should find tx1 for org2")
		require.True(t, foundTx2, "should find tx2 for org2")
	})

	t.Run("Pagination by Entity", func(t *testing.T) {
		entityId := uint64(40001)

		// Create multiple transactions for the same entity
		var expectedTxIds []uint64
		for i := 0; i < 5; i++ {
			txId := keeper.RecordTransaction(
				ctx,
				types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
				types.ModuleID_MODULE_ID_LEASE,
				entityId,
				sdk.NewCoin("token", math.NewInt(int64(1000+i*100))),
				uint64(50001+i),
				uint64(60001+i),
				sample.AccAddress(),
				sample.AccAddress(),
			)
			expectedTxIds = append(expectedTxIds, txId)
		}

		// Test pagination - get first 2 transactions
		req := &types.QueryListContractPaymentRequest{
			Id: entityId,
			Pagination: &query.PageRequest{
				Limit:  2,
				Offset: 0,
			},
		}

		resp, err := keeper.ListContractPayment(ctx, req)
		require.NoError(t, err, "should successfully paginate")
		require.Len(t, resp.Transaction, 2, "should get 2 transactions")
		require.Equal(t, uint64(5), resp.Pagination.Total, "should have correct total count")

		// Test pagination - get next 2 transactions
		req2 := &types.QueryListContractPaymentRequest{
			Id: entityId,
			Pagination: &query.PageRequest{
				Limit:  2,
				Offset: 2,
			},
		}

		resp2, err := keeper.ListContractPayment(ctx, req2)
		require.NoError(t, err, "should successfully paginate second page")
		require.Len(t, resp2.Transaction, 2, "should get 2 transactions on second page")

		// Test pagination - get remaining transaction
		req3 := &types.QueryListContractPaymentRequest{
			Id: entityId,
			Pagination: &query.PageRequest{
				Limit:  2,
				Offset: 4,
			},
		}

		resp3, err := keeper.ListContractPayment(ctx, req3)
		require.NoError(t, err, "should successfully paginate third page")
		require.Len(t, resp3.Transaction, 1, "should get 1 transaction on third page")

		// Verify all transactions are accounted for
		allRetrievedIds := make(map[uint64]bool)
		for _, tx := range resp.Transaction {
			allRetrievedIds[tx.Id] = true
		}
		for _, tx := range resp2.Transaction {
			allRetrievedIds[tx.Id] = true
		}
		for _, tx := range resp3.Transaction {
			allRetrievedIds[tx.Id] = true
		}

		require.Len(t, allRetrievedIds, 5, "should retrieve all 5 transactions across pages")
		for _, expectedId := range expectedTxIds {
			require.True(t, allRetrievedIds[expectedId], "should find expected transaction ID %d", expectedId)
		}
	})

	t.Run("Pagination by Organization", func(t *testing.T) {
		orgId := uint64(9999001)

		// Create multiple transactions involving the same organization
		var expectedTxIds []uint64
		for i := 0; i < 7; i++ {
			var fromOrg, toOrg uint64
			if i%2 == 0 {
				fromOrg = orgId // org as sender
				toOrg = uint64(9998001 + i)
			} else {
				fromOrg = uint64(9998001 + i)
				toOrg = orgId // org as recipient
			}

			txId := keeper.RecordTransaction(
				ctx,
				types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
				types.ModuleID_MODULE_ID_LEASE,
				uint64(9997001+i),
				sdk.NewCoin("token", math.NewInt(int64(500+i*50))),
				fromOrg,
				toOrg,
				sample.AccAddress(),
				sample.AccAddress(),
			)
			expectedTxIds = append(expectedTxIds, txId)
		}

		// First get all transactions and filter by organization to get total count
		allTransactionsResp, err := keeper.TransactionAll(ctx, &types.QueryAllTransactionRequest{})
		require.NoError(t, err, "should get all transactions")

		var totalTxs []types.Transaction
		for _, tx := range allTransactionsResp.Transaction {
			if tx.FromOrgId == orgId || tx.ToOrgId == orgId {
				totalTxs = append(totalTxs, tx)
			}
		}
		require.Len(t, totalTxs, 7, "should find all 7 transactions for org")

		// Test paginated query for organization (first page) by filtering all transactions
		allTransactionsPage1, err := keeper.TransactionAll(ctx, &types.QueryAllTransactionRequest{})
		require.NoError(t, err, "should get transactions for pagination")

		var allOrgTransactions []types.Transaction
		for _, tx := range allTransactionsPage1.Transaction {
			if tx.FromOrgId == orgId || tx.ToOrgId == orgId {
				allOrgTransactions = append(allOrgTransactions, tx)
			}
		}

		// Simulate first page by taking first 3 transactions
		var transactions []types.Transaction
		for i := 0; i < len(allOrgTransactions) && i < 3; i++ {
			transactions = append(transactions, allOrgTransactions[i])
		}
		require.LessOrEqual(t, len(transactions), 3, "should get at most 3 transactions on first page")
		if len(allOrgTransactions) >= 3 {
			require.Len(t, transactions, 3, "should get exactly 3 transactions if enough available")
		}

		// Test second page by taking transactions 3-5 from filtered results
		var transactions2 []types.Transaction
		for i := 3; i < len(allOrgTransactions) && i < 6; i++ {
			transactions2 = append(transactions2, allOrgTransactions[i])
		}
		if len(allOrgTransactions) >= 6 {
			require.Len(t, transactions2, 3, "should get exactly 3 transactions on second page if enough available")
		} else {
			require.LessOrEqual(t, len(transactions2), 3, "should get remaining transactions on second page")
		}

		// Test third page by taking transactions 6+ from filtered results
		var transactions3 []types.Transaction
		for i := 6; i < len(allOrgTransactions) && i < 9; i++ {
			transactions3 = append(transactions3, allOrgTransactions[i])
		}
		if len(allOrgTransactions) == 7 {
			require.Len(t, transactions3, 1, "should get 1 transaction on third page")
		} else {
			require.LessOrEqual(t, len(transactions3), 3, "should get remaining transactions on third page")
		}

		// Verify we got all expected transactions across pages
		allPageTransactions := append(transactions, transactions2...)
		allPageTransactions = append(allPageTransactions, transactions3...)
		require.Equal(t, len(allOrgTransactions), len(allPageTransactions), "should get all org transactions across pages")
	})
}
