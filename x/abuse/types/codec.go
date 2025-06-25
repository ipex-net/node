package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	// this line is used by starport scaffolding # 1
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitComplaint{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRespondComplaint{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddFollowup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCloseComplaint{},
	)
	// this line is used by starport scaffolding # 3

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
