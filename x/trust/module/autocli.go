package trust

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/trust"
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
					RpcMethod: "SubnetReputationAll",
					Use:       "list-subnet-reputation",
					Short:     "List all subnet-reputation",
				},
				{
					RpcMethod:      "SubnetReputation",
					Use:            "show-subnet-reputation [id]",
					Short:          "Shows a subnet-reputation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod: "BlacklistLogAll",
					Use:       "list-blacklist-log",
					Short:     "List all blacklist-log",
				},
				{
					RpcMethod:      "BlacklistLog",
					Use:            "show-blacklist-log [id]",
					Short:          "Shows a blacklist-log by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "BlacklistLogBySubnet",
					Use:            "blacklist-log-by-subnet [subnet-id]",
					Short:          "Query blacklist-log-by-subnet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "subnetId"}},
				},

				{
					RpcMethod:      "BlacklistLogByLessor",
					Use:            "blacklist-log-by-lessor [lessor-org-id]",
					Short:          "Query blacklist-log-by-lessor",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "lessorOrgId"}},
				},

				{
					RpcMethod:      "BlacklistLogByLease",
					Use:            "blacklist-log-by-lease [lease-id]",
					Short:          "Query blacklist-log-by-lease",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "leaseId"}},
				},

				{
					RpcMethod:      "BlacklistLogByLessee",
					Use:            "blacklist-log-by-lessee [lessee-org-id]",
					Short:          "Query blacklist-log-by-lessee",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "lesseeOrgId"}},
				},

				{
					RpcMethod: "LeaseReputationAll",
					Use:       "list-lease-reputation",
					Short:     "List all lease-reputation",
				},
				{
					RpcMethod:      "LeaseReputation",
					Use:            "show-lease-reputation [id]",
					Short:          "Shows a lease-reputation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod: "LesseeReputationAll",
					Use:       "list-lessee-reputation",
					Short:     "List all lessee-reputation",
				},
				{
					RpcMethod:      "LesseeReputation",
					Use:            "show-lessee-reputation [id]",
					Short:          "Shows a lessee-reputation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod: "LessorReputationAll",
					Use:       "list-lessor-reputation",
					Short:     "List all lessor-reputation",
				},
				{
					RpcMethod:      "LessorReputation",
					Use:            "show-lessor-reputation [id]",
					Short:          "Shows a lessor-reputation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
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
