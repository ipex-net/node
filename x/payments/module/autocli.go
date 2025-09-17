package payments

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/payments"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "TransactionAll",
					Use:       "list-transaction",
					Short:     "List all transaction",
				},
				{
					RpcMethod:      "Transaction",
					Use:            "show-transaction [id]",
					Short:          "Shows a transaction by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "ListContractPayment",
					Use:            "list-contract-payment [id]",
					Short:          "Query list-contract-payment",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},

				{
					RpcMethod:      "ListPaymentFiltered",
					Use:            "list-payment-filtered [org-id] [transaction-type] [module-id] [from-timestamp] [to-timestamp] [min-amount] [max-amount] [entity-id] [from-org-id] [to-org-id]",
					Short:          "Query list-payment-filtered",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "orgId"}, {ProtoField: "transactionType"}, {ProtoField: "moduleId"}, {ProtoField: "fromTimestamp"}, {ProtoField: "toTimestamp"}, {ProtoField: "minAmount"}, {ProtoField: "maxAmount"}, {ProtoField: "entityId"}, {ProtoField: "fromOrgId"}, {ProtoField: "toOrgId"}},
				},

				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              modulev1.Msg_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
