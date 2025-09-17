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

// TestPaymentsCrossModuleIntegration tests integration with other modules
func TestPaymentsCrossModuleIntegration(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(300)

	baseTime := ctx.BlockTime()

	t.Run("Lease Module Integration", func(t *testing.T) {
		// Simulate a complete lease payment workflow
		leaseContractId := uint64(50001)
		lessorOrgId := uint64(60001)
		lesseeOrgId := uint64(70001)

		lessorAddress := sample.AccAddress()
		lesseeAddress := sample.AccAddress()

		// 1. Initial lease payment
		leasePaymentAmount := sdk.NewCoin("utoken", math.NewInt(10000))
		paymentTxId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			leaseContractId,
			leasePaymentAmount,
			lesseeOrgId, // lessee pays
			lessorOrgId, // lessor receives
			lesseeAddress,
			lessorAddress,
		)

		// 2. Safety deposit collection
		depositAmount := sdk.NewCoin("utoken", math.NewInt(25000))
		depositTxId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			types.ModuleID_MODULE_ID_LEASE,
			leaseContractId,
			depositAmount,
			lesseeOrgId, // lessee pays deposit
			lessorOrgId, // lessor receives deposit
			lesseeAddress,
			lessorAddress,
		)

		// 3. Later: Safety deposit refund (when lease ends normally)
		ctx = ctx.WithBlockTime(baseTime.Add(time.Hour * 24 * 30)).WithBlockHeight(400) // 30 days later

		refundTxId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND,
			types.ModuleID_MODULE_ID_LEASE,
			leaseContractId,
			depositAmount, // full refund
			lessorOrgId,   // lessor refunds
			lesseeOrgId,   // lessee receives
			lessorAddress,
			lesseeAddress,
		)

		// Verify all transactions were recorded for the lease contract
		contractTransactions, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId:    lesseeOrgId, // Query by lessee organization
			EntityId: leaseContractId,
			Pagination: &query.PageRequest{Limit: 10},
		})
		require.NoError(t, err, "should get contract transactions")
		require.Len(t, contractTransactions.Transaction, 3, "should have 3 transactions for lease contract")

		// Verify transaction types and amounts
		var foundPayment, foundDeposit, foundRefund bool
		for _, tx := range contractTransactions.Transaction {
			require.Equal(t, leaseContractId, tx.EntityId, "all transactions should belong to lease contract")
			require.Equal(t, types.ModuleID_MODULE_ID_LEASE, tx.RelatedModuleId, "all should be lease module transactions")

			switch tx.Id {
			case paymentTxId:
				foundPayment = true
				require.Equal(t, types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT, tx.TransactionType)
				require.Equal(t, leasePaymentAmount, tx.Amount)
			case depositTxId:
				foundDeposit = true
				require.Equal(t, types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT, tx.TransactionType)
				require.Equal(t, depositAmount, tx.Amount)
			case refundTxId:
				foundRefund = true
				require.Equal(t, types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND, tx.TransactionType)
				require.Equal(t, depositAmount, tx.Amount)
			}
		}
		require.True(t, foundPayment, "should find lease payment transaction")
		require.True(t, foundDeposit, "should find deposit transaction")
		require.True(t, foundRefund, "should find refund transaction")

		// Verify organization transaction history
		lessorTransactionsResp, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: lessorOrgId,
			Pagination: &query.PageRequest{Limit: 10},
		})
		require.NoError(t, err, "should get lessor transactions")
		require.GreaterOrEqual(t, len(lessorTransactionsResp.Transaction), 3, "lessor should be involved in at least 3 transactions")

		lesseeTransactionsResp, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: lesseeOrgId,
			Pagination: &query.PageRequest{Limit: 10},
		})
		require.NoError(t, err, "should get lessee transactions")
		require.GreaterOrEqual(t, len(lesseeTransactionsResp.Transaction), 3, "lessee should be involved in at least 3 transactions")
	})

	t.Run("Subnet Sale Integration", func(t *testing.T) {
		// Simulate subnet sale transaction
		subnetId := uint64(80001)
		sellerOrgId := uint64(90001)
		buyerOrgId := uint64(90002)

		sellerAddress := sample.AccAddress()
		buyerAddress := sample.AccAddress()
		saleAmount := sdk.NewCoin("utoken", math.NewInt(1000000)) // 1M tokens

		saleTxId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE,
			types.ModuleID_MODULE_ID_SUBNET_SALE,
			subnetId,
			saleAmount,
			buyerOrgId,  // buyer pays
			sellerOrgId, // seller receives
			buyerAddress,
			sellerAddress,
		)

		// Verify subnet sale transaction
		saleTx, found := keeper.GetTransaction(ctx, saleTxId)
		require.True(t, found, "should find subnet sale transaction")
		require.Equal(t, types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE, saleTx.TransactionType)
		require.Equal(t, types.ModuleID_MODULE_ID_SUBNET_SALE, saleTx.RelatedModuleId)
		require.Equal(t, subnetId, saleTx.EntityId)
		require.Equal(t, saleAmount, saleTx.Amount)

		// Verify organizations involved in sale
		allTransactionsResp2, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: sellerOrgId, // Query by seller organization to get transactions they're involved in
			Pagination: &query.PageRequest{Limit: 1000},
		})
		require.NoError(t, err, "should get seller transactions")
		// Also get buyer transactions
		buyerTransactionsResp, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: buyerOrgId,
			Pagination: &query.PageRequest{Limit: 1000},
		})
		require.NoError(t, err, "should get buyer transactions")

		// sellerTxs are already filtered by seller org
		sellerTxs := allTransactionsResp2.Transaction

		// buyerTxs are already filtered by buyer org
		buyerTxs := buyerTransactionsResp.Transaction

		// Find the sale transaction in both org histories
		var foundInSeller, foundInBuyer bool
		for _, tx := range sellerTxs {
			if tx.Id == saleTxId {
				foundInSeller = true
				require.Equal(t, sellerOrgId, tx.ToOrgId, "seller should be recipient")
			}
		}
		for _, tx := range buyerTxs {
			if tx.Id == saleTxId {
				foundInBuyer = true
				require.Equal(t, buyerOrgId, tx.FromOrgId, "buyer should be sender")
			}
		}
		require.True(t, foundInSeller, "sale should appear in seller's transaction history")
		require.True(t, foundInBuyer, "sale should appear in buyer's transaction history")
	})

	t.Run("ASN Sale Integration", func(t *testing.T) {
		// Simulate ASN sale transaction
		asnId := uint64(65001)
		asnSellerOrgId := uint64(95001)
		asnBuyerOrgId := uint64(95002)

		asnSellerAddress := sample.AccAddress()
		asnBuyerAddress := sample.AccAddress()
		asnSaleAmount := sdk.NewCoin("utoken", math.NewInt(500000)) // 500K tokens

		asnSaleTxId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_ASN_SALE,
			types.ModuleID_MODULE_ID_ASN_SALE,
			asnId,
			asnSaleAmount,
			asnBuyerOrgId,  // buyer pays
			asnSellerOrgId, // seller receives
			asnBuyerAddress,
			asnSellerAddress,
		)

		// Verify ASN sale transaction
		asnSaleTx, found := keeper.GetTransaction(ctx, asnSaleTxId)
		require.True(t, found, "should find ASN sale transaction")
		require.Equal(t, types.TransactionType_TRANSACTION_TYPE_ASN_SALE, asnSaleTx.TransactionType)
		require.Equal(t, types.ModuleID_MODULE_ID_ASN_SALE, asnSaleTx.RelatedModuleId)
		require.Equal(t, asnId, asnSaleTx.EntityId)
		require.Equal(t, asnSaleAmount, asnSaleTx.Amount)
	})

	t.Run("Complex Multi-Module Transaction History", func(t *testing.T) {
		// Create a complex scenario with multiple modules and transaction types
		orgId := uint64(100001)

		// This org is involved in multiple types of transactions across modules
		transactions := []struct {
			txType   types.TransactionType
			moduleId types.ModuleID
			entityId uint64
			amount   sdk.Coin
			fromOrg  uint64
			toOrg    uint64
		}{
			// Lease transactions
			{types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT, types.ModuleID_MODULE_ID_LEASE, 101, sdk.NewCoin("utoken", math.NewInt(5000)), orgId, 200001},
			{types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT, types.ModuleID_MODULE_ID_LEASE, 102, sdk.NewCoin("utoken", math.NewInt(15000)), orgId, 200002},
			{types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND, types.ModuleID_MODULE_ID_LEASE, 103, sdk.NewCoin("utoken", math.NewInt(12000)), 200003, orgId},
			// Sales transactions
			{types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE, types.ModuleID_MODULE_ID_SUBNET_SALE, 201, sdk.NewCoin("utoken", math.NewInt(800000)), 300001, orgId},
			{types.TransactionType_TRANSACTION_TYPE_ASN_SALE, types.ModuleID_MODULE_ID_ASN_SALE, 301, sdk.NewCoin("utoken", math.NewInt(400000)), orgId, 300002},
		}

		var txIds []uint64
		for _, txData := range transactions {
			txId := keeper.RecordTransaction(
				ctx,
				txData.txType,
				txData.moduleId,
				txData.entityId,
				txData.amount,
				txData.fromOrg,
				txData.toOrg,
				sample.AccAddress(),
				sample.AccAddress(),
			)
			txIds = append(txIds, txId)
		}

		// Get all transactions for this organization
		allTransactionsResp3, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: orgId,
			Pagination: &query.PageRequest{Limit: 1000},
		})
		require.NoError(t, err, "should get org transactions")

		orgTransactions := allTransactionsResp3.Transaction
		require.GreaterOrEqual(t, len(orgTransactions), 5, "should have at least 5 transactions")

		// Verify transaction types and modules are represented
		moduleCount := make(map[types.ModuleID]int)
		typeCount := make(map[types.TransactionType]int)

		for _, tx := range orgTransactions {
			// Only count transactions we created for this test
			found := false
			for _, expectedTxId := range txIds {
				if tx.Id == expectedTxId {
					found = true
					break
				}
			}
			if found {
				moduleCount[tx.RelatedModuleId]++
				typeCount[tx.TransactionType]++
			}
		}

		// Verify we have transactions from multiple modules
		require.Greater(t, len(moduleCount), 1, "should have transactions from multiple modules")
		require.Contains(t, moduleCount, types.ModuleID_MODULE_ID_LEASE, "should have lease transactions")
		require.Contains(t, moduleCount, types.ModuleID_MODULE_ID_SUBNET_SALE, "should have subnet sale transactions")
		require.Contains(t, moduleCount, types.ModuleID_MODULE_ID_ASN_SALE, "should have ASN sale transactions")

		// Verify we have multiple transaction types
		require.Greater(t, len(typeCount), 3, "should have multiple transaction types")
	})
}

