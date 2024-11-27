package lease

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/lease"
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
					RpcMethod:      "ShowContract",
					Use:            "show-contract [id]",
					Short:          "Query show-contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},

				{
					RpcMethod:      "ListContract",
					Use:            "list-contract",
					Short:          "Query list-contract",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
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
					RpcMethod:      "CreateContract",
					Use:            "create-contract [network-address] [prefix-length] [rir] [country-code] [geofeed-url] [price] [deposit] [notice-period-blocks]",
					Short:          "Send a create-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "networkAddress"}, {ProtoField: "prefixLength"}, {ProtoField: "rir"}, {ProtoField: "countryCode"}, {ProtoField: "geofeedUrl"}, {ProtoField: "price"}, {ProtoField: "deposit"}, {ProtoField: "noticePeriodBlocks"}},
				},
				{
					RpcMethod:      "UpdateContract",
					Use:            "update-contract [price] [deposit] [notice-period-blocks] [id]",
					Short:          "Send a update-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "price"}, {ProtoField: "deposit"}, {ProtoField: "noticePeriodBlocks"}, {ProtoField: "id"}},
				},
				{
					RpcMethod:      "DeleteContract",
					Use:            "delete-contract [id]",
					Short:          "Send a delete-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
