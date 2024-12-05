package org

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"ipex/testutil/sample"
	orgsimulation "ipex/x/org/simulation"
	"ipex/x/org/types"
)

// avoid unused import issue
var (
	_ = orgsimulation.FindAccount
	_ = rand.Rand{}
	_ = sample.AccAddress
	_ = sdk.AccAddress{}
	_ = simulation.MsgEntryKind
)

const (
	opWeightMsgRegisterOrg = "op_weight_msg_register_org"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRegisterOrg int = 100

	opWeightMsgUpdateOrg = "op_weight_msg_update_org"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateOrg int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	orgGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&orgGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgRegisterOrg int
	simState.AppParams.GetOrGenerate(opWeightMsgRegisterOrg, &weightMsgRegisterOrg, nil,
		func(_ *rand.Rand) {
			weightMsgRegisterOrg = defaultWeightMsgRegisterOrg
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRegisterOrg,
		orgsimulation.SimulateMsgRegisterOrg(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateOrg int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdateOrg, &weightMsgUpdateOrg, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateOrg = defaultWeightMsgUpdateOrg
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateOrg,
		orgsimulation.SimulateMsgUpdateOrg(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgRegisterOrg,
			defaultWeightMsgRegisterOrg,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				orgsimulation.SimulateMsgRegisterOrg(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgUpdateOrg,
			defaultWeightMsgUpdateOrg,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				orgsimulation.SimulateMsgUpdateOrg(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