// TestPaymentsValidationIntegration tests validation logic and edge cases
func TestPaymentsValidationIntegration(t *testing.T) {
	keeper, ctx := keepertest.PaymentsKeeper(t)
	ctx = ctx.WithBlockTime(time.Now()).WithBlockHeight(500)

	t.Run("Transaction Type Validation", func(t *testing.T) {
		// Test all valid transaction types
		validTypes := []types.TransactionType{
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_REFUND,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COMPENSATION,
			types.TransactionType_TRANSACTION_TYPE_REFUND,
			types.TransactionType_TRANSACTION_TYPE_SUBNET_SALE,
			types.TransactionType_TRANSACTION_TYPE_ASN_SALE,
		}

		for i, txType := range validTypes {
			txId := keeper.RecordTransaction(
				ctx,
				txType,
				types.ModuleID_MODULE_ID_LEASE,
				uint64(600001+i),
				sdk.NewCoin("utoken", math.NewInt(1000)),
				uint64(700001+i),
				uint64(800001+i),
				sample.AccAddress(),
				sample.AccAddress(),
			)

			require.Greater(t, txId, uint64(0), "should create transaction for type %s", txType.String())

			// Verify transaction was stored with correct type
			tx, found := keeper.GetTransaction(ctx, txId)
			require.True(t, found, "should find transaction")
			require.Equal(t, txType, tx.TransactionType, "should have correct transaction type")
		}
	})

	t.Run("Module ID Validation", func(t *testing.T) {
		// Test all valid module IDs
		validModules := []types.ModuleID{
			types.ModuleID_MODULE_ID_LEASE,
			types.ModuleID_MODULE_ID_SUBNET_SALE,
			types.ModuleID_MODULE_ID_ASN_SALE,
		}

		for i, moduleId := range validModules {
			txId := keeper.RecordTransaction(
				ctx,
				types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
				moduleId,
				uint64(900001+i),
				sdk.NewCoin("utoken", math.NewInt(1000)),
				uint64(1000001+i),
				uint64(1100001+i),
				sample.AccAddress(),
				sample.AccAddress(),
			)

			require.Greater(t, txId, uint64(0), "should create transaction for module %s", moduleId.String())

			// Verify transaction was stored with correct module
			tx, found := keeper.GetTransaction(ctx, txId)
			require.True(t, found, "should find transaction")
			require.Equal(t, moduleId, tx.RelatedModuleId, "should have correct module ID")
		}
	})

	t.Run("Amount and Coin Validation", func(t *testing.T) {
		// Test various coin amounts and denominations
		testAmounts := []sdk.Coin{
			sdk.NewCoin("utoken", math.NewInt(1)),        // Minimum amount
			sdk.NewCoin("utoken", math.NewInt(1000000)),  // Large amount
			sdk.NewCoin("stake", math.NewInt(500)),       // Different denomination
			sdk.NewCoin("ibc/ABC123", math.NewInt(1000)), // IBC token
		}

		for i, amount := range testAmounts {
			txId := keeper.RecordTransaction(
				ctx,
				types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
				types.ModuleID_MODULE_ID_LEASE,
				uint64(1200001+i),
				amount,
				uint64(1300001+i),
				uint64(1400001+i),
				sample.AccAddress(),
				sample.AccAddress(),
			)

			require.Greater(t, txId, uint64(0), "should create transaction with amount %s", amount.String())

			// Verify amount was stored correctly
			tx, found := keeper.GetTransaction(ctx, txId)
			require.True(t, found, "should find transaction")
			require.Equal(t, amount, tx.Amount, "should have correct amount")
		}
	})

	t.Run("Organization ID Edge Cases", func(t *testing.T) {
		// Test transactions with same from/to org (internal transfers)
		sameOrgId := uint64(1500001)

		txId := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_REFUND,
			types.ModuleID_MODULE_ID_LEASE,
			1600001,
			sdk.NewCoin("utoken", math.NewInt(1000)),
			sameOrgId, // from
			sameOrgId, // to (same org)
			sample.AccAddress(),
			sample.AccAddress(),
		)

		require.Greater(t, txId, uint64(0), "should create transaction with same from/to org")

		// Verify transaction appears only once in org's transaction list
		// (the indexing logic should avoid duplicates for same from/to org)
		allTransactionsResp4, err := keeper.ListPaymentFiltered(ctx, &types.QueryListPaymentFilteredRequest{
			OrgId: sameOrgId,
			Pagination: &query.PageRequest{Limit: 1000},
		})
		require.NoError(t, err, "should get org transactions")

		orgTxs := allTransactionsResp4.Transaction

		// Count how many times this transaction appears
		count := 0
		for _, tx := range orgTxs {
			if tx.Id == txId {
				count++
			}
		}
		require.Equal(t, 1, count, "transaction should appear only once for same from/to org")

		// Test zero org IDs
		txId2 := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			1600002,
			sdk.NewCoin("utoken", math.NewInt(1000)),
			0, // zero from org
			1700001,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		require.Greater(t, txId2, uint64(0), "should create transaction with zero from org")

		tx2, found := keeper.GetTransaction(ctx, txId2)
		require.True(t, found, "should find transaction with zero from org")
		require.Equal(t, uint64(0), tx2.FromOrgId, "should preserve zero from org ID")
	})

	t.Run("Address Validation", func(t *testing.T) {
		// Test various address formats
		addresses := []string{
			sample.AccAddress(), // Valid bech32
			sample.AccAddress(), // Another valid bech32
			"",                  // Empty address (should be allowed)
		}

		for i, addr := range addresses {
			txId := keeper.RecordTransaction(
				ctx,
				types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
				types.ModuleID_MODULE_ID_LEASE,
				uint64(1800001+i),
				sdk.NewCoin("utoken", math.NewInt(1000)),
				uint64(1900001+i),
				uint64(2000001+i),
				addr, // from address
				addr, // to address
			)

			require.Greater(t, txId, uint64(0), "should create transaction with address %s", addr)

			tx, found := keeper.GetTransaction(ctx, txId)
			require.True(t, found, "should find transaction")
			require.Equal(t, addr, tx.FromAddress, "should have correct from address")
			require.Equal(t, addr, tx.ToAddress, "should have correct to address")
		}
	})

	t.Run("Block Height and Timestamp Consistency", func(t *testing.T) {
		// Test that block height and timestamp are recorded correctly
		initialHeight := ctx.BlockHeight()
		initialTime := ctx.BlockTime()

		txId1 := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_LEASE_PAYMENT,
			types.ModuleID_MODULE_ID_LEASE,
			2100001,
			sdk.NewCoin("utoken", math.NewInt(1000)),
			2200001,
			2300001,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Advance block
		newHeight := initialHeight + 10
		newTime := initialTime.Add(time.Minute * 10)
		ctx = ctx.WithBlockHeight(newHeight).WithBlockTime(newTime)

		txId2 := keeper.RecordTransaction(
			ctx,
			types.TransactionType_TRANSACTION_TYPE_SAFETY_DEPOSIT_COLLECT,
			types.ModuleID_MODULE_ID_LEASE,
			2100002,
			sdk.NewCoin("utoken", math.NewInt(2000)),
			2200002,
			2300002,
			sample.AccAddress(),
			sample.AccAddress(),
		)

		// Verify first transaction has initial block data
		tx1, found := keeper.GetTransaction(ctx, txId1)
		require.True(t, found, "should find first transaction")
		require.NotNil(t, tx1.Timestamp, "should have timestamp")
		require.Equal(t, initialTime, *tx1.Timestamp, "should have initial timestamp")

		// Verify second transaction has updated block data
		tx2, found := keeper.GetTransaction(ctx, txId2)
		require.True(t, found, "should find second transaction")
		require.NotNil(t, tx2.Timestamp, "should have timestamp")
		require.Equal(t, newTime, *tx2.Timestamp, "should have new timestamp")

		// Verify timestamps are different
		require.True(t, tx2.Timestamp.After(*tx1.Timestamp), "second transaction should be after first")
	})
}
