package abuse

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/abuse"
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
					RpcMethod: "ComplaintAll",
					Use:       "list-complaint",
					Short:     "List all complaint",
				},
				{
					RpcMethod:      "Complaint",
					Use:            "show-complaint [id]",
					Short:          "Shows a complaint",
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
				{
					RpcMethod:      "SubmitComplaint",
					Use:            "submit-complaint [lease-id] [subnet-id] [abuse-type] [incident-time] [source-ip] [lessor-ticket-ref] [evidence-hash] [summary]",
					Short:          "Send a submit-complaint tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "leaseId"}, {ProtoField: "subnetId"}, {ProtoField: "abuseType"}, {ProtoField: "incidentTime"}, {ProtoField: "sourceIp"}, {ProtoField: "lessorTicketRef"}, {ProtoField: "evidenceHash"}, {ProtoField: "summary"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
