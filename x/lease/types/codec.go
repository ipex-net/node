package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	// this line is used by starport scaffolding # 1
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRevokeContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgEnrollContract{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgTerminateContract{},
	)
	// this line is used by starport scaffolding # 3

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
