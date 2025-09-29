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
				{
					RpcMethod:      "ListSubnet",
					Use:            "list-subnet",
					Short:          "Query list-subnet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},

				{
					RpcMethod:      "ShowSubnet",
					Use:            "show-subnet [id]",
					Short:          "Query show-subnet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "SubnetByOwnerOrg",
					Use:            "subnet-by-owner-org [owner-org-id]",
					Short:          "Query subnets by owner org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "owner_org_id"}},
				},
				{
					RpcMethod:      "SubnetByUserOrg",
					Use:            "subnet-by-user-org [user-org-id]",
					Short:          "Query subnets by user org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "user_org_id"}},
				},
				{
					RpcMethod:      "SubnetByNetworkAddress",
					Use:            "subnet-by-network-address [network-address]",
					Short:          "Query subnets by network address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "network_address"}},
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
					Use:            "create-subnet [owner-org-id] [network-address] [mask] [country]",
					Short:          "Send a create-subnet tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "ownerOrgId"}, {ProtoField: "networkAddress"}, {ProtoField: "mask"}, {ProtoField: "country"}},
				},
				{
					RpcMethod:      "UpdateSubnetUser",
					Use:            "update-subnet-user [id] [user-org-id]",
					Short:          "Send a update-subnet-user tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "userOrgId"}},
				},
				{
					RpcMethod:      "RemoveSubnetUser",
					Use:            "remove-subnet-user [id]",
					Short:          "Send a remove-subnet-user tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "UpdateSubnetOrigin",
					Use:            "update-subnet-origin [id] [as]",
					Short:          "Send a update-subnet-origin tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "as"}},
				},
				{
					RpcMethod:      "UpdateSubnetGeolocation",
					Use:            "update-subnet-geolocation [id] [country] [region] [city] [postal-code]",
					Short:          "Send a update-subnet-geolocation tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "country"}, {ProtoField: "region"}, {ProtoField: "city"}, {ProtoField: "postalCode"}},
				},
				{
					RpcMethod:      "DeleteSubnet",
					Use:            "delete-subnet [id]",
					Short:          "Send a delete-subnet tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
