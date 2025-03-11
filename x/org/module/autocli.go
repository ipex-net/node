package org

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	modulev1 "ipex/api/ipex/org"
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
					RpcMethod: "OrgAll",
					Use:       "list-org",
					Short:     "List all org",
				},
				{
					RpcMethod:      "Org",
					Use:            "show-org [id]",
					Short:          "Shows a org by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "ShowOrg",
					Use:            "show-org [id]",
					Short:          "Query show-org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},

				{
					RpcMethod:      "ListOrg",
					Use:            "list-org",
					Short:          "Query list-org",
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
					RpcMethod:      "RegisterOrg",
					Use:            "register-org [name] [legal-name] [address] [registration-number] [country-code] [tax-id] [website] [abuse-email] [contact-email] [treasury]",
					Short:          "Send a register-org tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}, {ProtoField: "legalName"}, {ProtoField: "address"}, {ProtoField: "registrationNumber"}, {ProtoField: "countryCode"}, {ProtoField: "taxId"}, {ProtoField: "website"}, {ProtoField: "abuseEmail"}, {ProtoField: "contactEmail"}, {ProtoField: "treasury"}},
				},
				{
					RpcMethod:      "UpdateOrg",
					Use:            "update-org [id] [admin] [name] [legal-name] [address] [registration-number] [country-code] [tax-id] [website] [abuse-email] [contact-email] [treasury]",
					Short:          "Send a update-org tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "admin"}, {ProtoField: "name"}, {ProtoField: "legalName"}, {ProtoField: "address"}, {ProtoField: "registrationNumber"}, {ProtoField: "countryCode"}, {ProtoField: "taxId"}, {ProtoField: "website"}, {ProtoField: "abuseEmail"}, {ProtoField: "contactEmail"}, {ProtoField: "treasury"}},
				},
				{
					RpcMethod:      "DeleteOrg",
					Use:            "delete-org [id]",
					Short:          "Send a delete-org tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
