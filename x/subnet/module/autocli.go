package subnet

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/subnet"
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
					RpcMethod: "SubnetAll",
					Use:       "list-subnet",
					Short:     "List all subnet",
				},
				{
					RpcMethod:      "Subnet",
					Use:            "show-subnet [id]",
					Short:          "Shows a subnet by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
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
					RpcMethod:      "CreateSubnet",
					Use:            "create-subnet [owner-org-id] [network-address] [mask] [geofeed-url] [rir]",
					Short:          "Send a create-subnet tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "ownerOrgId"}, {ProtoField: "networkAddress"}, {ProtoField: "mask"}, {ProtoField: "geofeedUrl"}, {ProtoField: "rir"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
