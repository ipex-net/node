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

				{
					RpcMethod: "BidAll",
					Use:       "list-bid",
					Short:     "List all bid",
				},
				{
					RpcMethod:      "Bid",
					Use:            "show-bid [id]",
					Short:          "Shows a bid",
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
					RpcMethod: "CreateContract",
					Use:       "create-contract [owner-org] [price] [safety-deposit] [revocation-period-month] [subnet-id] [geo-change-allowed] [l2-blacklist-count-threshold] [accept-first-bidder] [allow-asn-drop-origin-transfer] [allow-unverified-bidders]",
					Short:     "Send a create-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "ownerOrg"},
						{ProtoField: "price"},
						{ProtoField: "safetyDeposit"},
						{ProtoField: "revocationPeriodMonth"},
						{ProtoField: "subnetId"},
						{ProtoField: "geoChangeAllowed"},
						{ProtoField: "l2BlacklistCountThreshold"},
						{ProtoField: "acceptFirstBidder"},
						{ProtoField: "allowASNDROPOriginTransfer"},
						{ProtoField: "allowUnverifiedBidders"},
					},
				},
				{
					RpcMethod: "UpdateContract",
					Use:       "update-contract [price] [id]",
					Short:     "Send a update-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "price"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod:      "DeleteContract",
					Use:            "delete-contract [id]",
					Short:          "Send a delete-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "CancelContract",
					Use:            "cancel-contract [id]",
					Short:          "Send a cancel-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "RevokeContract",
					Use:            "revoke-contract [id]",
					Short:          "Send a revoke-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "EnrollContract",
					Use:            "enroll-contract [contract-id] [lessee-org]",
					Short:          "Send a enroll-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contractId"}, {ProtoField: "lesseeOrg"}},
				},
				{
					RpcMethod:      "TerminateContract",
					Use:            "terminate-contract [id]",
					Short:          "Send a terminate-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "RenewContract",
					Use:            "renew-contract [contract-id]",
					Short:          "Send a renew-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contractId"}},
				},
				{
					RpcMethod:      "UndoCancelContract",
					Use:            "undo-cancel-contract [id]",
					Short:          "Send a undo-cancel-contract tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "RevokeBid",
					Use:            "revoke-bid [contract-id] [bidder-org-id]",
					Short:          "Send a revoke-bid tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contractId"}, {ProtoField: "bidderOrgId"}},
				},
				{
					RpcMethod:      "RejectBid",
					Use:            "reject-bid [contract-id] [bidder-org-id]",
					Short:          "Send a reject-bid tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contractId"}, {ProtoField: "bidderOrgId"}},
				},
				{
					RpcMethod:      "AcceptBid",
					Use:            "accept-bid [contract-id] [bidder-org-id]",
					Short:          "Send a accept-bid tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "contractId"}, {ProtoField: "bidderOrgId"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
